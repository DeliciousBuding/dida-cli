# Doctor Upgrade Check

## Task

Add an explicit upgrade check to `dida doctor` without changing the default
local-only behavior.

## Mode

LOCAL_ONLY. This was a narrow product slice, archived immediately after
implementation instead of leaving an active root tracker.

## Analysis

- `dida upgrade --check` already owned GitHub Releases lookup, semver compare,
  and latest release metadata.
- `dida doctor` was local-only unless `--verify` was passed.
- The clean integration point was an explicit `--check-upgrade` flag so normal
  doctor runs do not add a surprise network call.

## Plan

1. Reuse upgrade metadata lookup instead of duplicating GitHub release logic.
2. Add `upgrade_check` to doctor output, with `not_run`, `current`,
   `available`, and `failed` states.
3. Keep upgrade lookup failures advisory; Web API `--verify` remains the only
   doctor path that can fail on auth verification.
4. Update schema, README files, command reference, roadmap, and changelog.
5. Verify with focused and full test gates.

## Progress

- [x] Reused upgrade release metadata lookup from `dida upgrade`.
- [x] Added `dida doctor --check-upgrade`.
- [x] Added tests for default no-upgrade-check behavior, update available
  output, and advisory network failure.
- [x] Updated schema and user docs.
- [x] Archived this spec-driven slice.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
go test ./internal/cli -run "TestDoctor|TestUpgrade"
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
make release-check VERSION=v0.2.5
```
