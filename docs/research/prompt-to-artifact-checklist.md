# Prompt-To-Artifact Checklist

This checklist maps the active DidaCLI objective and the explicit distribution
productization request to repository artifacts. It is an audit aid, not a claim
that the full roadmap is complete.

Status labels:

- `done`: implemented and verified with local or release evidence
- `partial`: implemented but missing live verification or publication
- `blocked`: waiting on an external account, token, host, or platform

## Active Objective

Build DidaCLI as a production-grade, JSON-first Dida365/TickTick CLI:

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
| Agent channel selection is documented and exposed locally | `dida channel list --json`, `docs/research/api-channel-inventory.md`, `docs/agent-usage.md`, `skills/dida-cli/SKILL.md` | done | The CLI, inventory, agent guide, and repo skill now include job-to-channel selection, auth separation rules, and blocker exit criteria or safe-use notes. |
| Web API read coverage is documented | `docs/api-coverage.md`, `docs/research/webapi-gap-catalog.md` | partial | Task activity detail needs Pro-entitled success evidence; task-level attachment and private write flows need more evidence. |
| Schema discovery is token-efficient | `dida schema list --compact --json`, `dida schema show <id> --json`, `internal/cli/cli_test.go` | done | Compact schema list is the JSON-first command index; `schema show` keeps full HTTP/notes details available on demand. |
| Web API commands prefer JSON and compact output | `internal/cli/*`, `docs/commands.md`, `README.md` | done | `agent context --outline` adds task id references plus a deduplicated `taskIndex` for lower-token agent reads; continue adding compact output when new noisy reads are promoted. |
| Official MCP channel is explicit | `docs/research/official-mcp-tool-crosswalk.md`, `docs/research/official-mcp-vs-webapi.md` | done | Token-based health, tools, project get/data, task get/query/search/undone/filter, habit list/sections, and focus list were live-smoked on 2026-05-10. |
| Official MCP high-value wrappers exist | `internal/cli/official_cmd.go`, `docs/research/official-mcp-wrapping-policy.md` | partial | Core task/project/habit/focus reads are live-smoked where safe IDs exist; task batch-add, batch-update, and complete-project were live-smoked with cleanup; habit writes and focus delete have local dry-run previews; known-id reads need disposable habit and focus records. |
| Official OpenAPI channel is explicit | `docs/research/official-openapi-guide.md`, `docs/research/official-openapi-notes.md` | done | OAuth login and read smokes are verified without committing token material. |
| Official OpenAPI OAuth helpers exist | `internal/cli/openapi_cmd.go`, `internal/openapi/oauth.go`, `internal/openapi/oauth_test.go` | partial | Saved client config, auth-url generation, callback handling, token exchange, and token status are verified; write smokes need disposable targets. |
| Official OpenAPI resource wrappers exist | `internal/cli/openapi_cmd.go`, `docs/commands.md` | partial | Project/task/focus/habit read smokes are verified where safe IDs exist; write smokes need disposable targets. |

## Distribution Request Checklist

