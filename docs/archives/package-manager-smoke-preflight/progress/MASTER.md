# Package Manager Smoke Preflight

## Task

Add a repeatable preflight for Homebrew and Scoop publication readiness without
creating external repositories or publishing package-manager channels.

## Mode

LOCAL_ONLY. This was a focused release-governance slice.

## Analysis

- The repository already exports Homebrew tap and Scoop bucket layouts.
- The next external step still needs native package-manager smoke on hosts with
  Homebrew and Scoop installed.
- The smoke commands were documented in prose, but not packaged as a reusable
  repo-local check.

## Plan

1. Add a preflight script that exports and checks the package-manager layouts.
2. Keep install smoke opt-in through explicit `--run-*` flags.
3. Add tests that cover default preflight, fake native smoke commands, and
   missing host tooling failures.
4. Wire the tests into `make release-check`.
5. Update release, distribution, roadmap, changelog, and archive docs.

## Progress

- [x] Added `scripts/package-manager-smoke-preflight.sh`.
- [x] Added `scripts/package-manager-smoke-preflight.test.sh`.
- [x] Wired the test into `make release-check`.
- [x] Added repository governance checks for the preflight docs and test.
- [x] Updated maintainer docs and roadmap status.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/package-manager-smoke-preflight.test.sh
bash scripts/package-manager-smoke-preflight.sh --output dist/package-manager-preflight-check
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
git diff --check
```
