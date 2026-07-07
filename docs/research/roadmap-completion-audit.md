# Roadmap Completion Audit

This audit maps the active DidaCLI objective to concrete repository evidence.
It is conservative: a surface needs command coverage, tests, and safe live
evidence before it is marked complete.

For the detailed prompt-to-artifact checklist, see
`docs/research/prompt-to-artifact-checklist.md`.

## Objective Criteria

| Criterion | Evidence | Status |
| --- | --- | --- |
| Three explicit channels: Web API, official MCP, official OpenAPI | `README.md`, `docs/commands.md`, `docs/research/api-channel-inventory.md` | Implemented |
| JSON command surface for automation | `schema list/show`, `agent context`, stable JSON envelope tests in `internal/cli/cli_test.go` | Implemented |
| Release distribution | `.github/workflows/release.yml`, `install.sh`, `install.ps1`, release `v0.2.5`, npm `@delicious233/dida-cli@0.2.5` | Implemented and smoke-tested |
| Root cleanliness | Current tracked root contains only project-level directories/files; generated data stays ignored under `bin/`, `tmp/`, and `data/private/` | Ongoing rule |
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

- Task activity detail needs a Pro account or browser trace that proves the
  success response shape.
- Comment attachment upload/create is implemented through
  `comment create --file <path>` after reversible live evidence, including a
  2026-05-10 repeat smoke with a 1x1 PNG, read-back, comment delete, and
  disposable task cleanup. Task-level
  upload plus render/download path shapes are bundle-mapped, but task-level
  association/persistence, file limits, and orphan cleanup still need a
  reversible trace.
- Filter writes, column update/delete/order, and collaboration writes still
  need real request-body evidence and rollback plans.

### Official MCP

Implemented and documented:

- Discovery and generic tools: `official doctor/tools/show/call`.
- Local token config helpers: `official token status/set/clear`; environment
  tokens still take precedence over saved config.
- First-class project reads, task search/query/filter/undone reads, batch task
  dry-run wrappers, habit wrappers, and focus wrappers.
- Local dry-run previews exist for official MCP task batch writes, habit
  create/update/checkin, and focus delete without requiring a saved token.
- Token-based health, tools, project list, project get/data, task
  detail/time-query/search/undone/filter, habit list/sections, and focus range
  reads were live-smoked on 2026-05-10 without committing private payloads.
- Official MCP task `batch-add`, `batch-update`, project-scoped `task get`,
  and `complete-project` were live-smoked on 2026-05-10 with disposable tasks,
  then cleaned up through the verified Web API task delete path.
- Promotion policy: `docs/research/official-mcp-wrapping-policy.md`.
- Crosswalk: `docs/research/official-mcp-tool-crosswalk.md`.

Not complete:

- Known-id habit/focus reads need disposable habit and focus records.
- Destructive focus delete and habit write smokes still need disposable live
  targets.

### Official OpenAPI

Implemented and documented:

- OAuth client config: `openapi client status/set/clear`.
- OAuth helpers: `openapi doctor/status/auth-url/listen-callback/exchange-code/login/logout`.
- Project CRUD/data, task, focus, and habit wrappers.
- OpenAPI guide and notes under `docs/research/`.
- Redacted client config plus `openapi auth-url` were verified on 2026-05-10.
- `openapi login --browser` now validates loopback callback URLs, honors local
  `--redirect-uri`, and fails fast with one JSON error when callback setup is
  invalid.
- 2026-05-10 redacted OAuth verification succeeded, and
  `dida openapi status --json` plus `dida openapi project list --json`
  returned valid response shapes without committing token material or private
  project data.
- 2026-05-10 follow-up live read smokes succeeded for `openapi project get`,
  `openapi project data`, project-scoped `openapi task get`, bounded
  `openapi task filter`, bounded `openapi task completed`, bounded
  `openapi focus list` for type 0 and type 1, and `openapi habit list`.
  Private task and project payloads were not committed.

Not complete:

- Known-id OpenAPI `habit get` and `focus get` need disposable habit and focus
  records.
- OpenAPI write smokes still need disposable live resources.

## Distribution Audit

Implemented:

- `v0.2.5` release exists.
- Release workflow builds Windows, Linux, and macOS assets on amd64/arm64.
- `checksums.txt` is attached.
- Windows installer latest smoke passed against the current release line.
- WSL Linux installer latest smoke passed against the current release line.
- Pinned release install smoke passed on Windows and WSL Linux.
- Installer smoke covered release binary startup outside the repository.
- Linux/amd64 `install.sh` smoke passed under WSL.
- Installed binary smoke passed for `version`, `doctor --json`, and
  `openapi client set/status/clear`.
- npm installer smoke passed on Windows; this
  verified download/checksum, wrapper startup, `version`, and `doctor --json`.
- npm installer smoke passed on WSL Linux; this also
  verified the Unix wrapper/binary split where `bin/dida` remains a Node wrapper
  and the downloaded binary is stored as ignored `bin/dida-bin`.
- npm package `@delicious233/dida-cli@0.2.5` is published, and npm registry
  metadata includes the package README.
- Package manager templates exist for Homebrew and Scoop under `packaging/`,
  pinned to `v0.2.5` release assets and checksums.
- Homebrew and Scoop template URL/hash static validation passed against the
  `v0.2.5` release `checksums.txt` for all six release assets.
- Homebrew formula install path logic was checked against the release archive
  layout: assets unpack under a top-level `dida_v..._<os>_<arch>/` directory,
  so the formula locates the nested `dida` binary before `bin.install`.
- Scoop `extract_dir` was checked against the Windows release zip layout:
  `dida.exe` lives under `dida_v..._windows_<arch>/`.
- `scripts/export-package-manager-repos.sh` exports repo roots for
  `DeliciousBuding/homebrew-dida` and `DeliciousBuding/scoop-bucket`.
- Release workflow exports package-manager repo layouts as
  `dida-package-manager-repos-vX.Y.Z` after release checksums exist. The next
  tag release will produce the artifact.
- winget submission notes exist under `packaging/winget/`, and
  `scripts/winget-submission-preflight.sh` checks the current release URL,
  package id, and `winget validate --manifest` handoff before manifest
  generation.
- Release workflow now uses action major versions that avoid the Node 20
  deprecation warning observed on earlier release runs.

Remaining:

- macOS installer smoke should be repeated on a native macOS host.
- Homebrew and Scoop templates are not yet published to external package
  repositories, and native package-manager install smoke remains pending.
- winget manifest generation and submission remain deferred until release
  cadence and package identity are final and `winget validate --manifest`
  passes on a Windows packaging host.
- npm installer package is published; native macOS npm smoke remains pending.

## Current Blocking Preconditions

1. Pro account or trace for task activity detail.
2. Disposable files/tasks/projects for task-level attachment and write-flow smoke tests.
3. Disposable targets for Official MCP known-id habit/focus reads and habit/focus write smoke.
4. Disposable OpenAPI tasks/projects/habits/focus records for write smokes.

2026-05-10 recheck:

- Web API cookie auth still verifies successfully.
- Official MCP still connects; known-id habit/focus smokes need disposable
  records.
- OpenAPI read smokes returned valid response shapes without committing private
  payloads.
- Additional upload smokes need disposable attachment quota.

## Next Best Actions

1. Create disposable OpenAPI task/project/habit/focus targets and live-smoke
   write paths with cleanup.
2. Live-smoke remaining Official MCP read filters with narrow queries, then
   writes only with disposable targets.
3. Capture task-level Web API attachment download/preview and association flows.
4. Keep `docs/api-coverage.md`, `docs/research/*`, schema, skill, and README
   synchronized with every new command.
