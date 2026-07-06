#!/usr/bin/env bash
set -euo pipefail

repo="DeliciousBuding/dida-cli"
changelog="CHANGELOG.md"
required_tag=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      if [[ $# -lt 2 ]]; then
        echo "--tag requires a value" >&2
        exit 1
      fi
      required_tag="$2"
      shift 2
      ;;
    --changelog)
      if [[ $# -lt 2 ]]; then
        echo "--changelog requires a path" >&2
        exit 1
      fi
      changelog="$2"
      shift 2
      ;;
    --repo)
      if [[ $# -lt 2 ]]; then
        echo "--repo requires a value" >&2
        exit 1
      fi
      repo="$2"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "$changelog" ]]; then
  echo "changelog not found: $changelog" >&2
  exit 1
fi

normalized="$(mktemp)"
trap 'rm -f "$normalized"' EXIT
tr -d '\r' < "$changelog" > "$normalized"
changelog="$normalized"

if ! grep -q '^## \[Unreleased\]' "$changelog"; then
  echo "CHANGELOG.md must contain a ## [Unreleased] section" >&2
  exit 1
fi

mapfile -t versions < <(sed -n 's/^## \[\(v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)\] - \([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]\)$/\1/p' "$changelog")
if [[ "${#versions[@]}" -eq 0 ]]; then
  echo "CHANGELOG.md must contain at least one dated vX.Y.Z release section" >&2
  exit 1
fi

if [[ -n "$required_tag" ]]; then
  if [[ ! "$required_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "required changelog tag must use vX.Y.Z; got $required_tag" >&2
    exit 1
  fi
  if ! printf '%s\n' "${versions[@]}" | grep -qx "$required_tag"; then
    echo "CHANGELOG.md must contain ## [$required_tag] before release" >&2
    exit 1
  fi
fi

latest="${versions[0]}"
unreleased_url="https://github.com/${repo}/compare/${latest}...HEAD"
if ! grep -Fqx "[Unreleased]: ${unreleased_url}" "$changelog"; then
  echo "CHANGELOG.md must contain [Unreleased]: ${unreleased_url}" >&2
  exit 1
fi

for i in "${!versions[@]}"; do
  version="${versions[$i]}"
  if (( i + 1 < ${#versions[@]} )); then
    previous="${versions[$((i + 1))]}"
    url="https://github.com/${repo}/compare/${previous}...${version}"
  else
    url="https://github.com/${repo}/releases/tag/${version}"
  fi
  if ! grep -Fqx "[${version}]: ${url}" "$changelog"; then
    echo "CHANGELOG.md must contain [${version}]: ${url}" >&2
    exit 1
  fi
done

echo "changelog structure valid"
