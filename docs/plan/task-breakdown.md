# Task Breakdown

## Overview

- **Total Phases**: 4
- **Total Tasks**: 10
- **Estimated Total Effort**: M

## S.U.P.E.R Design Constraints

- **S**: Keep release checks in focused scripts instead of large workflow blocks.
- **U**: CI calls scripts; scripts do not depend on workflow-only state unless passed through arguments or environment.
- **P**: Release metadata contract is explicit: tag, npm version, changelog section, main ancestry.
- **E**: Shell scripts use LF; CI test coverage paths avoid `.out` suffix ambiguity.
- **R**: Release notes generation can run locally or in GitHub Actions.
- **P**: Changelog structure and npm package contents are explicit release contracts.

## Phase 1: Stabilize Main CI

**Goal**: Fix current red main checks.
**S.U.P.E.R Focus**: P, E

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 1.1 | Validate task list filters before auth | P0 | S | - | A | P, E | Add/use regression test with empty config | None | Empty config invalid filter returns validation |
| 1.2 | Fix cross-platform coverage profile path | P0 | S | - | B | E | actionlint and CI run | None | Windows runner no longer treats `.out` as package |

## Phase 2: Release Governance

**Goal**: Make release/tag/changelog/npm gates local-testable.
**S.U.P.E.R Focus**: S, P, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 2.1 | Add release metadata validator | P0 | M | - | A | S, P | Shell tests | Update `RELEASE.md` | Validates semver tag, npm version, changelog |
| 2.2 | Add release notes generator | P0 | M | - | B | S, R | Shell tests | Update `RELEASE.md` | Generates notes from changelog and downloads table |
| 2.3 | Wire scripts into CI and release workflow | P0 | S | 2.1, 2.2 | A | U, R | actionlint | Update AGENTS if rule changes | CI tests scripts; release workflow calls scripts |

## Phase 3: Open-Source Maintenance Polish

**Goal**: Improve maintainer docs and dependency automation.
**S.U.P.E.R Focus**: E, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 3.1 | Add maintainer release guide and Makefile target | P1 | S | 2.1 | A | P, R | `make release-check VERSION=v0.2.4` | Update AGENTS | Maintainers have one local preflight command |
| 3.2 | Add Dependabot for Actions/npm | P2 | S | - | B | R | YAML lint | None | Dependabot config exists for weekly updates |

### Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 1.1, 2.1, 2.3, 3.1 | M | Medium | `internal/cli`, workflows, scripts, docs |
| B | 1.2, 2.2, 3.2 | M | Low | workflows, scripts, dependabot |

## Phase 4: Provenance and Contract Hardening

**Goal**: Align release automation with current open-source supply-chain practice by validating changelog/package contracts locally and preferring short-lived npm publishing credentials.
**S.U.P.E.R Focus**: S, P, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 4.1 | Add changelog structure validator | P1 | S | 2.1 | A | S, P | Shell tests | Update `RELEASE.md` | `Unreleased`, release section, and compare links are validated locally |
| 4.2 | Extract npm package validator | P1 | S | 2.3 | A | S, R | Shell tests plus release-check | npm package files/name/version contract is reusable outside workflow |
| 4.3 | Prefer npm Trusted Publishing/OIDC with token fallback | P1 | S | 4.2 | B | P, E | actionlint and release-check | Update `RELEASE.md` | Workflow can publish with OIDC when configured and still supports `NPM_TOKEN` fallback |

### Phase 4 Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 4.1, 4.2 | S | Medium | `scripts/`, `Makefile`, `CHANGELOG.md` |
| B | 4.3 | S | Medium | `.github/workflows/release.yml`, `RELEASE.md` |
