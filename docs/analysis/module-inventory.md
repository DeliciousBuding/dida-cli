# Module Inventory

| Module | Responsibility | Dependencies | Files | Lines | Complexity | S.U.P.E.R Score |
|:--|:--|:--|--:|--:|:--|:--|
| CLI commands | Command routing, parsing, JSON envelope | auth, model, webapi, openapi, officialmcp | many | high | High | S yellow U green P yellow E yellow R yellow |
| Web API client | Dida private Web API requests | standard library | many | medium | Medium | S green U green P yellow E green R yellow |
| OpenAPI client | Official OAuth/OpenAPI requests | standard library | few | medium | Medium | S green U green P yellow E green R yellow |
| npm wrapper | Downloads release binary during postinstall | Node stdlib | 3 | medium | Medium | S green U green P yellow E green R green |
| CI workflows | Test, lint, build, release, publish | GitHub Actions, Go, npm | 2 | medium | Medium | S yellow U green P yellow E yellow R yellow |
| release scripts | Release metadata and artifact validation | bash, node, git | scripts | medium | Medium | S green U green P green E green R green |
| docs/governance | Maintainer and agent operating rules | markdown | many | medium | Medium | S yellow U green P yellow E green R yellow |

## Module Details

### CLI Commands
- **Path**: `internal/cli/`
- **Responsibility**: Parse commands, validate flags, call channel clients, write stable JSON.
- **S.U.P.E.R Assessment**: Validation should happen before auth/network dependencies. The filter validation fix in this run improves P and E by making error behavior independent of local credentials.

### CI Workflows
- **Path**: `.github/workflows/`
- **Responsibility**: Enforce tests, hygiene, builds, releases, npm publish.
- **S.U.P.E.R Assessment**: Inline shell made release behavior harder to test. Extracting metadata and notes logic into scripts improves S and R.

### Release Scripts
- **Path**: `scripts/`
- **Responsibility**: Validate release metadata, packaging templates, archives, private state, and release notes.
- **S.U.P.E.R Assessment**: Scripted gates are replaceable and testable. Shell scripts must keep LF line endings.

### npm Wrapper
- **Path**: `npm/`
- **Responsibility**: Publish a small package that downloads checksum-verified release binaries.
- **S.U.P.E.R Assessment**: Good separation from Go binary. Release workflow must keep npm version aligned with Git tag.
