# DidaCLI Roadmap

This roadmap is the execution map for turning DidaCLI into a complete,
agent-friendly, production-grade multi-channel CLI for Dida365 / TickTick.

It is written so another agent can pick up any phase and execute it directly.

## Goal

Ship a stable CLI with three explicit upstream channels:

1. `webapi`
2. `official mcp`
3. `official openapi`

The end state is not "one giant code dump". The end state is:

- broad coverage
- clean command surface
- stable JSON
- explicit auth boundaries
- strong docs
- repeatable verification

## Current Baseline

As of the current main branch:

- `webapi` is the primary implemented channel
- `webapi` has live-smoked trash pagination reads through
  `trash list --cursor/--limit --compact --json`
- `official mcp` has:
  - `official doctor`
  - `official tools`
  - `official show`
  - `official call`
  - local token config helpers under `official token`
  - first-class project read wrappers
  - first-class task read/filter/query wrappers
  - first-class official task batch-add, batch-update, and project completion
    wrappers with local dry-run previews
  - first-class habit wrappers
  - first-class focus wrappers
- `official openapi` has:
  - OAuth foundation
  - auth URL generation
  - callback listener
  - code exchange
  - token persistence
  - project wrappers
  - task wrappers
  - focus wrappers
  - habit wrappers
- docs already include:
  - API coverage matrix at `docs/api-coverage.md`
  - Web API notes
  - OpenAPI guide
  - MCP vs Web API comparison
  - channel inventory
  - Web API gap catalog
  - MCP tool crosswalk
- distribution is live:
  - `v0.1.10` GitHub Release exists
  - release assets cover Windows, Linux, and macOS on amd64/arm64
  - `install.ps1` has been smoke-tested against the latest `v0.1.10` release asset on Windows
  - installed `v0.1.4` binary has smoke-tested `openapi client set/status/clear`
  - `install.sh` has been smoke-tested against the latest `v0.1.10` release asset on Linux/amd64
  - npm installer skeleton has been smoke-tested against `v0.1.4` on Windows
    and WSL Linux
  - package manager templates exist under `packaging/` for Homebrew and Scoop;
    winget is documented as a future submission

## Non-Negotiable Rules

- Do not mix the three auth models.
- Do not write secrets into the repo.
- Do not expose raw writes.
- Do not ship guessed private write payloads.
- Every new command must update:
  - schema
  - docs
  - tests
- Every new read or write surface must be validated against real runtime
  behavior, not just bundle text or docs.

## Channel Definitions

### Web API

- Auth: browser cookie `t`
- Role: widest coverage, main task surface
- Risk: private and drift-prone

### Official MCP

- Auth: `DIDA365_TOKEN=dp_...` or saved local official token config
- Role: official token-based tool surface
- Risk: smaller surface, but cleaner contracts

### Official OpenAPI

- Auth: OAuth access token
- Role: official REST channel
- Risk: more setup, narrower than Web API, still partially unverified live

## Success Criteria

The roadmap is complete only when:

- all three channels have clear auth flows
- the CLI exposes first-class commands where they add real value
- every remaining known gap is either:
  - implemented
  - intentionally deferred with documented reason
  - blocked by an external precondition with documented evidence
- docs tell a new contributor or agent:
  - what exists
  - what is missing
  - what to do next

## Workstream A: Finish Web API Coverage

This remains the highest-value channel because it covers the most.

### A1. Read Gaps

- Task activity detail stream
  - target: `GET /task/activity/{taskId}` on the legacy v1 base
  - current evidence: real webapp target is legacy v1
    `GET /task/activity/{taskId}` with optional `skip` and `lastId`
  - blocker: observed account reaches the v1 route but receives `need_pro`
  - need: Pro-account response shape and cursor semantics

Acceptance:

- command exists
- schema exists
- live read succeeds
- docs updated

### A2. Attachment Flows

- comment attachment upload
- attach to comment
- download / reference model

Current evidence:

- comment attachment upload/display paths, multipart field `file`, upload
  response keys, and comment create payload shape are documented in
  `docs/research/webapi-attachment-flow-notes.md`
- `comment create --file <path>` is implemented and covered by schema, docs,
  dry-run tests, multipart request-shape tests, and reversible live evidence
- the CLI checks attachment quota before upload; the current observed account
  later reached `underQuota=false`, so future live upload smokes need available
  attachment quota
- task-level upload and render/download path shapes are now bundle-mapped:
  `/api/v1/attachment/upload/{projectId}/{taskId}/{attachmentId}` and
  `/api/v1/attachment/{projectId}/{taskId}/{attachmentId}` with optional
  `action=download` or `action=preview`
