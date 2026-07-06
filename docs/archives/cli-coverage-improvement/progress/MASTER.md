# CLI Coverage Improvement

## Task

Raise the `internal/cli` test baseline without requiring real Dida365 auth,
network writes, OAuth tokens, or external services.

## Mode

LOCAL_ONLY. This was a focused test and tooling slice, archived immediately
after implementation.

## Analysis

- `go test ./internal/cli -coverprofile=coverage-cli` measured 43.9% before
  this slice.
- Low-risk gaps were local command dispatch, help output, task/project dry-run
  previews, sync-backed read wrappers, and OpenAPI task dry-run previews.
- The 60% roadmap target needs more command families, but this slice avoids
  live credentials and keeps coverage work repeatable.

## Plan

1. Remove the stale generated `coverage-cli` profile before editing.
2. Add CLI tests for local help and task/project dry-run paths.
3. Add httptest-backed read tests for sync, task, project, filter, settings,
   trash, due-counts, and raw GET.
4. Add OpenAPI task dry-run tests that do not require a saved token.
5. Add a Makefile helper for repeatable CLI coverage.
6. Update roadmap, changelog, and archive index.

## Progress

- [x] Added `internal/cli/coverage_test.go`.
- [x] Added `make coverage-cli`.
- [x] Updated `ROADMAP.md`, `CHANGELOG.md`, `CLAUDE.md`, and archive index.
- [x] Re-measured `internal/cli` at 50.8%.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | M |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 1: Go wrote the coverage profile as `coverage-cli`, so the helper uses that exact filename |

## Verification

Run before commit:

```bash
make coverage-cli
go test -count=1 ./...
go vet ./...
make staticcheck
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
make release-check VERSION=v0.2.5
```
