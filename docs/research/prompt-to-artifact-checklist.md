# Prompt-To-Artifact Checklist

This checklist maps the active DidaCLI objective and the explicit distribution
productization request to repository artifacts. It is an audit aid, not a claim
that the full roadmap is complete.

Status labels:

- `done`: implemented and verified with local or release evidence
- `partial`: implemented but missing live verification or publication
- `blocked`: waiting on an external account, token, host, or platform

## Active Objective

Build DidaCLI as a production-grade, agent-first Dida365/TickTick CLI:

- three explicit channels: Web API, official MCP, official OpenAPI
- API excavation docs before high-value first-class commands
- unit tests, CLI smoke tests, safe live tests, docs, and schema updates
- repository packaging, README/docs, agent skill, performance, and security
  boundaries
- no secrets in the repository
- small scoped commits pushed promptly

## Channel Checklist

| Requirement | Evidence | Status | Gap |
| --- | --- | --- | --- |
| Web API channel is explicit | `README.md`, `docs/commands.md`, `docs/api-coverage.md`, `docs/research/api-channel-inventory.md` | done | Private Web API remains drift-prone by design. |
| Web API read coverage is documented | `docs/api-coverage.md`, `docs/research/webapi-gap-catalog.md` | partial | Task activity detail still blocked by `need_pro`; attachment and write flows need more evidence. |
| Web API commands prefer JSON and compact output | `internal/cli/*`, `docs/commands.md`, `README.md` | done | Continue adding compact output when new noisy reads are promoted. |
| Official MCP channel is explicit | `docs/research/official-mcp-tool-crosswalk.md`, `docs/research/official-mcp-vs-webapi.md` | done | Token-based health, tools, project get/data, task get/query/search/undone/filter, habit list/sections, and focus list were live-smoked on 2026-05-10. |
| Official MCP high-value wrappers exist | `internal/cli/official_cmd.go`, `docs/research/official-mcp-wrapping-policy.md` | partial | Core task/project/habit/focus reads are live-smoked where safe IDs exist; known-id detail reads and write wrappers still need safe targets. |
| Official OpenAPI channel is explicit | `docs/research/official-openapi-guide.md`, `docs/research/official-openapi-notes.md` | done | OAuth live approval is not complete on this machine. |
| Official OpenAPI OAuth helpers exist | `internal/cli/openapi_cmd.go`, `internal/openapi/oauth.go`, `internal/openapi/oauth_test.go` | partial | Saved client config and auth-url generation are verified; OAuth browser approval and saved access token are still missing. |
| Official OpenAPI resource wrappers exist | `internal/cli/openapi_cmd.go`, `docs/commands.md` | partial | Project/task/focus/habit live calls need a saved OAuth token. |

## Distribution Request Checklist

| Explicit request | Evidence | Status | Gap |
| --- | --- | --- | --- |
| Tag-push GitHub Release workflow | `.github/workflows/release.yml` | done | None for current workflow. |
| Build Windows amd64/arm64 `dida.exe` | `.github/workflows/release.yml`, release `v0.1.9` assets | done | None. |
| Build Linux amd64/arm64 `dida` | `.github/workflows/release.yml`, release `v0.1.9` assets | done | None. |
| Build Darwin amd64/arm64 `dida` | `.github/workflows/release.yml`, release `v0.1.9` assets | done | Native macOS install smoke is still unavailable. |
| Archive as zip/tar.gz | `.github/workflows/release.yml`, release `v0.1.9` assets | done | None. |
| Generate `checksums.txt` | `.github/workflows/release.yml`, release `v0.1.9` | done | None. |
| Release notes include install methods | `.github/workflows/release.yml` release-notes step | done | None. |
| `install.sh` OS/arch detection and checksum verification | `install.sh` | done | WSL Linux latest smoke passed against `v0.1.9`; macOS native smoke pending. |
| `install.ps1` OS/arch detection and checksum verification | `install.ps1` | done | Windows latest smoke passed against `v0.1.9`. |
| `DIDA_VERSION`, `DIDA_INSTALL_DIR`, `DIDA_REPO` | `install.sh`, `install.ps1`, `npm/scripts/install.js` | done | npm uses `DIDA_INSTALL_DIR` only for local smoke isolation when invoked directly; package installs into package `bin/`. |
| Install runs `dida version` and `dida doctor --json` | `install.sh`, `install.ps1` | done | npm postinstall intentionally only downloads; wrapper commands are tested separately. |
| README English Quickstart | `README.md`, `docs/quickstart.md` | done | Keep examples synchronized with command changes. |
| README Chinese Quickstart | `README.zh-CN.md`, `docs/quickstart.zh-CN.md` | done | Keep examples synchronized with command changes. |
| LLM/Agent quickstart | `docs/llm-quickstart.md` | done | Keep short and command-first. |
| Agent warning not to paste cookies/tokens | `README.md`, `README.zh-CN.md`, `docs/quickstart*.md`, `docs/llm-quickstart.md` | done | None. |
| npm installer skeleton | `npm/package.json`, `npm/bin/dida`, `npm/scripts/install.js` | partial | Smoke-tested on Windows and WSL Linux; package is not published. |
| Homebrew plan | `docs/distribution.md`, `packaging/homebrew/dida.rb` | partial | Template exists; no external tap or native install smoke yet. |
| Scoop plan | `docs/distribution.md`, `packaging/scoop/dida.json` | partial | Template exists; no external bucket or Scoop install smoke on this machine. |
| winget plan | `docs/distribution.md`, `packaging/winget/README.md` | partial | Submission deferred until release cadence and package identity are final. |

## Verification Evidence

Recently run successfully:

- `go test ./...`
- `go vet ./...`
- `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- `git diff --check`
- local path and known secret scan
- Windows `install.ps1` latest smoke against `v0.1.9`
- WSL Linux `install.sh` latest smoke against `v0.1.9`
- Windows npm installer smoke against `v0.1.4`
- WSL Linux npm installer smoke against `v0.1.4`
- WSL Linux `install.sh` smoke against `v0.1.4`
- Web API `auth status --verify`, `agent context`, `attachment quota`, and
  empty `comment list` live reads on 2026-05-10
- Web API task activity raw probes on 2026-05-10 confirmed the surface remains
  blocked or unstable rather than command-ready
- Scoop manifest JSON parse
- release checksum comparison against `v0.1.9/checksums.txt`

Skipped or blocked verification:

- native macOS install smoke: no macOS host in this environment
- Homebrew formula syntax/install smoke: `brew` and `ruby` are unavailable here
- Scoop install smoke: `scoop` is unavailable here
- Official MCP write smoke: read-only token smoke succeeded, but writes need
  disposable targets
- Official OpenAPI live smoke: saved client config is present, but no OAuth
  access token is present
- Web API task activity detail: current account receives `need_pro`

## Completion Rule

Do not mark the roadmap complete until every `partial` or `blocked` item above
is either implemented and verified, explicitly deferred with rationale, or tied
to a durable external precondition with enough evidence for another agent to
resume without rediscovery.
