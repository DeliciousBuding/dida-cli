# Risk Assessment

## S.U.P.E.R Health

| Principle | Status | Finding | Priority |
|:--|:--|:--|:--|
| Single Purpose | Green | Completion generation is isolated in `completion_cmd.go`. | Low |
| Unidirectional Flow | Green | Root dispatch calls the completion command; completion does not call back into network or auth code. | Low |
| Ports over Implementation | Green | The command contract is recorded in schema output and docs. | Medium |
| Environment-Agnostic | Green | Scripts are emitted to stdout; install paths remain examples in docs. | Low |
| Replaceable Parts | Green | Shell templates can be edited without touching auth or API clients. | Low |

## Risks

| Risk | Impact | Mitigation |
|:--|:--|:--|
| Generated scripts become stale as root commands change | Completion misses commands | Tests assert core command names; the command list is centralized in `completionCommands`. |
| `--json` confusion | Scripts could be wrapped in JSON accidentally | `dida completion ... --json` fails with a validation error. |
| Shell-specific syntax errors | Completion fails for one shell | Keep templates small and static; unit tests check shell-specific markers. |

## Testing Risk

The test suite checks output shape and command inclusion, not live shell installation. That is acceptable for this slice because the command only writes script text.
