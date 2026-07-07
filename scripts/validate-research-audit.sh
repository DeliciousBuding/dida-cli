#!/usr/bin/env bash
set -euo pipefail

version=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      if [[ $# -lt 2 ]]; then
        echo "--version requires a value" >&2
        exit 1
      fi
      version="${2#v}"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$version" ]]; then
  version="$(node -e 'console.log(require("./npm/package.json").version)')"
fi

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "version must be X.Y.Z or vX.Y.Z; got $version" >&2
  exit 1
fi

tag="v$version"
escaped_version="${version//./\\.}"
escaped_tag="${tag//./\\.}"

files=(
  docs/research/prompt-to-artifact-checklist.md
  docs/research/roadmap-completion-audit.md
)

for file in "${files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "research audit file is missing: $file" >&2
    exit 1
  fi
done

require_text() {
  local file="$1"
  local pattern="$2"
  local label="$3"
  if ! grep -Eq -- "$pattern" "$file"; then
    echo "$file is missing $label" >&2
    exit 1
  fi
}

reject_text() {
  local file="$1"
  local pattern="$2"
  local label="$3"
  if grep -Eq -- "$pattern" "$file"; then
    echo "$file contains stale $label" >&2
    exit 1
  fi
}

for file in "${files[@]}"; do
  reject_text "$file" 'v0\.2\.1|0\.2\.1|v0\.1\.16|0\.1\.16' "pre-v0.2.5 release baseline"
  require_text "$file" "${escaped_tag}" "current release tag ${tag}"
  require_text "$file" "@delicious233/dida-cli@${escaped_version}" "current npm package ${version}"
  require_text "$file" 'dida-package-manager-repos-vX\.Y\.Z' "package-manager release artifact handoff"
done

require_text docs/research/prompt-to-artifact-checklist.md 'Package-manager export artifact' "distribution checklist package-manager artifact row"
require_text docs/research/roadmap-completion-audit.md 'Release workflow exports package-manager repo layouts' "roadmap audit package-manager artifact evidence"

echo "research audit freshness valid for ${tag}"
