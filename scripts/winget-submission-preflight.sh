#!/usr/bin/env bash
set -euo pipefail

repo="DeliciousBuding/dida-cli"
package_id="DeliciousBuding.DidaCLI"
version_override=""
checksums_file=""
manifest_dir=""
winget_command="winget"
wingetcreate_command="wingetcreate"
require_tools=false
run_validate=false

usage() {
  cat <<'EOF'
usage: bash scripts/winget-submission-preflight.sh [options]

Options:
  --version <vX.Y.Z>          Release version. Default: packaging version
  --repo <owner/name>         GitHub release repository. Default: DeliciousBuding/dida-cli
  --package-id <id>           winget package id. Default: DeliciousBuding.DidaCLI
  --checksums-file <path>     Optional local checksums.txt for full hash checks
  --manifest-dir <dir>        Existing winget manifest directory to validate
  --run-validate              Run `winget validate --manifest <dir>`
  --require-tools             Require winget and wingetcreate to be available
  --winget-command <command>  winget command. Default: winget
  --wingetcreate-command <command>
                              wingetcreate command. Default: wingetcreate

The default mode checks local packaging notes and prints the Windows Package
Manager submission commands. It does not generate manifests or submit packages.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      if [[ $# -lt 2 ]]; then
        echo "--version requires a value" >&2
        exit 1
      fi
      version_override="${2#v}"
      shift 2
      ;;
    --repo)
      if [[ $# -lt 2 ]]; then
        echo "--repo requires owner/name" >&2
        exit 1
      fi
      repo="$2"
      shift 2
      ;;
    --package-id)
      if [[ $# -lt 2 ]]; then
        echo "--package-id requires a value" >&2
        exit 1
      fi
      package_id="$2"
      shift 2
      ;;
    --checksums-file)
      if [[ $# -lt 2 ]]; then
        echo "--checksums-file requires a path" >&2
        exit 1
      fi
      checksums_file="$2"
      shift 2
      ;;
    --manifest-dir)
      if [[ $# -lt 2 ]]; then
        echo "--manifest-dir requires a directory" >&2
        exit 1
      fi
      manifest_dir="$2"
      shift 2
      ;;
    --run-validate)
      run_validate=true
      shift
      ;;
    --require-tools)
      require_tools=true
      shift
      ;;
    --winget-command)
      if [[ $# -lt 2 ]]; then
        echo "--winget-command requires a command" >&2
        exit 1
      fi
      winget_command="$2"
      shift 2
      ;;
    --wingetcreate-command)
      if [[ $# -lt 2 ]]; then
        echo "--wingetcreate-command requires a command" >&2
        exit 1
      fi
      wingetcreate_command="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "repo must use owner/name format; got $repo" >&2
  exit 1
fi
if [[ ! "$package_id" =~ ^[A-Za-z0-9][A-Za-z0-9.-]*\.[A-Za-z0-9][A-Za-z0-9.-]*$ ]]; then
  echo "winget package id must use publisher.package format; got $package_id" >&2
  exit 1
fi

scoop_version="$(node -e 'console.log(require("./packaging/scoop/dida.json").version)')"
version="${version_override:-$scoop_version}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "version must be X.Y.Z or vX.Y.Z; got $version" >&2
  exit 1
fi

tag="v$version"
base_url="https://github.com/${repo}/releases/download/${tag}"
asset_windows_amd64="dida_${tag}_windows_amd64.zip"
asset_windows_arm64="dida_${tag}_windows_arm64.zip"
url_windows_amd64="${base_url}/${asset_windows_amd64}"
url_windows_arm64="${base_url}/${asset_windows_arm64}"

validate_args=(--version "$tag" --metadata-only)
if [[ -n "$checksums_file" ]]; then
  validate_args=(--version "$tag" --checksums-file "$checksums_file")
fi
bash scripts/validate-packaging.sh "${validate_args[@]}" >/dev/null

require_text() {
  local path="$1"
  local pattern="$2"
  local label="$3"
  if ! grep -Fq -- "$pattern" "$path"; then
    echo "$path is missing $label" >&2
    exit 1
  fi
}

require_text packaging/winget/README.md "$package_id" "winget package id"
require_text packaging/winget/README.md "$url_windows_amd64" "current Windows amd64 release URL"
require_text packaging/winget/README.md "wingetcreate new" "wingetcreate command"
require_text packaging/winget/README.md "winget validate --manifest" "winget validate command"

if [[ "$require_tools" == true || "$run_validate" == true ]]; then
  if ! command -v "$winget_command" >/dev/null 2>&1; then
    echo "winget command is required: $winget_command" >&2
    exit 1
  fi
fi
if [[ "$require_tools" == true ]]; then
  if ! command -v "$wingetcreate_command" >/dev/null 2>&1; then
    echo "wingetcreate command is required: $wingetcreate_command" >&2
    exit 1
  fi
fi

if [[ "$run_validate" == true ]]; then
  if [[ -z "$manifest_dir" ]]; then
    echo "--run-validate requires --manifest-dir" >&2
    exit 1
  fi
  if [[ ! -d "$manifest_dir" ]]; then
    echo "manifest directory is missing: $manifest_dir" >&2
    exit 1
  fi
fi

echo "winget submission preflight passed for ${package_id} ${tag}"
echo "Windows amd64 archive: $url_windows_amd64"
echo "Windows arm64 archive: $url_windows_arm64"
echo "This preflight does not generate manifests or submit packages."

cat <<EOF

Manual submission flow on a Windows packaging host:

  wingetcreate new "$url_windows_amd64"
  winget validate --manifest <manifest-directory>

Keep generated manifests out of this repository until the package identifier
and release cadence are final.
EOF

if [[ "$run_validate" == true ]]; then
  "$winget_command" validate --manifest "$manifest_dir"
  echo "winget manifest validation passed"
fi
