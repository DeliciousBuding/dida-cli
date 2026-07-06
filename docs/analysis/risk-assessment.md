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
| Windows runner parses coverage path differently | CI red on main | High | High | Use `coverage/profile.txt` through Bash |
| Local shell scripts fail under WSL due CRLF | Maintainer checks unreliable | High | Medium | `.gitattributes` for `*.sh` |

## Testing Risks

Release logic must be unit-tested as shell scripts, not only executed on tag push. CI must run those script tests on every push and PR.

## Project Governance Risks

`CLAUDE.md` referenced an active `docs/progress/MASTER.md` that did not exist after the previous archive. This run recreates a current MASTER index for release-governance work.

## Compatibility Concerns

Release tags remain strict `vX.Y.Z`; no prerelease semver is introduced. Existing package manager template versions remain pinned to `v0.2.4` until a new release is prepared.
