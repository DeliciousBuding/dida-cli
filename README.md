# dida-cli

Clean-room Go CLI for Dida365 / TickTick task operations.

This project is intended to replace the old `dida365-ai-tools` + OpenClaw glue with a maintainable CLI that Hermes and other agents can call safely.

## Naming

- Repository: `DidaCLI`
- Future repository/module name: `dida-cli`
- Binary: `dida`
- Config directory: `~/.dida-cli`

## Design Boundary

The CLI is Web API first. It will support two API surfaces:

- Web private API: primary path, cookie-auth, broad account control for the operator's own account.
- Official Open API: secondary compatibility path for stable supported operations.

Private API support must stay explicit and easy to disable. It should not bypass access controls, scrape other users, or hide risky write operations.

## Architecture Target

The old Doris setup mixed API access, report rendering, cron, and OpenClaw delivery. This repo separates those layers:

- `internal/auth`: stores and redacts cookie/OAuth credentials.
- `internal/webapi`: typed client for `https://api.dida365.com/api/v2`.
- `internal/openapi`: typed client for official OAuth APIs.
- `internal/model`: stable internal task/project/tag types.
- `internal/report`: deterministic reports from normalized models.
- `internal/cli`: command surface and JSON envelopes.

The CLI design follows the useful parts of `larksuite/cli`:

- three command layers: shortcuts, resource/API commands, and raw API;
- machine-readable health checks and auth status;
- dry-run previews for writes;
- structured error envelopes that agents can act on;
- companion agent skill docs after the CLI is usable.

Web API operations are grouped by real Dida resources:

- sync/settings
- tasks
- projects/folders/columns
- tags
- completed/history
- batch operations
- raw read-only endpoint probes

Write commands must support `--dry-run` before live writes.

## Command Layers

### 1. Shortcuts

Human and agent friendly workflows with strong defaults:

```text
dida +today [--json]
dida +nightly-report [--json]
dida +inbox-zero [--dry-run]
```

### 2. Resource Commands

Stable commands mapped to Dida resources:

```text
dida task list --filter today --json
dida task create --project <id> --title <title> --dry-run
dida project list --json
dida tag list --json
```

### 3. Raw Web API

Read-first escape hatch for reverse-engineering and coverage gaps:

```text
dida raw get /batch/check/0 --json
```

Non-GET raw requests should stay disabled until a specific safe workflow needs them.

## Safety Model

- Read commands can run directly.
- Write commands support `--dry-run`.
- Destructive writes should require explicit `--yes` later.
- Credentials are never printed.
- Error JSON should include actionable `hint` fields where possible.

## Planned Command Surface

```text
dida doctor [--json]
dida auth login [--json]
dida auth status [--json]
dida auth status --verify [--json]
dida auth cookie set
dida auth logout
dida auth oauth start
dida sync all [--json]
dida project list [--json]
dida project columns <project-id> [--json]
dida tag list [--json]
dida completed list [--from YYYY-MM-DD --to YYYY-MM-DD] [--json]
dida task today [--json]
dida task create --project <id> --title <title> [--dry-run]
dida task update <task-id> --project <id> [--title ...] [--due ...] [--dry-run]
dida task complete <task-id> --project <id> [--dry-run]
dida task move <task-id> --from-project <id> --to-project <id> [--dry-run]
dida batch apply <file.json> [--dry-run]
dida report nightly [--json]
dida raw get <path> [--json]
```

## JSON Policy

With `--json`, commands return stable JSON objects. Errors must also be JSON and must never contain full cookies, bearer tokens, or request headers.

Initial success envelope:

```json
{
  "ok": true,
  "command": "doctor",
  "data": {}
}
```

Initial error envelope:

```json
{
  "ok": false,
  "command": "doctor",
  "error": {
    "message": "missing auth"
  }
}
```

## Current Status

Implemented:

- `dida doctor [--json]`
- `dida auth login [--json]`
- `dida auth status [--verify] [--json]`
- `dida auth cookie set --token-stdin`
- `dida auth logout`
- `dida sync all [--json]`
- `dida project list [--json]`
- `dida task list --filter today|all [--limit N] [--json]`
- `dida task today [--limit N] [--json]`
- `dida +today [--limit N] [--json]`
- `dida raw get <path> [--json]`

Credentials are stored only under `~/.dida-cli/`, not in this repository.

## Agent Workflow

Recommended read-only flow:

```bash
dida doctor --json
dida auth status --verify --json
dida project list --json
dida +today --json
dida task list --filter all --limit 50 --json
```

Login flow:

```bash
dida auth login --json
# User completes browser / WeChat / QR login and copies only cookie "t".
dida auth cookie set --token-stdin
dida auth status --verify --json
```

Agents must not ask the user to paste cookies into chat. Tokens should go to stdin or a local secret manager only.

## Build

```bash
go test ./...
go build -o bin/dida ./cmd/dida
```
