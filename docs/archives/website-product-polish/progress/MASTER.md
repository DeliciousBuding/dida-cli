# Website Product Polish

## Task

Align the GitHub Pages homepage with the current `v0.2.5` package and command
surface, then add a release gate that catches stale website copy.

## Mode

LOCAL_ONLY. This was a focused documentation and governance slice.

## Analysis

- The npm registry now reports `@delicious233/dida-cli@0.2.5` with a README,
  but an older npm page view still showed `0.2.4` and no README.
- `docs/index.html` still highlighted older command examples, including
  `dida +today --json`, and referenced parent `../assets/` paths that can break
  when GitHub Pages is served from `docs/`.
- Existing governance checks covered README, npm README, workflows, release
  metadata, changelog, roadmap, and packaging, but not website copy.

## Plan

1. Rewrite `docs/index.html` around the current install, auth, verify, schema,
   latest-task, completion, README, npm, and security paths.
2. Add a website validator and mutation tests.
3. Wire the validator into `make release-check`.
4. Update changelog, roadmap, and archive index.

## Progress

- [x] Replaced the homepage with a compact command-oriented page for `v0.2.5`.
- [x] Added `scripts/validate-website.sh`.
- [x] Added `scripts/validate-website.test.sh`.
- [x] Added website validation to `make release-check`.
- [x] Updated `CHANGELOG.md`, `ROADMAP.md`, and archive index.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/validate-website.sh
bash scripts/validate-website.test.sh
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
```
