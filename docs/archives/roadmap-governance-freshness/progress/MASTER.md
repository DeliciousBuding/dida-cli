# Roadmap Governance Freshness

## Task

Keep roadmap state tied to release metadata so future releases do not leave
`ROADMAP.md` pointing at stale baselines or completed next tasks.

## Mode

LOCAL_ONLY. This was a focused repository-governance slice, archived after
implementation.

## Analysis

- `ROADMAP.md` still described the current baseline as `v0.2.1`, while npm and
  repository metadata are at `v0.2.5`.
- Existing release gates validate changelog, npm metadata, workflows,
  packaging, and repository governance files, but not roadmap freshness.
- The smallest durable fix is a Bash validator that compares roadmap baseline
  text with the current package/release version and the computed next minor
  milestone.

## Plan

1. Add `scripts/validate-roadmap.sh`.
2. Add mutation tests in `scripts/validate-roadmap.test.sh`.
3. Wire the validator into `make release-check`.
4. Update `ROADMAP.md`, `CHANGELOG.md`, and archive index.

## Progress

- [x] Added roadmap freshness validation.
- [x] Added validator tests for stale release, stale heading, stale immediate
  release block, missing next milestone, and invalid version format.
- [x] Updated `ROADMAP.md` to `v0.2.5` current state and `v0.3.0` next
  milestone tasks.
- [x] Updated release-check coverage and docs.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/validate-roadmap.sh --version v0.2.5
bash scripts/validate-roadmap.test.sh
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
```
