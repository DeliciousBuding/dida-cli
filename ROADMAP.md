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
- `official mcp` has:
  - `official doctor`
  - `official tools`
  - `official show`
  - `official call`
  - first-class project read wrappers
  - first-class task read/filter wrappers
  - first-class habit wrappers
  - first-class focus wrappers
- `official openapi` has:
  - OAuth foundation
  - auth URL generation
  - callback listener
  - code exchange
  - token persistence
  - `project list`
- docs already include:
  - API coverage matrix
  - Web API notes
  - OpenAPI guide
  - MCP vs Web API comparison
  - channel inventory
  - Web API gap catalog
  - MCP tool crosswalk

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

- Auth: `DIDA365_TOKEN=dp_...`
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
  - target: `GET /api/v1/task/activity/{taskId}`
  - need: real successful request shape, cursor semantics
- Trash pagination
  - target: `GET /project/all/trash/page?...`
  - need: `type` semantics, paging cursor, stable output contract

Acceptance:

- command exists
- schema exists
- live read succeeds
- docs updated

### A2. Attachment Flows

- upload attachment
- attach to comment
- download / reference model

Acceptance:

- multipart flow fully mapped
- reversible live test
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

1. `get_project_by_id`
2. `get_project_with_undone_tasks`
3. `complete_tasks_in_project`
4. `batch_add_tasks`
5. `batch_update_tasks`
6. `filter_tasks`
7. `list_undone_tasks_by_date`
8. `search_task`
9. `get_habit`
10. `create_habit`
11. `update_habit`
12. `upsert_habit_checkins`
13. `get_focus`
14. `get_focuses_by_time`
15. `delete_focus`

Acceptance:

- each promoted command must be better than `official call`
- each promoted command must have a reason documented in the crosswalk

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

- finish `openapi login`
- make it pleasant for human and agent use
- keep auth flow separated from browser cookie auth

Acceptance:

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

1. `openapi project list/get/data`
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

Status: implemented; needs first tag release smoke.

### F2. Install Scripts

- `install.sh`
- `install.ps1`
- OS/arch detection
- latest release download by default
- `DIDA_VERSION`, `DIDA_INSTALL_DIR`, `DIDA_REPO`
- checksum verification
- install-time `dida version` and `dida doctor --json`

Status: implemented; syntax and parser checks required before each release.

### F3. npm Installer

- placeholder package under `npm/`
- package name placeholder: `@vectorcontrol/dida-cli`
- postinstall downloads matching GitHub Release binary
- `bin/dida` forwards to the downloaded binary

Status: skeleton only; do not publish until release assets are proven.

### F4. Homebrew / Scoop

- Homebrew tap formula
- Scoop bucket manifest
- both should reference GitHub Release assets and checksums

Status: planned.

### F5. winget

- winget manifest after release cadence stabilizes

Status: planned.

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
2. Promote high-value official MCP tools
3. Close Web API read gaps
4. Map Web API write gaps with evidence
5. Polish docs and command ergonomics

## Current Best Next Tasks

Top five next tasks:

1. Finish and live-verify `openapi login`
2. Live-smoke official MCP project, habit, and focus wrappers where a safe target exists
3. Create the first tagged GitHub Release and test install scripts from release assets
4. Capture a successful Web API task activity request
5. Decode trash pagination semantics

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
