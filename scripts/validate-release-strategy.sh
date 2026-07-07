#!/usr/bin/env bash
set -euo pipefail

decision_file="docs/research/release-strategy-goreleaser.md"

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    echo "required release strategy file is missing: $path" >&2
    exit 1
  fi
}

require_text() {
  local path="$1"
  local pattern="$2"
  local label="$3"
  if ! grep -Eq -- "$pattern" "$path"; then
    echo "$path is missing $label" >&2
    exit 1
  fi
}

reject_text() {
  local path="$1"
  local pattern="$2"
  local label="$3"
  if grep -Eq -- "$pattern" "$path"; then
    echo "$path contains stale release strategy text: $label" >&2
    exit 1
  fi
}

require_file "$decision_file"
require_file ROADMAP.md
require_file RELEASE.md

require_text "$decision_file" '^## Decision$' "decision heading"
require_text "$decision_file" 'Keep the current hand-written release workflow for `v0\.3\.x`' "current workflow decision"
require_text "$decision_file" 'Do not migrate to[[:space:]]+GoReleaser in the next milestone\.' "no-migration decision"
require_text "$decision_file" '^## Revisit Conditions$' "revisit conditions heading"
require_text "$decision_file" 'homebrew-tap' "Homebrew tap revisit condition"
require_text "$decision_file" 'scoop-bucket' "Scoop bucket revisit condition"
require_text "$decision_file" 'GoReleaser dry run produces the same archive names' "dry-run parity condition"
require_text "$decision_file" 'npm publishing keeps the same README' "npm parity condition"
require_text "$decision_file" 'GitHub artifact attestation' "attestation condition"
require_text "$decision_file" 'cross-repository publishing token is scoped' "cross-repo token condition"

require_text ROADMAP.md 'goreleaser migration \| Deferred \| Keep the current release workflow through `v0\.3\.x`' "deferred GoReleaser roadmap row"
require_text ROADMAP.md 'Re-evaluate GoReleaser only after archive, checksum, npm provenance, attestation, and package-manager publishing parity are proven' "GoReleaser re-evaluation task"
require_text RELEASE.md 'release-strategy-goreleaser\.md' "release strategy reference"

reject_text ROADMAP.md 'goreleaser migration is undecided' "undecided GoReleaser baseline"
reject_text ROADMAP.md 'Decide whether to keep the current release workflow or migrate to goreleaser' "stale next task"

echo "release strategy decision valid"
