# Staticcheck Quality Gate

## Task

Add Staticcheck to the repository's repeatable quality gates without adding Go
module dependencies or unpinned GitHub Actions.

## Mode

LOCAL_ONLY. This was a focused governance slice, archived immediately after
implementation.

## Analysis

- The roadmap listed `staticcheck in CI` as a remaining v0.3.0 quality item.
- Existing tooling already used `go run <tool>@<version>` for `govulncheck`
  and `actionlint`.
- The repository requires external GitHub Actions to be SHA-pinned, so a direct
  `go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...` gate fit better
  than adding a new action.
- The first local Staticcheck run found unused helpers and one token parser
  shape where the loop condition variable never changed.

## Plan

1. Run Staticcheck locally and fix existing findings.
2. Add a `make staticcheck` target with a pinned default version.
3. Add Staticcheck to CI and release validation.
4. Include Staticcheck in `make release-check`.
5. Update agent/contributor governance docs and the governance validator.
6. Verify with Staticcheck, actionlint, repo governance tests, full Go tests,
   vulnerability scanning, private-state scanning, and release-check.

## Progress

- [x] Fixed Staticcheck findings in CLI auth parsing and unused helpers.
- [x] Added `make staticcheck`.
- [x] Added Staticcheck to CI and release workflows.
- [x] Added Staticcheck to `make release-check`.
- [x] Updated `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, PR template,
  `CHANGELOG.md`, and `ROADMAP.md`.
- [x] Extended `scripts/validate-repo-governance.sh` and its tests.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 1: existing Staticcheck findings had to be fixed first |

## Verification

Run before commit:

```bash
make staticcheck
bash scripts/validate-repo-governance.sh
bash scripts/validate-repo-governance.test.sh
make actionlint
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
make release-check VERSION=v0.2.5
```
