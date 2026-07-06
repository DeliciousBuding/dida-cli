<p align="center">
  <img src="assets/logo-icon.svg" alt="DidaCLI" width="100">
</p>

<h1 align="center">DidaCLI</h1>

<p align="center">
  <b>JSON-first CLI for <a href="https://dida365.com">Dida365</a> / <a href="https://ticktick.com">TickTick</a></b>
</p>

<p align="center">
  <a href="https://github.com/DeliciousBuding/dida-cli/blob/main/LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-blue?style=flat-square"></a>
  <img alt="Go 1.26+" src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white">
  <a href="https://github.com/DeliciousBuding/dida-cli/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/DeliciousBuding/dida-cli?style=flat-square&logo=github"></a>
  <a href="https://github.com/DeliciousBuding/dida-cli/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/DeliciousBuding/dida-cli/ci.yml?branch=main&label=ci&style=flat-square&logo=github-actions"></a>
  <a href="https://www.npmjs.com/package/@delicious233/dida-cli"><img alt="npm" src="https://img.shields.io/npm/v/@delicious233/dida-cli?style=flat-square&logo=npm"></a>
</p>

<p align="center">
  <a href="README.zh-CN.md">中文</a> ·
  <a href="https://deliciousbuding.github.io/dida-cli/">Website</a> ·
  <a href="docs/commands.md">Commands</a>
</p>

---

DidaCLI builds as a single Go binary with no external Go modules. It keeps Web API cookie auth, Official MCP tokens, and OpenAPI OAuth separate. Commands return a consistent JSON envelope.

```bash
$ dida task today --compact --json
```

## Install

```bash
npm i -g @delicious233/dida-cli          # npm
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

<details>
<summary>All install options</summary>

### npm (recommended)

```bash
npm install -g @delicious233/dida-cli
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

### Go

```bash
go install github.com/DeliciousBuding/dida-cli/cmd/dida@latest
```

### Pin a specific version

```bash
DIDA_VERSION=vX.Y.Z curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

After install:

```bash
dida version && dida doctor --json
dida upgrade --check
```

</details>

## Quick Start

```bash
# 1. Login with the Dida365 browser cookie named "t"
dida auth cookie set --token-stdin --json

# Optional browser capture path
dida auth login --browser --json

# 2. Verify
dida doctor --verify --json

# 3. See today
dida +today --json

# 4. Create a task (preview first)
dida task create --project <id> --title "Ship v1" --dry-run --json
```

## Feature Coverage

DidaCLI exposes 140 local command contracts through `dida schema list --compact --json`. The main feature groups are:

| Area | What it covers | Entry points |
|---|---|---|
| Output and safety | Stable JSON envelope, compact task output, local schema discovery, write previews, destructive confirmations, token redaction | `--json`, `--compact`, `--dry-run`, `--yes`, `schema` |
| Auth and diagnostics | Browser cookie login, stdin cookie import, auth verification, logout, local config and endpoint checks | `auth`, `doctor` |
| Channel selection | Separate Web API, Official MCP, and Official OpenAPI auth models with local guidance for agents | `channel list` |
| Agent context | One-call context pack with projects, folders, tags, filters, today, upcoming, quadrant buckets, and outline mode | `agent context` |
| Sync | Full sync and checkpoint-based incremental sync with normalized views and raw-compatible deltas | `sync all`, `sync checkpoint` |
| Tasks | Today, latest captured tasks, active lists, task detail, search, upcoming, due counts, create, update, complete, delete, move, parent/subtask assignment | `+today`, `task` |
| Task fields | Content, rich description, start and due dates, timezone, priority, tags, checklist items, column, reminders, repeat metadata, all-day, floating | `task create`, `task update` |
| Projects and organization | Project list/tasks/CRUD, folder CRUD, tag CRUD/rename/merge, filters, Kanban column list and experimental create | `project`, `folder`, `tag`, `filter`, `column` |
| Comments and files | Comment list/create/update/delete, comment attachment upload, attachment quota, existing task attachment download | `comment`, `attachment` |
| History | Completed tasks, closed-history items, and deleted tasks from trash | `completed`, `closed`, `trash` |
| Productivity | Pomodoro preferences, records, timing, stats, timeline, task records, habit preferences, habits, sections, check-ins, Eisenhower quadrants | `pomo`, `habit`, `quadrant` |
| Account metadata | Settings, daily reminders, share contacts, project sharing state, calendar subscriptions/accounts, account stats, project templates, global search, user profile/status/sessions | `settings`, `reminder`, `share`, `calendar`, `stats`, `template`, `search`, `user` |
| Official MCP | Token management, tool discovery, schema display, raw tool call, project reads, task search/filter/batch writes, habit reads/writes, focus reads/delete | `official` |
| Official OpenAPI | OAuth client config, browser/manual OAuth, project/task/focus/habit wrappers, completed and filtered task reads | `openapi` |
| Raw read probes | Read-only GET escape hatch for verified Web API exploration with structured error details | `raw get` |
| Distribution and upgrade | Single binary builds, shell completion scripts, npm/install scripts, packaging templates, release archives, checksum-verified self-upgrade | `completion`, `upgrade` |

## Command Examples

<details>
<summary>Web API reads</summary>

```bash
dida task today --json
dida task latest --limit 10 --project inbox --compact --json
dida task list --filter all --limit 50 --compact --json
dida task upcoming --days 14 --json
dida task search --query "exam" --json
dida task due-counts --json
dida project list --json
dida project tasks <project-id> --limit 50 --compact --json
dida folder list --json
dida tag list --json
dida filter list --json
dida column list <project-id> --json
dida completed today --json
dida closed list --status 2 --limit 50 --json
dida trash list --cursor 20 --compact --json
dida settings get --include-web --json
dida reminder daily --json
dida attachment quota --json
dida share contacts --json
dida share project shares <project-id> --json
dida calendar subscriptions --json
dida stats general --json
dida template project list --limit 50 --json
dida search all --query "exam" --limit 20 --json
dida user profile --json
dida pomo stats --json
dida habit list --json
dida quadrant list --json
dida sync all --json
```
</details>

### Latest captured tasks

When several related tasks are captured from WeChat or another inbox flow with
their context and materials, use `task latest` to read the newest captured
items first:

```bash
dida task latest --limit 10 --project inbox --compact --json
```

The command sorts active tasks by creation time, maps `--project inbox` to the
real inbox project from sync data, and falls back to modified time when creation
metadata is missing. Omit `--project inbox` to read the newest active tasks
across all projects.

<details>
<summary>Web API writes</summary>

```bash
dida task create --project <id> --title "New task" --dry-run --json
dida task update <id> --project <p> --title "Updated" --dry-run --json
dida task complete <id> --project <p> --dry-run --json
dida task move <id> --from <p> --to <p> --dry-run --json
dida task parent <id> --parent <parent-id> --project <p> --dry-run --json
dida task delete <id> --project <p> --dry-run --yes --json
dida project create --name "New project" --dry-run --json
dida folder create --name "Work" --dry-run --json
dida tag rename old-name new-name --dry-run --json
dida column create --project <p> --name "Doing" --dry-run --json
dida comment create --project <p> --task <id> --text "Looks good" --dry-run --json
dida comment create --project <p> --task <id> --text "See file" --file ./probe.png --dry-run --json
dida attachment download --project <p> --task <id> --attachment <a> --output ./file.doc --json
```
</details>

<details>
<summary>Official MCP and OpenAPI</summary>

```bash
# MCP (token)
DIDA365_TOKEN=dp_xxx dida official project list --json
dida official tools --limit 20 --json
dida official task filter --project <id> --status 0 --json
dida official task batch-add --args-json '{"tasks":[{"title":"Task"}]}' --dry-run --json
dida official habit list --json
dida official focus list --from-time 2026-05-01T00:00:00+08:00 --to-time 2026-05-09T23:59:59+08:00 --type 1 --json

