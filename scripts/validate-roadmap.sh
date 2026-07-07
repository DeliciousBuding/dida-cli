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
      version="$2"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$version" ]]; then
  version="v$(node -e 'console.log(require("./npm/package.json").version)')"
fi

if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "roadmap version must use semver tag format vX.Y.Z; got $version" >&2
  exit 1
fi

if [[ ! -f ROADMAP.md ]]; then
  echo "ROADMAP.md is missing" >&2
  exit 1
fi

require_text() {
  local pattern="$1"
  local label="$2"
  if ! grep -Eq -- "$pattern" ROADMAP.md; then
    echo "ROADMAP.md is missing $label" >&2
    exit 1
  fi
}

next_minor="$(node -e '
const version = process.argv[1].replace(/^v/, "").split(".").map(Number);
console.log(`v${version[0]}.${version[1] + 1}.0`);
' "$version")"

escaped_version="${version//./\\.}"
escaped_next_minor="${next_minor//./\\.}"

require_text "^## Current Baseline \\(as of ${escaped_version}\\)$" "current baseline heading for $version"
require_text "^Latest release: \`${escaped_version}\` \\([0-9]{4}-[0-9]{2}-[0-9]{2}\\)\\.$" "latest release line for $version"
require_text "For ${escaped_next_minor} \\(next milestone\\):" "next milestone section for $next_minor"
require_text "CLI coverage floor \\| Done \\|" "completed CLI coverage roadmap row"
require_text "Status: implemented and published through \`${escaped_version}\`" "release workflow status for $version"
require_text "Status: published as \`@delicious233/dida-cli@${version#v}\`" "npm status for $version"
require_text "dida-package-manager-repos-vX\\.Y\\.Z" "package-manager release artifact handoff"

if grep -Eq 'Status: implemented[^[:cntrl:]]*through `v0\.2\.1`|against[[:space:]]+`v0\.2\.1`|against[[:space:]]+`v0\.1\.16`|published through `v0\.2\.1`' ROADMAP.md; then
  echo "ROADMAP.md still contains stale distribution release baseline text" >&2
  exit 1
fi

next_tasks="$(
  awk '
    /^## Current Best Next Tasks$/ { in_section = 1; next }
    /^## Done Means Done$/ { in_section = 0 }
    in_section { print }
  ' ROADMAP.md
)"

if grep -Eq 'For v[0-9]+\.[0-9]+\.[0-9]+ release \(immediate\):' <<<"$next_tasks"; then
  echo "ROADMAP.md Current Best Next Tasks still contains a stale immediate release block" >&2
  exit 1
fi

echo "roadmap metadata valid for $version"
