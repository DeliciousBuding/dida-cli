# winget Submission Preflight

## Task

Add a repeatable winget submission preflight without generating or submitting a
manifest from this repository.

## Mode

LOCAL_ONLY. This was a focused distribution-governance slice.

## Analysis

- The roadmap listed winget as deferred and only had submission notes.
- `packaging/winget/README.md` named `wingetcreate new`, but did not encode the
  local validation handoff as a reusable check.
- Manifest generation should stay manual until the package id and release
  cadence are final.

## Plan

1. Add a preflight script that checks release URLs, package id, packaging
   metadata, and the submission boundary.
2. Keep generated manifests out of the repository.
3. Allow optional validation of an existing manifest directory with
   `winget validate --manifest`.
4. Add mutation tests and wire the test into `make release-check`.
5. Update release, distribution, roadmap, research audit, and archive docs.

## Progress

- [x] Added `scripts/winget-submission-preflight.sh`.
- [x] Added `scripts/winget-submission-preflight.test.sh`.
- [x] Wired the test into `make release-check`.
- [x] Updated generated winget notes and packaging-template tests.
- [x] Updated roadmap, release docs, distribution docs, and research audit.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/winget-submission-preflight.test.sh
bash scripts/winget-submission-preflight.sh --checksums-file <checksums.txt>
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
git diff --check
```
