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

## Current Baseline (as of v0.2.5)

Latest release: `v0.2.5` (2026-07-07).

### Three Channels — All Functional

| Channel | Auth | Status |
|---|---|---|
| Web API | Browser cookie `t` | Primary, widest coverage |
| Official MCP | `DIDA365_TOKEN=dp_...` | 22 first-class wrappers + generic `call` |
| Official OpenAPI | OAuth access token | Full CRUD for project/task/focus/habit |

### Distribution

- GitHub Releases: 6-platform binary archives + checksums.txt + package-manager repo export artifact
- `install.sh` / `install.ps1`: smoke-tested on Linux/macOS/Windows
- npm: `@delicious233/dida-cli` with postinstall binary download
- `dida upgrade`: self-update with SHA-256 verification (new in v0.2.1)
- CI: tag-triggered release, provenance, release gates, and npm auto-publish

### Engineering Quality

- Test coverage: webapi 84%, officialmcp 85%, openapi 83%, model 91%, config 83%
- CLI package coverage: 61.3% after local command coverage tests
- All HTTP clients have explicit timeouts (30-60s) and response size limits
- Error messages redact tokens and sensitive patterns
- Upgrade enforces checksum verification (fails if checksums.txt missing)

### What's NOT Done Yet

- goreleaser migration is deferred through `v0.3.x`; the current hand-written release workflow remains the release path
- No Homebrew tap or Scoop bucket (templates are generated and validated, not published)
- Remaining OpenAPI and Official MCP live smokes need suitable account state
- No i18n (all errors English-only)
- Website first pass is complete; deeper documentation content can still improve

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
- 2026-05-10 repeat smoke verified the comment attachment path with an
  available quota, a 1x1 PNG, read-back through `comment list`, comment delete,
  and disposable task cleanup
- the CLI checks attachment quota before upload; future live upload smokes need
  available attachment quota and should use a known-supported file type such as
  PNG unless testing the file matrix explicitly
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
9. `filter_tasks` - implemented as `official task filter`
10. `list_undone_tasks_by_date` - implemented as `official task undone`
11. `search_task` - implemented as `official task search`
12. `list_undone_tasks_by_time_query` - implemented as `official task query`
13. `list_habits` - implemented as `official habit list`
14. `list_habit_sections` - implemented as `official habit sections`
15. `get_habit` - implemented as `official habit get`
16. `create_habit` - implemented as `official habit create`
17. `update_habit` - implemented as `official habit update`
18. `upsert_habit_checkins` - implemented as `official habit checkin`
19. `get_habit_checkins` - implemented as `official habit checkins`
20. `get_focus` - implemented as `official focus get`
21. `get_focuses_by_time` - implemented as `official focus list`
22. `delete_focus` - implemented as `official focus delete`

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
- Local dry-run previews exist for official MCP task batch writes,
  `official habit create/update/checkin`, and `official focus delete`; these
  do not require a saved official token.
- 2026-05-10 live reversible smokes verified official MCP task `batch-add`,
  `batch-update`, project-scoped `task get`, and `complete-project`; temporary
  tasks were then removed through the already verified Web API task delete path.
- Known-id habit/focus read smokes are currently blocked by account state:
  2026-05-10 token smokes found zero habits and zero focus records, including a
  365-day focus range.
- Habit writes and focus delete remain blocked until disposable habit/focus
  targets exist.

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
- 2026-05-10: end-to-end OpenAPI OAuth succeeded on the current account via
  `listen-callback`, `exchange-code`, `status`, and `project list`; the
  remaining OpenAPI work is now resource-family live coverage and disposable
  write smoke.

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

Status: implemented and published through `v0.2.5`; release `v0.2.5` includes
six platform archives, `checksums.txt`, npm provenance, archive attestations,
and a package-manager repo export artifact handoff for the next tag release.
Workflow-dispatch trigger exists for manual re-triggering.

### F2. Install Scripts

- `install.sh`
- `install.ps1`
- OS/arch detection
- latest release download by default
- `DIDA_VERSION`, `DIDA_INSTALL_DIR`, `DIDA_REPO`
- checksum verification
- install-time `dida version` and `dida doctor --json`

Status: implemented; Windows `install.ps1` latest smoke and WSL Linux
`install.sh` latest smoke are covered by the current release line. The
installed-binary OpenAPI client config smoke passed during release hardening.
Native macOS installer smoke remains pending.

### F3. npm Installer

- package: `@delicious233/dida-cli`
- postinstall downloads matching GitHub Release binary
- `bin/dida` forwards to the downloaded binary
- npm auto-publish on tag push (release workflow)

Status: published as `@delicious233/dida-cli@0.2.5` with npm README metadata.
npm auto-publish on tag push is in the release workflow. Postinstall binary
download is covered by Windows and Linux npm install smoke jobs.

### F4. Homebrew / Scoop

- Homebrew tap formula
- Scoop bucket manifest
- both should reference GitHub Release assets and checksums

