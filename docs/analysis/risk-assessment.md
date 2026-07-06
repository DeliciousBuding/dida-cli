# Risk Assessment

## S.U.P.E.R Architecture Health Summary

| Principle | Status | Key Findings | Transformation Priority |
|:--|:--|:--|:--|
| S Single Purpose | Yellow | Release workflow had large inline shell blocks. | High |
| U Unidirectional Flow | Green | CLI, clients, and packaging are mostly separated. | Low |
| P Ports over Implementation | Yellow | Release metadata rules were implicit in YAML; CLI filter validation was after auth. | High |
| E Environment-Agnostic | Yellow | Windows shell and coverage path behavior differed from Linux. | High |
| R Replaceable Parts | Yellow | Release notes generation was not reusable outside workflow. | Medium |

**Overall Health**: 1/5 principles fully healthy - refactoring needed for release governance.

**Phase 4 Update**: changelog and npm package validation are now extracted into focused scripts, reducing release workflow duplication and making package metadata contracts locally testable.

**Phase 6 Update**: CodeQL and OpenSSF Scorecard workflows are now treated as repository governance contracts, so removal of the security workflows or their code-scanning permissions fails CI hygiene.

**Phase 7 Update**: external workflow actions are pinned to full commit SHAs, with version comments retained so reviews and Dependabot PRs still show the intended major version.

**Phase 8 Update**: release archives now get GitHub artifact attestations from `dist/checksums.txt`; governance checks require the release OIDC and attestation permissions.

### S.U.P.E.R Violation Hotspots

1. `.github/workflows/release.yml`: inline validation and notes generation.
2. `.github/workflows/ci.yml`: coverage path used `.out`, which was unsafe on Windows runner shell behavior.
3. `internal/cli/task_cmd.go`: filter validation happened after auth/sync load.
4. Shell scripts: LF line ending was not enforced.

## Risk Matrix

| Risk | Impact | Likelihood | Severity | Mitigation |
|:--|:--|:--|:--|:--|
| Tag push discovers missing changelog or npm mismatch too late | Broken release | Medium | High | Scripted `make release-check` and CI tests |
| npm publish token invalid | Release stalls after GitHub release | Medium | High | Existing npm preflight remains before release creation |
| npm long-lived token exposure | Package publishing blast radius | Medium | High | Prefer Trusted Publishing/OIDC, with `NPM_TOKEN` retained only as a fallback |
| changelog compare links drift | Confusing release notes and broken references | Medium | Medium | Validate `Unreleased` and release compare links in CI |
| npm registry listing has no README | New users cannot evaluate install and usage from npm | High | Medium | Include `npm/README.md` and validate package contents before publish |
| public README starts with internal metadata | GitHub visitors see agent-only state before product identity | Medium | Medium | Remove frontmatter and validate README starts with public content |
| contribution templates miss release or secret checks | PRs can bypass repo-specific safety gates | Medium | High | Validate PR template and issue forms in CI |
| static code scanning is absent | security regressions rely only on reviewer attention | Medium | Medium | Run CodeQL for Go with extended security queries |
| repository security posture drifts silently | supply-chain regressions are noticed late | Medium | Medium | Run OpenSSF Scorecard and require workflow permissions in governance validation |
| workflow actions move under a mutable tag | compromised or changed action code runs in CI | Medium | High | Pin external Actions to full commit SHAs and validate the format in CI |
| SHA-pinned actions become stale | security fixes in Actions are missed | Medium | Medium | Keep version comments, Dependabot coverage, and manual review of action update PRs |
| release archive provenance is missing | users can verify checksums but not build origin | Medium | Medium | Generate GitHub artifact attestations for release archives from `checksums.txt` |
| attestation permissions are removed during workflow edits | release provenance silently stops | Medium | Medium | Require `id-token: write`, `attestations: write`, and pinned `actions/attest` in governance validation |
| Windows runner parses coverage path differently | CI red on main | High | High | Use `coverage/profile.txt` through Bash |
| Local shell scripts fail under WSL due CRLF | Maintainer checks unreliable | High | Medium | `.gitattributes` for `*.sh` |

## Testing Risks

Release logic must be unit-tested as shell scripts, not only executed on tag push. CI must run those script tests on every push and PR.

Phase 4 adds coverage for changelog structure and npm package contents, including README presence, so release metadata and registry listing drift are caught before tag pushes. Phase 5 extends the same idea to repository governance files and public entry-point copy.

Phase 6 extends governance coverage to security automation. The local validator now requires CodeQL, Scorecard, SARIF upload, and code-scanning permissions before CI can pass.

Phase 8 extends governance coverage to release archive provenance. The validator now requires the release attestation action, its OIDC permission, and the checksum-based subject input.

## Project Governance Risks

`CLAUDE.md` referenced an active `docs/progress/MASTER.md` that did not exist after the previous archive. This run recreates a current MASTER index for release-governance work.

## Compatibility Concerns

Release tags remain strict `vX.Y.Z`; no prerelease semver is introduced. Package manager templates now track the latest checksum-verified GitHub Release after release assets exist.
