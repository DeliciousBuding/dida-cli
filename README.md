<h1 align="center">DidaCLI</h1>

<p align="center">
  <b>A clean, agent-friendly CLI for Dida365 / TickTick.</b><br/>
  <b>面向 Agent 和自动化工作流的滴答清单命令行工具。</b>
</p>

<p align="center">
  <a href="https://github.com/DeliciousBuding/dida-cli/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/DeliciousBuding/dida-cli/ci.yml?branch=main&label=ci"></a>
  <a href="https://github.com/DeliciousBuding/dida-cli/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/DeliciousBuding/dida-cli"></a>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white">
  <img alt="JSON" src="https://img.shields.io/badge/JSON-agent--ready-2ea44f">
</p>

<p align="center">
  <a href="https://deliciousbuding.github.io/dida-cli/">Website</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#commands">Commands</a> ·
  <a href="#agent-workflows">Agent Workflows</a> ·
  <a href="#中文说明">中文说明</a> ·
  <a href="docs/commands.md">Docs</a>
</p>

---

## Why DidaCLI

DidaCLI is a clean-room Go CLI for Dida365 / TickTick task operations. It is designed for humans, shell scripts, Hermes, Codex, and other agents that need a stable command surface, predictable JSON, and safe task automation.

The primary integration surface is the Dida365 Web API used by the official web app. That gives broader control than the public Open API while keeping the tool explicit, inspectable, and easy to disable.

| For operators | For agents | For developers |
|---|---|---|
| Browser cookie login, readable commands, no token printing | Stable JSON envelopes, bounded list commands, dry-run support | Small Go codebase, unit tests, CI, documented Web API assumptions |
| 日常操作命令短、输出清晰 | 错误结构化，便于自动恢复 | 结构简单，方便扩展和审计 |

## Features

- Web API first: sync, settings, projects, folders, tags, filters, columns, comments, tasks, completed history, and raw GET probes.
- Full task CRUD plus comment CRUD, project/folder/tag CRUD, task move/subtask operations, and conservative kanban column support.
- Agent-safe JSON: every `--json` response uses a consistent envelope.
- Ergonomic writes: create/update/complete run directly; `--dry-run` previews; destructive delete requires `--yes`.
- Browser login: visible Dida365 login captures only the `t` cookie into `~/.dida-cli/`.
- No secrets in output: tokens are redacted and never committed.
- Raw escape hatch: `dida raw get /path --json` for read-only API exploration.

## Quick Start

### Install From Source

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git
cd dida-cli
go test ./...
go build -o bin/dida ./cmd/dida
```

Install locally on Unix-like systems:

```bash
make install-local
dida doctor --json
```

On Windows PowerShell:

```powershell
go build -o bin\dida.exe .\cmd\dida
Copy-Item .\bin\dida.exe $env:USERPROFILE\.local\bin\dida.exe -Force
dida doctor --json
```

### Login

Recommended:

```bash
dida auth login --browser --json
dida auth status --verify --json
```

Fallback:

```bash
dida auth login --json
# Complete Dida365 / WeChat / QR login in your browser.
# Import only the cookie named "t" through stdin. Do not paste cookies into chat.
dida auth cookie set --token-stdin
dida auth status --verify --json
```

## Commands

### Read Commands

```bash
dida doctor --json
dida schema list --json
dida schema show task.create --json
dida agent context --json
dida auth status --verify --json
dida sync all --json
dida sync checkpoint <checkpoint> --json
dida settings get --json
dida project list --json
dida project tasks <project-id> --compact --json
dida project columns <project-id> --json
dida folder list --json
dida tag list --json
dida filter list --json
dida column list <project-id> --json
dida comment list --project <project-id> --task <task-id> --json
dida task today --json
dida task list --filter all --limit 50 --compact --json
dida task search --query "exam" --limit 10 --json
dida task upcoming --days 14 --json
dida quadrant list --json
dida completed today --json
dida completed list --from 2026-05-01 --to 2026-05-09 --compact --json
dida raw get /batch/check/0 --json
```

### Write Commands

```bash
# Preview a create request.
dida task create --project <project-id> --title "Read paper" --dry-run --json

# Create directly.
dida task create --project <project-id> --title "Read paper" --priority 3 --json

# Update directly.
dida task update <task-id> --project <project-id> --title "Read paper carefully" --json

# Comment on a task.
dida comment create --project <project-id> --task <task-id> --text "Looks good" --dry-run --json

# Complete directly.
dida task complete <task-id> --project <project-id> --json

# Delete is destructive and requires --yes.
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json

