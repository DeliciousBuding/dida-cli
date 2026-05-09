# Roadmap Completion Audit

This audit maps the active DidaCLI objective to concrete repository evidence.
It is intentionally conservative: if a surface lacks live verification, it is
not considered complete.

For the detailed prompt-to-artifact checklist, see
`docs/research/prompt-to-artifact-checklist.md`.

## Objective Criteria

| Criterion | Evidence | Status |
| --- | --- | --- |
| Three explicit channels: Web API, official MCP, official OpenAPI | `README.md`, `docs/commands.md`, `docs/research/api-channel-inventory.md` | Implemented |
| Agent-first JSON command surface | `schema list/show`, `agent context`, stable JSON envelope tests in `internal/cli/cli_test.go` | Implemented |
| Release distribution | `.github/workflows/release.yml`, `install.sh`, `install.ps1`, release `v0.1.9` | Implemented and smoke-tested |
| Root cleanliness | Current root contains only project-level directories/files; generated data stays ignored under `bin/`, `tmp/`, `data/` | Ongoing rule |
| Secrets kept out of repo | Sensitive scans during changes; auth docs use env/stdin/placeholders | Ongoing rule |

## Channel Audit

### Web API

Implemented and documented:

- Sync, agent context, settings, projects, folders, tags, filters, columns,
  comments, tasks, completed history, closed history, trash, search, user
  metadata, sharing reads, calendar reads, statistics, templates, Pomodoro,
  habits, attachment quota, comment attachment create, and raw GET probes.
- Coverage truth: `docs/api-coverage.md`.
- Gap truth: `docs/research/webapi-gap-catalog.md`.

Not complete:

- Task activity detail is blocked by `need_pro` on the current account.
- Comment attachment upload/create is implemented through
  `comment create --file <path>` after reversible live evidence. Task-level
  attachment association, download/preview, and orphan cleanup are still only
  partially mapped.
- Filter writes, column update/delete/order, and collaboration writes still
  need real request-body evidence and rollback plans.

### Official MCP

Implemented and documented:

- Discovery and generic tools: `official doctor/tools/show/call`.
- Local token config helpers: `official token status/set/clear`; environment
  tokens still take precedence over saved config.
- First-class project reads, task search/query/filter/undone reads, batch task
  dry-run wrappers, habit wrappers, and focus wrappers.
- Promotion policy: `docs/research/official-mcp-wrapping-policy.md`.
- Crosswalk: `docs/research/official-mcp-tool-crosswalk.md`.

Not complete:

- Token-based health, tools, project list, project get/data, task
  detail/time-query/search/undone/filter, habit list/sections, and focus range reads
  were live-smoked on 2026-05-10 without committing private payloads.
- Known-id habit/focus reads are blocked on the current account because live
  token smokes found no habits and no focus records, including a 365-day focus
  range.
- Destructive focus delete and task batch writes still need disposable live
  targets.

### Official OpenAPI

Implemented and documented:

- OAuth client config: `openapi client status/set/clear`.
- OAuth helpers: `openapi doctor/status/auth-url/listen-callback/exchange-code/login/logout`.
- Project CRUD/data, task, focus, and habit wrappers.
- OpenAPI guide and notes under `docs/research/`.
- Saved client config plus `openapi auth-url` were verified on 2026-05-10
  without recording secrets or local paths.

Not complete:

- Full OAuth browser approval has not been live-verified on the current account.
- Project/task/focus/habit live calls require a saved OAuth access token.
- Write smokes require disposable live resources.

## Distribution Audit

Implemented:

- `v0.1.9` release exists.
- Release workflow builds Windows, Linux, and macOS assets on amd64/arm64.
- `checksums.txt` is attached.
- Windows installer latest smoke passed against `v0.1.9`.
- WSL Linux installer latest smoke passed against `v0.1.9`.
- Linux/amd64 `install.sh` smoke passed against `v0.1.4` under WSL.
- Installed `v0.1.4` binary smoke passed for `version`,
  `schema show openapi.clientSet`, and `openapi client set/status/clear`.
- npm installer skeleton smoke passed on Windows against `v0.1.4` from a
  temporary copy of `npm/`.
- npm installer skeleton smoke passed on WSL Linux against `v0.1.4`; this also
  verified the Unix wrapper/binary split where `bin/dida` remains a Node wrapper
  and the downloaded binary is stored as `bin/dida-bin`.
- Package manager templates exist for Homebrew and Scoop under `packaging/`,
  pinned to `v0.1.9` release assets and checksums.
- winget submission notes exist under `packaging/winget/`.
- Release workflow now uses action major versions that avoid the Node 20
  deprecation warning observed on earlier release runs.

Remaining:

- macOS installer smoke should be repeated for `v0.1.9` on a native macOS host.
- Homebrew and Scoop templates are not yet published to external package
  repositories.
- winget manifest generation and submission remain deferred until release
  cadence and package identity are final.
- npm installer skeleton is smoke-tested on Windows and WSL Linux but is not
  published.

## Current Blocking Preconditions

1. OpenAPI developer app redirect URL configured to match local callback.
2. Successful OpenAPI OAuth approval to save an access token.
3. Pro account or trace for task activity detail.
4. Disposable files/tasks/projects for task-level attachment and write-flow smoke tests.
5. Disposable targets for Official MCP known-id habit/focus reads and write smoke.

## Next Best Actions

1. Run `dida openapi client set --id <client-id> --secret-stdin --json`
   locally, then complete `dida openapi login --json`.
2. Live-smoke `dida openapi project list --json` after OAuth token persistence.
3. Live-smoke remaining Official MCP read filters with narrow queries, then
   writes only with disposable targets.
4. Capture task-level Web API attachment download/preview and association flows.
5. Keep `docs/api-coverage.md`, `docs/research/*`, schema, skill, and README
   synchronized with every new command.
