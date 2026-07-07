# Package Manager Template Automation

## Task

Make Homebrew and Scoop template updates repeatable from release checksums, so
future external tap and bucket work starts from generated files instead of
manual edits.

## Mode

LOCAL_ONLY. This was a focused distribution-governance slice.

## Analysis

- Current templates under `packaging/` were already pinned to `v0.2.5` and
  passed checksum validation.
- The remaining maintenance risk was manual drift: version, URLs, hashes,
  Scoop `extract_dir`, packaging README, and winget notes could be changed
  independently.
- Homebrew formula practice expects release URLs, SHA-256 values, and a test
  block. Scoop manifests support architecture-specific URLs and hashes plus
  `checkver`/`autoupdate`. DidaCLI already had both shapes; it needed a tested
  generator.

## Plan

1. Add a generator that reads a release `checksums.txt`.
2. Regenerate Homebrew, Scoop, packaging README, and winget notes.
3. Add mutation tests for generated versions, hashes, extract directories, and
   missing checksums.
4. Wire the generator test into `make release-check`.
5. Update release, distribution, changelog, roadmap, and archive docs.

## Progress

- [x] Added `scripts/update-packaging-templates.sh`.
- [x] Added `scripts/update-packaging-templates.test.sh`.
- [x] Regenerated current `v0.2.5` packaging files from the published release
  checksum asset.
- [x] Added the generator test to `make release-check`.
- [x] Updated release and distribution docs.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/update-packaging-templates.test.sh
bash scripts/validate-packaging.sh --version v0.2.5
bash scripts/validate-packaging.test.sh
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
```