- task-level attachment association/persistence, accepted file matrix, and
  uploaded-but-not-attached cleanup behavior remain unverified

Acceptance:

- comment multipart flow fully mapped
- reversible live test for comment attachments
- task-level flow documented as research-only until a reversible association
  trace exists
- no secrets or file dumps committed

### A3. Filter and Column Writes

- `POST /batch/filter`
- `POST /batch/columnProject`

Acceptance:

- request bodies verified from real traffic
- dry-run surface exists
- rollback path documented
- commands stay conservative until semantics are confirmed

### A4. Collaboration Writes

- invite create/delete
- permission changes
- member-level write actions

Acceptance:

- only after multi-user semantics are understood
- must document rollback and operator risk

## Workstream B: Productize Official MCP

Current state is generic introspection plus generic call.
Next step is selective promotion, not blind duplication.

### B1. Promote High-Value MCP Tools

Priority list:

1. `list_projects` - implemented as `official project list`
2. `get_project_by_id` - implemented as `official project get`
3. `get_project_with_undone_tasks` - implemented as `official project data`
4. `complete_tasks_in_project` - implemented as `official task complete-project`
5. `batch_add_tasks` - implemented as `official task batch-add`
6. `batch_update_tasks` - implemented as `official task batch-update`
7. `get_task_by_id` - implemented as `official task get`
8. `get_task_in_project` - implemented as `official task get --project`
9. `filter_tasks`
10. `list_undone_tasks_by_date`
11. `search_task`
12. `list_undone_tasks_by_time_query` - implemented as `official task query`
13. `list_habits` - implemented as `official habit list`
14. `list_habit_sections` - implemented as `official habit sections`
15. `get_habit`
16. `create_habit`
17. `update_habit`
18. `upsert_habit_checkins`
19. `get_habit_checkins` - implemented as `official habit checkins`
20. `get_focus`
21. `get_focuses_by_time`
22. `delete_focus`

Acceptance:

- each promoted command must be better than `official call`
- each promoted command must have a reason documented in the crosswalk
- broad official task writes must support local `--dry-run` previews before
  requiring `DIDA365_TOKEN`

Live evidence:

- 2026-05-10 token-based smoke succeeded for `official doctor`,
  `official tools`, generic `list_projects`, `official project get`,
  `official project data`,
  `official task query --query today`, `official task search`, `official task
  undone`, `official task filter`, and bounded `official focus list` for both
  focus types.
- Known-id habit/focus read smokes are currently blocked by account state:
  2026-05-10 token smokes found zero habits and zero focus records, including a
  365-day focus range.
- Write smokes remain blocked until disposable project/task/habit/focus targets
  exist.

### B2. MCP Contract Layer

- keep `official call` generic
- add compact output for common tools when practical
- preserve raw schema access via `official show`

Acceptance:

- official channel remains usable for exploration
- first-class wrappers do not hide tool names or schemas

## Workstream C: Complete Official OpenAPI

This channel is not complete until a real OAuth flow is verified live.

### C1. OAuth Login Experience

- save OAuth client config with `openapi client set --id ... --secret-stdin`
- finish `openapi login`
- make it pleasant for human and agent use
- keep auth flow separated from browser cookie auth
- 2026-05-10: saved client config, `openapi doctor`, and `openapi auth-url`
  verified locally without committing secrets or local paths; browser approval
  and token persistence still need completion.
- 2026-05-10: `openapi login` now supports explicit `--browser`, honors local
  `--redirect-uri`, rejects non-loopback callback hosts before browser launch,
  and returns immediate JSON errors when the callback listener cannot be
  configured. Unit tests cover callback normalization and invalid callback
  shapes.

Acceptance:

- client id and secret are available from env or saved local config
- a user can start login, authorize in browser, and persist an access token
- token status is visible through `openapi doctor` / `openapi status`

### C2. Live Channel Verification

- verify `project list`
- add live verification for at least one resource in:
  - task
  - project
  - focus
  - habit

Acceptance:

- no more "research only" status for OpenAPI
- live resource calls succeed with a real OAuth token

### C3. OpenAPI First-Class Commands

Priority:

1. `openapi project list/get/data/create/update/delete` - implemented with local dry-run previews for writes
2. `openapi task get/create/update/complete/delete/move`
3. `openapi focus list/get/delete`
4. `openapi habit list/get/create/update/checkins`

Acceptance:

- commands stay clearly namespaced under `openapi`
- no confusion with MCP or Web API auth

## Workstream D: CLI Product Quality

### D1. Output Quality

- compact mode where payloads are too noisy
- stable envelopes everywhere
- bounded list output by default

### D2. Safety and Stability