Status: templates exist under `packaging/` and can be regenerated from release
`checksums.txt` with `scripts/update-packaging-templates.sh`. Repo-ready
Homebrew tap and Scoop bucket layouts can be exported locally with
`scripts/export-package-manager-repos.sh`, and the release workflow uploads the
same layouts as `dida-package-manager-repos-vX.Y.Z` after each GitHub Release.
They are not published to external repositories yet. Homebrew tap and Scoop
bucket publication remains planned for v0.3.0.

### F5. winget

- winget manifest after release cadence stabilizes

Status: submission notes added under `packaging/winget/`; manifest generation is
deferred until release cadence and package identifier are final.

## Workstream G: Self-Update & CLI Ergonomics

### G1. v0.2.x Completed Scope

| Item | Status | Notes |
|---|---|---|
| `dida upgrade` command | Done | SHA-256 verified, Windows rename-replace |
| Checksum enforcement | Done | Fails if checksums.txt missing |
| HTTP timeout hardening | Done | All clients 30-60s, response limits |
| Download progress output | Done | Percentage counter on stderr, suppressed in --json |
| Upgrade integration test | Done | httptest mock of full flow + failure paths |
| Schema registry entry | Done | `upgrade` registered |
| CHANGELOG update | Done | Consolidated unreleased items |
| README badges | Done | CI + version + npm badges already present |
| README rewrite | Done | Condensed, better structure, collapsed verbose sections |
| `dida completion` | Done | bash/zsh/fish/powershell, hardcoded templates |
| `dida doctor` upgrade check | Done | Explicit `--check-upgrade`; JSON status plus one-line text output |
| Staticcheck gate | Done | `make staticcheck`, CI, release validation, and release-check use Staticcheck v0.7.0 |
| CLI coverage floor | Done | `internal/cli` coverage is 61.3%; keep new work above 60% |

### G2. v0.3.0 Scope (next milestone)

| Item | Priority | Notes |
|---|---|---|
| goreleaser migration | Deferred | Keep the current release workflow through `v0.3.x`; decision record is `docs/research/release-strategy-goreleaser.md` |
| Package-manager template generator | Done | `scripts/update-packaging-templates.sh` regenerates Homebrew, Scoop, packaging README, and winget notes from `checksums.txt`; release-check tests it |
| Package-manager repo export | Done | `scripts/export-package-manager-repos.sh` prepares repo roots for `DeliciousBuding/homebrew-dida` and `DeliciousBuding/scoop-bucket`; release-check tests it |
| Release package-manager artifact | Done | `.github/workflows/release.yml` uploads `dida-package-manager-repos-vX.Y.Z` after release checksums exist |
| Homebrew tap | Medium | Create and publish external repo `homebrew-dida` after native Homebrew smoke |
| Scoop bucket | Medium | Create and publish external repo `scoop-bucket` after native Scoop smoke |
| Website polish | Done | Pages homepage now matches `v0.2.5` install, auth, schema, task latest, completion, and security paths; release-check validates it |
| Live smoke backlog | Medium | OpenAPI/Official MCP reads and disposable writes where account state allows |

### G3. v0.4.0+ (long-term)

| Item | Notes |
|---|---|
| `DIDA_LANG=zh` error messages | Lightweight i18n, no framework, message table |
| `dida watch` (file-system trigger) | Watch a markdown file, sync changes to Dida365 |
| Plugin system | User-defined commands via `~/.dida-cli/plugins/` |
| TUI mode | Interactive task browser (bubbletea or similar) |
| winget submission | After release cadence stabilizes |
| Proxy/mirror support | `DIDA_PROXY` for corporate environments |

## Long-Term Vision

DidaCLI aims to be the definitive command-line and agent interface for
Dida365/TickTick. The end state:

1. **Zero-friction install**: `npm i -g`, `brew install`, `scoop install`,
   `dida upgrade` — any path works in under 30 seconds
2. **Agent-native**: structured JSON output, schema discovery, dry-run
   previews, and safety rails make it the preferred tool for AI agents
3. **Three-channel coverage**: Web API for breadth, Official MCP for
   token-based automation, OpenAPI for OAuth integrations
4. **Self-maintaining**: auto-update, stale issue cleanup, CI badges,
   changelog generation — minimal human maintenance overhead
5. **Community-ready**: clear contribution guide, good test coverage,
   accessible docs in EN + ZH

The project is NOT trying to be:
- A full GUI replacement (use the app for that)
- A sync engine (read/write, not bidirectional sync)
- A general-purpose task manager (it's Dida365-specific)

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

1. Live-smoke remaining OpenAPI read families and disposable writes
2. Live-smoke remaining known-id official MCP reads and safe write dry-run surfaces
3. Close Web API read gaps
4. Map Web API write gaps with evidence
5. Polish docs and command ergonomics

## Current Best Next Tasks

For v0.3.0 (next milestone):

1. Create and publish the external Homebrew tap and Scoop bucket from the exported, validated layouts.
2. Re-evaluate GoReleaser only after archive, checksum, npm provenance, attestation, and package-manager publishing parity are proven.
3. Live-smoke remaining OpenAPI read families and disposable writes.
4. Live-smoke known-id Official MCP habit/focus reads when account state allows.

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
