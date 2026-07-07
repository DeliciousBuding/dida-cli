# Release Strategy GoReleaser Decision

## Task

Resolve the roadmap item asking whether to keep the current release workflow or
migrate to GoReleaser.

## Mode

LOCAL_ONLY. This was a focused release-governance decision slice.

## Analysis

- The current release workflow already validates tags, changelog sections, npm
  package metadata, tests, vet, Staticcheck, vulnerability scanning, private
  state, pinned actions, package-manager templates, and archive layout.
- The workflow also handles project-specific behavior: npm Trusted Publishing
  fallback to `NPM_TOKEN`, npm install smoke tests, GitHub artifact
  attestations, and release-note fallback rules.
- GoReleaser can generate archives and package-manager updates, but replacing
  the current workflow would require parity for npm provenance, attestations,
  changelog behavior, and cross-repository Homebrew/Scoop publishing tokens.

## Decision

Keep the current hand-written release workflow through `v0.3.x`. Re-evaluate
GoReleaser only after archive, checksum, npm provenance, attestation, and
package-manager publishing parity are proven.

## Plan

1. Add a GoReleaser decision record under `docs/research/`.
2. Add a release-strategy validator and mutation tests.
3. Wire the validator into `make release-check`.
4. Update `RELEASE.md`, `ROADMAP.md`, `CHANGELOG.md`, and archive index.

## Progress

- [x] Added `docs/research/release-strategy-goreleaser.md`.
- [x] Added `scripts/validate-release-strategy.sh`.
- [x] Added `scripts/validate-release-strategy.test.sh`.
- [x] Updated roadmap status from undecided to deferred through `v0.3.x`.
- [x] Added release-check coverage for the decision.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/validate-release-strategy.sh
bash scripts/validate-release-strategy.test.sh
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
```