- keep token handling local
- guard against unsafe deletes
- keep error bodies sanitized by default
- avoid accidental output bloat

### D3. Performance

- avoid unnecessary repeated syncs
- reduce duplicate calls where bundle commands can be packed
- keep official channel sessions reusable where appropriate

## Workstream E: Documentation Productization

### E1. Core Docs

- keep `README.md` English-only
- keep `README.zh-CN.md` Chinese-only
- keep `docs/commands.md` as user-facing command reference
- keep `docs/api-coverage.md` as implementation truth for Web API

### E2. Research Docs

Maintain and evolve:

- `api-channel-inventory.md`
- `official-mcp-tool-crosswalk.md`
- `official-mcp-vs-webapi.md`
- `official-openapi-guide.md`
- `official-openapi-notes.md`
- `webapi-gap-catalog.md`

### E3. Missing Docs To Add

Added and now maintained:

- `docs/research/README.md`
- `docs/research/official-channel-validation-matrix.md`
- `docs/research/official-mcp-wrapping-policy.md`
- `docs/research/webapi-probe-log.md`
- `docs/research/openapi-live-validation-log.md`
- `docs/research/roadmap-completion-audit.md`

## Workstream F: Distribution

Priority order:

1. GitHub Releases
2. `install.sh` / `install.ps1`
3. npm installer
4. Homebrew / Scoop
5. winget

### F1. GitHub Releases

- tag-triggered release workflow
- multi-platform binaries:
  - Windows amd64/arm64
  - Linux amd64/arm64
  - macOS amd64/arm64
- archive assets:
  - Windows `.zip`
  - Linux/macOS `.tar.gz`
- `checksums.txt`
- release notes with install commands

Status: implemented and smoke-tested through `v0.1.10`; release workflow uses Node 24-compatible action major versions.

### F2. Install Scripts

- `install.sh`
- `install.ps1`
- OS/arch detection
- latest release download by default
- `DIDA_VERSION`, `DIDA_INSTALL_DIR`, `DIDA_REPO`
- checksum verification
- install-time `dida version` and `dida doctor --json`

Status: implemented; Windows `install.ps1` and WSL Linux `install.sh` latest
smokes passed against `v0.1.10`. The installed-binary OpenAPI client config
smoke passed against `v0.1.4`.

### F3. npm Installer

- placeholder package under `npm/`
- package name placeholder: `@vectorcontrol/dida-cli`
- postinstall downloads matching GitHub Release binary
- `bin/dida` forwards to the downloaded binary

Status: skeleton smoke-tested on Windows and WSL Linux against `v0.1.4`; do not
publish until package ownership, final npm scope, macOS smoke, and publishing
policy are confirmed.

### F4. Homebrew / Scoop

- Homebrew tap formula
- Scoop bucket manifest
- both should reference GitHub Release assets and checksums

Status: templates added under `packaging/`; not published to an external tap or
bucket yet.

### F5. winget

- winget manifest after release cadence stabilizes

Status: submission notes added under `packaging/winget/`; manifest generation is
deferred until release cadence and package identifier are final.

## Commit Strategy

Do not batch unrelated work.

Preferred commit shape:

- one resource area
- one auth improvement
- one doc batch
- one command family

Examples:

- `feat: add task activity reads`
- `feat: wrap official mcp habit writes`
- `feat: verify openapi focus reads`
- `docs: add official validation matrix`

## Verification Checklist Per Change

Before commit:

1. `go test ./...`
2. `go vet ./...`
3. `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`

Before claiming an API surface is done:

1. request shape verified
2. command added
3. schema updated
4. docs updated
5. live test run
6. rollback or failure behavior understood

## Recommended Execution Order

If another agent takes over, the best sequence is:

1. Finish OpenAPI live OAuth verification
2. Live-smoke remaining known-id official MCP reads and safe write dry-run surfaces
3. Close Web API read gaps
4. Map Web API write gaps with evidence
5. Polish docs and command ergonomics

## Current Best Next Tasks

Top five next tasks:

1. Complete browser authorization for `openapi login` after the developer app redirect URL is configured, then live-smoke OpenAPI resources
2. Create or identify disposable official MCP habit/focus targets for known-id reads
3. Capture a successful Web API task activity request with a Pro account or trace
4. Capture a reversible task-level attachment association trace, then decide whether to expose task attachment upload/download commands
5. Live-smoke official OpenAPI project, task, focus, and habit wrappers after OAuth is available

## Done Means Done

A channel is not "done" because:

- a doc exists
- a schema exists
- a bundle mentions an endpoint
- a command compiles

A channel is only done when:

- auth works
- command works
- live test works
- docs explain it
- another agent can continue without rediscovery
