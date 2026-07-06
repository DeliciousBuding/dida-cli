# Task Breakdown

## Overview

- **Total Phases**: 8
- **Total Tasks**: 20
- **Estimated Total Effort**: M

## S.U.P.E.R Design Constraints

- **S**: Keep release checks in focused scripts instead of large workflow blocks.
- **U**: CI calls scripts; scripts do not depend on workflow-only state unless passed through arguments or environment.
- **P**: Release metadata contract is explicit: tag, npm version, changelog section, main ancestry.
- **E**: Shell scripts use LF; CI test coverage paths avoid `.out` suffix ambiguity.
- **R**: Release notes generation can run locally or in GitHub Actions.
- **P**: Changelog structure and npm package contents are explicit release contracts.
- **P**: Public repository entry points are treated as explicit governance contracts, not informal Markdown.
- **P**: Security automation is a first-class repository contract, enforced by the same governance validator as public docs and templates.
- **P**: External GitHub Actions are pinned to full commit SHAs while keeping version comments for review and update context.
- **P**: Release archives get GitHub artifact attestations derived from the same checksum file that consumers use.

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
| 4.4 | Add npm package README contract | P1 | S | 4.2 | A | P, R | npm package validator tests | None | npm package includes README and the validator fails when it is missing |

### Phase 4 Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 4.1, 4.2, 4.4 | S | Medium | `scripts/`, `Makefile`, `CHANGELOG.md`, `npm/README.md` |
| B | 4.3 | S | Medium | `.github/workflows/release.yml`, `RELEASE.md` |

## Phase 5: Public Repository Governance

**Goal**: Make the public contributor and user entry points cleaner and enforce their required structure in CI.
**S.U.P.E.R Focus**: P, E, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 5.1 | Remove internal metadata from public README | P1 | S | 4.4 | A | P | Governance validator | None | README starts with the visible project identity, not agent frontmatter |
| 5.2 | Strengthen PR and issue contributor templates | P1 | S | 3.1 | A | P, E | Governance validator tests | Update `AGENTS.md` and `RELEASE.md` | PR checklist covers tests, vet, vulnerability scan, private-state guard, docs, changelog, and package gates |
| 5.3 | Add repository governance validator | P1 | S | 5.1, 5.2 | B | P, R | Shell tests and CI hygiene | Update `Makefile` | CI fails when README, npm README, PR template, issue forms, security policy, or contribution guide lose required content |

### Phase 5 Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 5.1, 5.2 | S | Medium | `README.md`, `.github/`, `CONTRIBUTING.md` |
| B | 5.3 | S | Medium | `scripts/`, `Makefile`, `.github/workflows/ci.yml` |

## Phase 6: Supply-Chain Security Automation

**Goal**: Add static analysis and repository security posture checks that run without maintainer intervention.
**S.U.P.E.R Focus**: P, E, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 6.1 | Add CodeQL analysis for Go | P1 | S | 5.3 | A | P, R | actionlint and remote workflow run | Update `AGENTS.md` | CodeQL analyzes Go with extended security queries and uploads code scanning results |
| 6.2 | Add OpenSSF Scorecard and governance enforcement | P1 | S | 5.3 | B | P, E | governance validator tests and actionlint | Update `AGENTS.md` and `CONTRIBUTING.md` | Scorecard publishes SARIF and governance validation fails if security workflows or permissions are removed |

### Phase 6 Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 6.1 | S | Low | `.github/workflows/codeql.yml` |
| B | 6.2 | S | Medium | `.github/workflows/scorecard.yml`, `scripts/validate-repo-governance.sh`, docs |

## Phase 7: Pinned GitHub Actions

**Goal**: Reduce workflow supply-chain drift by pinning external GitHub Actions to full commit SHAs and validating that contract locally and in CI.
**S.U.P.E.R Focus**: P, E, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 7.1 | Pin external workflow actions by SHA | P1 | S | 6.2 | A | P, E | actionlint and remote workflow run | Update `AGENTS.md` and `RELEASE.md` | Every external `uses:` reference in `.github/workflows/` uses a 40-character commit SHA with a version comment |
| 7.2 | Add local pinned-actions validator | P1 | S | 7.1 | A | P, R | Shell tests plus `make release-check` | Update `CONTRIBUTING.md` | CI and release validation fail when a workflow action is tag-pinned or lacks version context |

### Phase 7 Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 7.1, 7.2 | S | Medium | `.github/workflows/`, `scripts/`, `Makefile`, docs |

## Phase 8: Release Archive Provenance

**Goal**: Attach GitHub artifact attestations to release archives without adding long-lived signing keys.
**S.U.P.E.R Focus**: P, E, R

| # | Task | Priority | Effort | Depends On | Lane | S.U.P.E.R | Test Expectation | Memory Impact | Acceptance Criteria |
|:--|:--|:--|:--|:--|:--|:--|:--|:--|:--|
| 8.1 | Generate archive attestations in the release workflow | P1 | S | 7.2 | A | P, E | actionlint and release workflow syntax checks | Update `AGENTS.md` and `RELEASE.md` | Release job has OIDC and attestation permissions and runs pinned `actions/attest` against `dist/checksums.txt` |
| 8.2 | Enforce and document archive provenance | P1 | S | 8.1 | A | P, R | Governance validator tests plus `make release-check` | Update distribution docs | Local governance checks fail if the attestation step or permissions are removed; maintainers have a `gh attestation verify` command |

### Phase 8 Parallel Lanes

| Lane | Tasks | Combined Effort | Merge Risk | Key Files |
|:--|:--|:--|:--|:--|
| A | 8.1, 8.2 | S | Medium | `.github/workflows/release.yml`, `scripts/`, `RELEASE.md`, docs |
