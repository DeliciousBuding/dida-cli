#!/usr/bin/env bash
set -euo pipefail

output_dir="dist/package-manager-repos"
run_homebrew_smoke=false
run_scoop_smoke=false
brew_command="brew"
scoop_command="scoop"

usage() {
  cat <<'EOF'
usage: bash scripts/package-manager-smoke-preflight.sh [options]

Options:
  --output <dir>            Export directory. Default: dist/package-manager-repos
  --run-homebrew-smoke      Run native Homebrew audit/install/test/uninstall
  --run-scoop-smoke         Run native Scoop install/uninstall smoke
  --brew-command <command>  Homebrew command. Default: brew
  --scoop-command <command> Scoop command. Default: scoop

By default this script exports and checks package-manager repository layouts,
then prints the native smoke commands. It does not install anything unless a
--run-* flag is passed.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      if [[ $# -lt 2 ]]; then
        echo "--output requires a directory" >&2
        exit 1
      fi
      output_dir="$2"
      shift 2
      ;;
    --run-homebrew-smoke)
      run_homebrew_smoke=true
      shift
      ;;
    --run-scoop-smoke)
      run_scoop_smoke=true
      shift
      ;;
    --brew-command)
      if [[ $# -lt 2 ]]; then
        echo "--brew-command requires a command" >&2
        exit 1
      fi
      brew_command="$2"
      shift 2
      ;;
    --scoop-command)
      if [[ $# -lt 2 ]]; then
        echo "--scoop-command requires a command" >&2
        exit 1
      fi
      scoop_command="$2"
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

bash scripts/export-package-manager-repos.sh --output "$output_dir" >/dev/null

homebrew_formula="$output_dir/homebrew-tap/Formula/dida.rb"
scoop_manifest="$output_dir/scoop-bucket/bucket/dida.json"
scoop_manifest_for_command="$scoop_manifest"
if command -v cygpath >/dev/null 2>&1; then
  case "$(uname -s 2>/dev/null || true)" in
    MINGW*|MSYS*|CYGWIN*)
      scoop_manifest_for_command="$(cygpath -w "$scoop_manifest")"
      ;;
  esac
fi

for required in \
  "$homebrew_formula" \
  "$output_dir/homebrew-tap/README.md" \
  "$output_dir/homebrew-tap/LICENSE" \
  "$scoop_manifest" \
  "$output_dir/scoop-bucket/README.md" \
  "$output_dir/scoop-bucket/LICENSE"
do
  if [[ ! -f "$required" ]]; then
    echo "missing exported package-manager file: $required" >&2
    exit 1
  fi
done

homebrew_version="$(sed -n 's/.*version "\([^"]*\)".*/\1/p' "$homebrew_formula" | head -n 1)"
scoop_version="$(node -e 'const fs = require("fs"); console.log(JSON.parse(fs.readFileSync(process.argv[1], "utf8")).version)' "$scoop_manifest")"
if [[ -z "$homebrew_version" || "$homebrew_version" != "$scoop_version" ]]; then
  echo "exported package-manager versions differ: homebrew=${homebrew_version:-missing} scoop=${scoop_version:-missing}" >&2
  exit 1
fi

case "$(grep -R -E 'C:\\|/Users/|/home/|TOKEN|SECRET|cookie' "$output_dir" || true)" in
  "")
    ;;
  *)
    echo "export contains local paths or secret-like words" >&2
    exit 1
    ;;
esac

require_command() {
  local command_name="$1"
  local label="$2"
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "$label command is required for native smoke: $command_name" >&2
    exit 1
  fi
}

run_binary_smoke() {
  require_command dida "installed dida"
  dida version >/dev/null
  dida doctor --json >/dev/null
}

echo "package-manager export preflight passed for v${homebrew_version}"
echo "Homebrew formula: $homebrew_formula"
echo "Scoop manifest: $scoop_manifest"
echo "This preflight does not create repositories or publish package-manager channels."

if [[ "$run_homebrew_smoke" != true && "$run_scoop_smoke" != true ]]; then
  cat <<EOF

No package-manager install smoke was run.

Run Homebrew smoke on a host with Homebrew:

  bash scripts/package-manager-smoke-preflight.sh --run-homebrew-smoke

Equivalent commands:

  brew audit --strict --online --formula "$homebrew_formula"
  brew install --formula "$homebrew_formula"
  brew test dida
  dida version
  dida doctor --json
  brew uninstall dida

Run Scoop smoke on a Windows host with Scoop:

  bash scripts/package-manager-smoke-preflight.sh --run-scoop-smoke

Equivalent commands:

  scoop install "$scoop_manifest_for_command"
  dida version
  dida doctor --json
  scoop uninstall dida
EOF
fi

if [[ "$run_homebrew_smoke" == true ]]; then
  require_command "$brew_command" "Homebrew"
  "$brew_command" audit --strict --online --formula "$homebrew_formula"
  "$brew_command" install --formula "$homebrew_formula"
  run_binary_smoke
  "$brew_command" test dida
  "$brew_command" uninstall dida
  echo "Homebrew smoke passed"
fi

if [[ "$run_scoop_smoke" == true ]]; then
  require_command "$scoop_command" "Scoop"
  "$scoop_command" install "$scoop_manifest_for_command"
  run_binary_smoke
  "$scoop_command" uninstall dida
  echo "Scoop smoke passed"
fi