# Resource CRUD.
dida project create --name "New project" --dry-run --json
dida folder create --name "Work" --dry-run --json
dida tag create planning --dry-run --json
dida task move <task-id> --from <project-id> --to <project-id> --json
```

See [docs/commands.md](docs/commands.md) for the full command reference.

## JSON Contract

Success:

```json
{
  "ok": true,
  "command": "task today",
  "meta": {
    "count": 1,
    "total": 1
  },
  "data": {
    "tasks": []
  }
}
```

Error:

```json
{
  "ok": false,
  "command": "task delete",
  "error": {
    "type": "confirmation_required",
    "message": "task delete requires --yes",
    "hint": "preview first with: dida task delete <task-id> --project <project-id> --dry-run"
  }
}
```

## Agent Workflows

Recommended read-only context pack:

```bash
dida doctor --json
dida schema list --json
dida agent context --json
dida auth status --verify --json
dida project list --json
dida folder list --json
dida tag list --json
dida +today --json
dida task upcoming --days 14 --limit 50 --json
dida quadrant list --json
dida completed today --json
```

Safe write flow:

```bash
dida task create --project <project-id> --title "Agent-created task" --dry-run --json
dida task create --project <project-id> --title "Agent-created task" --json
dida project create --name "Agent staging" --dry-run --json
```

Repo-local skill:

```text
skills/dida-cli/SKILL.md
```

Install notes for Codex, Claude Code, OpenClaw, and Hermes Agent are in [docs/skill-installation.md](docs/skill-installation.md).

Agent rules:

- Do not ask users to paste cookies into chat.
- Prefer `dida agent context --json` for the first task-management context pack.
- Use `--json` for automation.
- Use `--dry-run` before broad or generated writes.
- Use `--yes` only for explicit destructive actions.
- Prefer high-level commands before `raw get`.

## Web API Scope

DidaCLI currently uses:

- `GET /batch/check/0`
- `GET /batch/check/{checkpoint}`
- `GET /user/preferences/settings`
- `GET /project/{projectId}/tasks`
- `GET /project/all/completed?...`
- `POST /batch/task`
- `POST /batch/taskProject`
- `POST /batch/taskParent`
- `POST /batch/project`
- `POST /batch/projectGroup`
- `POST /batch/tag`
- `PUT /tag/rename`
- `PUT /tag/merge`
- `DELETE /tag?name=...`
- sync-returned `filters`
- `POST /column`
- `GET /column/project/{projectId}`
- `GET/POST/PUT/DELETE /project/{projectId}/task/{taskId}/comment(s)`

Private Web API behavior can change. See [docs/web-api.md](docs/web-api.md), [docs/api-coverage.md](docs/api-coverage.md), and [docs/research/api-surfaces.md](docs/research/api-surfaces.md) for implementation notes.

`sync checkpoint` exposes both normalized view data and raw-compatible delta sections so agents can see task deletes, order changes, and reminder changes.

For token-sensitive agent reads, add `--compact` or `--brief` to task-heavy commands such as `task list`, `task upcoming`, `task search`, `project tasks`, and `completed list`. Compact output preserves IDs, titles, dates, priority, status, column, and tags while omitting large text/checklist/reminder/raw fields.

## Project Layout

```text
cmd/dida/          CLI entrypoint
internal/auth/     Cookie capture, storage, redaction
internal/cli/      Command dispatch and JSON envelopes
internal/model/    Normalized task/project models
internal/webapi/   Dida365 Web API client
docs/              User and API documentation
skills/dida-cli/   Repo-local agent skill
```

## Development

```bash
go test ./...
go build -o bin/dida ./cmd/dida
```

CI runs `go test ./...` on every push and pull request.

Repository constraints:

- Do not commit cookies, tokens, response dumps, or private fixtures.
- Put private research notes under ignored `data/private/`.
- Keep raw Web API probes read-only unless a command is explicitly designed and documented.
- Add tests for request builders before adding live write behavior.

## 中文说明

DidaCLI 是一个用 Go 编写的滴答清单 / TickTick 命令行工具，目标是替代零散脚本和旧 OpenClaw glue，让 Hermes、Codex 这类 Agent 可以稳定、安全地操作任务。

核心设计：

- Web API 优先，覆盖比官方 Open API 更完整的个人账号操作面。
- 命令输出稳定 JSON，适合 Agent 自动解析。
- 登录只保存本地浏览器 cookie `t`，不会把 token 打印出来。
- 任务、项目、文件夹、标签都提供明确 CRUD 命令；删除类操作必须显式 `--yes`。
- 所有写操作支持 `--dry-run` 预览。
- 仓库内置 `skills/dida-cli/SKILL.md`，可给 Claude Code、Codex、OpenClaw、Hermes Agent 使用。
- 研究和私密材料不进入仓库，统一放到被忽略的 `data/private/`。

常用命令：

```bash
dida auth login --browser --json
dida auth status --verify --json
dida project list --json
dida folder list --json
dida tag list --json
dida +today --json
dida task upcoming --days 14 --json
dida quadrant list --json
dida completed today --json
dida task create --project <project-id> --title "新任务" --json
dida project create --name "新项目" --dry-run --json
```

## License

MIT. See [LICENSE](LICENSE).