| Explicit request | Evidence | Status | Gap |
| --- | --- | --- | --- |
| Tag-push GitHub Release workflow | `.github/workflows/release.yml` | done | None for current workflow. |
| Build Windows amd64/arm64 `dida.exe` | `.github/workflows/release.yml`, release `v0.2.1` assets | done | None. |
| Build Linux amd64/arm64 `dida` | `.github/workflows/release.yml`, release `v0.2.1` assets | done | None. |
| Build Darwin amd64/arm64 `dida` | `.github/workflows/release.yml`, release `v0.2.1` assets | done | Native macOS install smoke is still unavailable. |
| Archive as zip/tar.gz | `.github/workflows/release.yml`, release `v0.2.1` assets | done | None. |
| Generate `checksums.txt` | `.github/workflows/release.yml`, release `v0.2.1` | done | None. |
| Release notes include install methods | `.github/workflows/release.yml` release-notes step | done | None. |
| `install.sh` OS/arch detection and checksum verification | `install.sh` | done | WSL Linux latest smoke passed; macOS native smoke pending. |
| `install.ps1` OS/arch detection and checksum verification | `install.ps1` | done | Windows latest smoke passed. |
| `DIDA_VERSION`, `DIDA_INSTALL_DIR`, `DIDA_REPO` | `install.sh`, `install.ps1` | done | The npm package supports `DIDA_VERSION` and `DIDA_REPO`; npm owns the package binary directory, so `DIDA_INSTALL_DIR` is limited to the standalone install scripts. |
| Install runs `dida version` and `dida doctor --json` | `install.sh`, `install.ps1` | done | npm postinstall intentionally only downloads; wrapper commands are tested separately. |
| README English Quickstart | `README.md`, `docs/quickstart.md` | done | Keep examples synchronized with command changes. |
| README Chinese Quickstart | `README.zh-CN.md`, `docs/quickstart.zh-CN.md` | done | Keep examples synchronized with command changes. |
| LLM/Agent quickstart | `docs/llm-quickstart.md` | done | Keep short and command-first. |
| Agent warning not to paste cookies/tokens | `README.md`, `README.zh-CN.md`, `docs/quickstart*.md`, `docs/llm-quickstart.md` | done | None. |
| npm installer package | `npm/package.json`, `npm/bin/dida`, `npm/scripts/install.js` | done | `@delicious233/dida-cli@0.2.1` is published; package dry-run verifies wrapper, installer, and manifest contents. |
| Homebrew plan | `docs/distribution.md`, `packaging/homebrew/dida.rb` | partial | Template URL/hash static validation passed against `v0.2.1`; no external tap or native install smoke yet. |
| Scoop plan | `docs/distribution.md`, `packaging/scoop/dida.json` | partial | Template JSON and URL/hash static validation passed against `v0.2.1`; external bucket publication and native install smoke remain pending. |
| winget plan | `docs/distribution.md`, `packaging/winget/README.md` | partial | Submission deferred until release cadence and package identity are final. |

## Verification Evidence

Recently run successfully:

- `go test ./...`
- `go vet ./...`
- `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
- `git diff --check`
- local path and known secret scan
- Official MCP local dry-run smokes without `DIDA365_TOKEN`:
  `official habit create`, `official habit update`, `official habit checkin`,
  and `official focus delete`
- Official MCP live reversible task write smoke on 2026-05-10:
  `official task batch-add`, `official task batch-update`,
  `official task get --project`,
  `official task complete-project`, followed by Web API `task delete` cleanup
- Windows `install.ps1` latest smoke
- WSL Linux `install.sh` latest smoke
- Windows and WSL Linux pinned `DIDA_VERSION` install smoke passed
- Installer smoke covered release binary startup outside the repository,
  including `version`, `schema show official.call`, `openapi doctor`, and Web
  API `task list`
- Windows npm installer smoke
- WSL Linux npm installer smoke
- Windows npm latest smoke using release
  `latest/download/checksums.txt` instead of the GitHub API
- Installed binary OpenAPI client config smoke:
  `openapi client set/status/clear`
- Web API `auth status --verify`, `agent context`, `attachment quota`, and
  empty `comment list` live reads on 2026-05-10
- Web API `agent context --outline --json` live read on 2026-05-10 with
  deduplicated task index and task id references
- Web API comment attachment upload/create was implemented as
  `comment create --file <path>` after reversible live evidence confirmed
  multipart field `file`, upload response keys, comment attach payload,
  read-back, and cleanup; additional upload smokes need disposable attachment
  quota
- Web API task activity raw probes on 2026-05-10 confirmed that the surface
  remains blocked or unstable, so it stays out of first-class commands
- Scoop manifest JSON parse
- Homebrew/Scoop template URL and checksum static validation against current
  release checksums
- Homebrew formula install path checked against release archive layout
- Scoop `extract_dir` checked against Windows release zip layout
- release checksum comparison against current release checksums

Skipped or blocked verification:

- native macOS install smoke remains pending
- Homebrew formula syntax/install smoke remains pending on a packaging host
  with `brew` and `ruby`
- Scoop install smoke remains pending on a Windows host with Scoop
- winget manifest generation remains pending until `wingetcreate` validation
  is run
- Official MCP habit/focus write smoke: task write smoke succeeded with cleanup;
  habit/focus writes still need disposable targets
- Official MCP known-id habit/focus reads need disposable habit and focus records.
- Official OpenAPI live smokes need a disposable OAuth session.
- Web API task activity detail needs a Pro account or a browser trace that
  confirms the response shape.
- Additional Web API comment attachment upload smokes need a disposable account
  with attachment quota available.

## Completion Rule

Do not mark the roadmap complete until every `partial` or `blocked` item above
is either implemented and verified, explicitly deferred with rationale, or tied
to a durable external precondition with enough evidence for another agent to
resume without rediscovery.
