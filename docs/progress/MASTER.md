# Release Governance Optimization - Progress Tracker

> **Task**: Clean up DidaCLI CI/CD, Actions, Release, npm publish, tag, and changelog governance.
> **Started**: 2026-07-06
> **Last Updated**: 2026-07-06
> **Mode**: GITHUB_STANDARD
> **Repo**: DeliciousBuding/dida-cli

## References

- [Project Overview](../analysis/project-overview.md)
- [Module Inventory](../analysis/module-inventory.md)
- [Risk Assessment](../analysis/risk-assessment.md)
- [Task Breakdown](../plan/task-breakdown.md)
- [Dependency Graph](../plan/dependency-graph.md)
- [Milestones](../plan/milestones.md)

## Phase Checklist

- [x] Phase 1: Stabilize Main CI (2/2 tasks)
- [x] Phase 2: Release Governance (3/3 tasks)
- [x] Phase 3: Open-Source Maintenance Polish (2/2 tasks)
- [x] Phase 4: Provenance and Contract Hardening (4/4 tasks)
- [x] Phase 5: Public Repository Governance (3/3 tasks)

## Current Status

**Active Phase**: Release publication
**Active Task**: Prepare and publish `v0.2.5`
**Blockers**: None. Local release checks and core verification passed; next step is push, tag, and release workflow verification.

## Governance Status

**Shared instruction surface**: `AGENTS.md`
**Claude Code instruction surface**: `CLAUDE.md`
**Other platform rule surfaces**: none
**Memory surface**: native Codex memory used for prior DidaCLI release context; repo fallback not created.
**Memory fallback path**: none

## Adaptive Control State

```yaml
adaptive:
  drift_score: 0
  strategy: "single-branch release-governance hardening"
  thresholds:
    annotate: 3
    replan: 6
    rescope: 9
  total_tasks: 14
  completed_tasks: 14
  last_updated: "2026-07-07"
```

## Task Telemetry Log

| Date | Task | Actual Effort | S.U.P.E.R Score | Unplanned Dependencies | Notes |
|:--|:--|:--|:--|--:|:--|
| 2026-07-06 | 1.1 | S | P/E pass | 0 | Reproduced empty-config invalid-filter failure before fix. |
| 2026-07-06 | 1.2 | S | E pass | 0 | Coverage profile moved to `coverage/profile.txt` with Bash shell. |
| 2026-07-06 | 2.1 | M | S/P/R pass | 1 | Added LF shell normalization via `.gitattributes`. |
| 2026-07-06 | 2.2 | M | S/R pass | 0 | Release notes generation now tested locally. |
| 2026-07-06 | 2.3 | S | U/R pass | 0 | CI and release workflows call scripts. |
| 2026-07-06 | 3.1 | S | P/R pass | 0 | Added `RELEASE.md` and `make release-check`. |
| 2026-07-06 | 3.2 | S | R pass | 0 | Added Dependabot for Actions and npm. |
| 2026-07-06 | 4.1 | S | S/P pass | 0 | Added tested changelog structure validation. |
| 2026-07-06 | 4.2 | S | S/R pass | 0 | Extracted npm package contents validation from workflow. |
| 2026-07-06 | 4.3 | S | P/E pass | 0 | Release workflow prefers npm Trusted Publishing/OIDC and retains `NPM_TOKEN` fallback. |
| 2026-07-06 | 4.4 | S | P/R pass | 0 | Added npm package README and made README presence part of package validation. |
| 2026-07-07 | 5.1 | S | P pass | 0 | Removed internal agent metadata from public README. |
| 2026-07-07 | 5.2 | S | P/E pass | 0 | Strengthened PR checklist, issue secret warnings, and contributing verification steps. |
| 2026-07-07 | 5.3 | S | P/R pass | 0 | Added CI-tested repository governance validator. |
| 2026-07-07 | Release publication | M | S/P/R pass | 1 | Prepared `v0.2.5`, verified npm tarball includes `README.md`, and kept package-manager checksum templates tied to the latest checksum-verified release until new assets exist. |

## Next Steps

1. Push the `v0.2.5` release-prep commit and wait for main CI/Pages.
2. Create and push the annotated `v0.2.5` tag after main CI is green.
3. Verify GitHub Release assets, npm provenance, and npm README rendering after the release workflow finishes.

## Session Log

| Date | Session | Summary |
|:--|:--|:--|
| 2026-07-06 | PR #2 recovery | Reopened and merged PR #2 correctly, then identified red main CI. |
| 2026-07-06 | Release governance hardening | Fixed CI root causes and added scripted release gates, docs, and automation. |
| 2026-07-06 | Remote verification | Confirmed PR #2 is `MERGED`; latest CI, Pages rerun, and Dependabot update checks passed on `main`. |
| 2026-07-06 | Provenance and contract hardening | Added tested changelog/npm package validators and OIDC-first npm publish path with token fallback. |
| 2026-07-06 | npm package listing polish | Added `npm/README.md` so the next npm publish fixes the registry README warning. |
| 2026-07-07 | Public governance gate | Removed README frontmatter and added a tested governance validator for public repo entry points. |
| 2026-07-07 | v0.2.5 release prep | Moved governance changes from Unreleased to `v0.2.5`, bumped npm package metadata, and adjusted packaging validation so pre-release checks do not require fake future checksums. |
