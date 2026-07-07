# Package Manager Repo Export

## Task

Prepare publishable Homebrew tap and Scoop bucket repository layouts from the
validated packaging templates, without creating or pushing external
repositories.

## Mode

LOCAL_ONLY. This was a focused distribution-governance slice.

## Analysis

- `packaging/homebrew/dida.rb` and `packaging/scoop/dida.json` are already
  generated from release checksums and validated in `make release-check`.
- External package-manager publishing still needs separate repository roots:
  Homebrew expects the formula under `Formula/`; Scoop buckets expect manifests
  under `bucket/`.
- Publishing those repositories is an account and native-smoke step, so this
  slice keeps export local and repeatable.

## Plan

1. Add a script that exports Homebrew tap and Scoop bucket repo roots.
2. Add tests for directory layout, copied templates, install commands, and
   argument validation.
3. Wire the export test into `make release-check`.
4. Update packaging, distribution, roadmap, changelog, and archive docs.

## Progress

- [x] Added `scripts/export-package-manager-repos.sh`.
- [x] Added `scripts/export-package-manager-repos.test.sh`.
- [x] Added the export test to `make release-check`.
- [x] Documented the separation between template generation, local export, and
  external repository publishing.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/export-package-manager-repos.test.sh
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
git diff --check
```