# OpenAPI (OAuth)
dida openapi client set --id <id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
dida openapi task create --args-json '{"projectId":"<project-id>","title":"Task"}' --dry-run --json
dida openapi task filter --args-json '{"projectIds":["<project-id>"],"status":[0]}' --json
dida openapi habit checkins --habit-ids <habit-id> --from 20260401 --to 20260407 --json
```
</details>

<details>
<summary>Agent, schema, raw reads, and upgrade</summary>

```bash
dida schema list --compact --json
dida schema show task.create --json
dida completion bash
dida channel list --json
dida agent context --outline --json
dida raw get /user/preferences/settings --json
dida upgrade --check --json
```
</details>

Full reference: [docs/commands.md](docs/commands.md). Machine-readable reference: `dida schema list --compact --json`.

## Auth Channels

| | Web API | Official MCP | Official OpenAPI |
|---|---|---|---|
| **Auth** | Browser cookie | Token | OAuth |
| **Coverage** | Web API resources outside official channels | MCP tool-based | Standard REST |
| **Setup** | One login | Get token | Register app |

The three auth channels stay separate.

## Agent Integration

```bash
dida schema list --compact --json        # discover commands
dida agent context --outline --json      # build context
dida task create ... --dry-run --json    # preview writes
```

| Agent | Install |
|---|---|
| Claude Code | Copy [`skills/dida-cli/SKILL.md`](skills/dida-cli/SKILL.md) |
| Codex / Others | See [docs/skill-installation.md](docs/skill-installation.md) |

Preview resource writes with `--dry-run` when supported. Destructive commands require `--yes`. The CLI does not print full token values. See [Agent Usage](docs/agent-usage.md).

## Docs

- [Commands Reference](docs/commands.md)
- [Agent Usage](docs/agent-usage.md)
- [API Coverage](docs/api-coverage.md)
- [OpenAPI Setup](docs/openapi-setup.md)

## Contributing

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git && cd dida-cli
go test ./... && go build -o bin/dida ./cmd/dida
```

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security and Conduct

Report vulnerabilities through the private advisory process in [SECURITY.md](SECURITY.md). Do not post live cookies, tokens, private task exports, or full API response dumps in public issues.

Project discussion follows [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE)

---

DidaCLI is an independent open-source CLI for [Dida365](https://dida365.com) / [TickTick](https://ticktick.com)-compatible workflows. Use it with accounts and automations you control, subject to the upstream services' terms.
