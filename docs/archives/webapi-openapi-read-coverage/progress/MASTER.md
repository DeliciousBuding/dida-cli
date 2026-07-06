# WebAPI and OpenAPI Read Coverage

## Task

Raise the `internal/cli` coverage baseline past the 60% roadmap target without
using real Dida365 credentials, live network calls, or write operations.

## Mode

LOCAL_ONLY. This was a focused follow-up to the CLI coverage slice, archived
immediately after implementation.

## Analysis

- `make coverage-cli` measured 50.8% before this follow-up.
- The lowest-risk gaps were read-only Web API command wrappers that can share
  the existing sync-backed test server.
- OpenAPI read commands could use a saved fake OAuth token and a local
  `DIDA_OPENAPI_BASE_URL`, keeping channel isolation intact.

## Plan

1. Extend the existing sync-backed test server for folders, tags, quadrants,
   agent context, closed items, and attachment downloads.
2. Add an OpenAPI read-command test server with Authorization header checks.
3. Re-measure `internal/cli` coverage and update roadmap/changelog/archive
   state.
4. Run the full local release verification gates before commit.

## Progress

- [x] Covered more Web API read wrappers and sync-derived views.
- [x] Covered OpenAPI project, task, focus, and habit read commands.
- [x] Re-measured `internal/cli` at 61.3%.
- [x] Updated `ROADMAP.md`, `CHANGELOG.md`, and archive index.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

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
