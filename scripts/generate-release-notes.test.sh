#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
generator="$repo_root/scripts/generate-release-notes.sh"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

cat >"$work/CHANGELOG.md" <<'EOF'
# Changelog

## [Unreleased]

## [v1.2.3] - 2026-07-06

### Added
- New release feature.

### Fixed
- Important release fix.

## [v1.2.2] - 2026-07-01

### Fixed
- Previous fix.
EOF

bash "$generator" --tag v1.2.3 --repo owner/repo --changelog "$work/CHANGELOG.md" --output "$work/notes.md"

grep -q "## What's Changed" "$work/notes.md"
grep -q "New release feature" "$work/notes.md"
grep -q "Important release fix" "$work/notes.md"
grep -q "dida_v1.2.3_linux_amd64.tar.gz" "$work/notes.md"
grep -q "checksums.txt" "$work/notes.md"

if bash "$generator" --tag v9.9.9 --repo owner/repo --changelog "$work/CHANGELOG.md" --output "$work/missing.md" 2>"$work/missing.err"; then
  echo "missing changelog section should fail without fallback" >&2
  exit 1
fi
grep -q "CHANGELOG.md must contain" "$work/missing.err"

git -C "$repo_root" log --oneline -1 >/dev/null
bash "$generator" --tag v9.9.9 --repo owner/repo --changelog "$work/CHANGELOG.md" --output "$work/fallback.md" --allow-changelog-fallback
grep -q "Changes since" "$work/fallback.md"

echo "generate-release-notes tests passed"
