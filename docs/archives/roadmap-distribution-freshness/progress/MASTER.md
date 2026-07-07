# Roadmap Distribution Freshness

## Task

Keep `ROADMAP.md` distribution status aligned with the current release, npm
package, and package-manager export path.

## Mode

LOCAL_ONLY. This was a focused roadmap-governance slice.

## Analysis

- `ROADMAP.md` already had a `v0.2.5` current baseline.
- The detailed distribution workstream still referenced older release evidence
  in F1, F2, and F3.
- Future agents use `ROADMAP.md` to choose the next slice, so stale
  distribution status can cause repeated rediscovery or outdated release
  assumptions.

## Plan

1. Update F1/F2/F3 distribution status to the current release line.
2. Record the npm `@delicious233/dida-cli@0.2.5` package state.
3. Keep package-manager export artifacts visible in roadmap status.
4. Extend `scripts/validate-roadmap.sh` and mutation tests to reject stale
   distribution baselines.
5. Update changelog and archive index.

## Progress

- [x] Updated `ROADMAP.md` release, installer, and npm distribution status.
- [x] Extended `scripts/validate-roadmap.sh`.
- [x] Added mutation coverage for stale distribution status.
- [x] Updated changelog and archive index.

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
git diff --check
```
