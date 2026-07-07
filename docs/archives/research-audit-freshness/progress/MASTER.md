# Research Audit Freshness

## Task

Keep the objective and distribution audit docs aligned with the current release
instead of leaving them on older release evidence.

## Mode

LOCAL_ONLY. This was a focused documentation-governance slice.

## Analysis

- `ROADMAP.md`, release docs, npm metadata, and packaging docs now point to
  `v0.2.5`.
- `docs/research/prompt-to-artifact-checklist.md` and
  `docs/research/roadmap-completion-audit.md` still described distribution
  evidence from older releases.
- Those files are used as completion audits, so stale release evidence can
  cause future agents to make the wrong next move.

## Plan

1. Update the research audit docs to the `v0.2.5` release and npm package.
2. Record the package-manager export script and release artifact handoff.
3. Add a freshness validator that rejects stale release baselines.
4. Add mutation tests and wire the validator into `make release-check`.
5. Update changelog and archive index.

## Progress

- [x] Added `scripts/validate-research-audit.sh`.
- [x] Added `scripts/validate-research-audit.test.sh`.
- [x] Wired research audit validation into `make release-check`.
- [x] Updated the prompt-to-artifact checklist and roadmap completion audit.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/validate-research-audit.sh --version v0.2.5
bash scripts/validate-research-audit.test.sh
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
git diff --check
```
