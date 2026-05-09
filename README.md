<h1 align="center">DidaCLI</h1>

<p align="center">
  <b>A clean, agent-friendly CLI for Dida365 / TickTick.</b>
</p>

<p align="center">
  <a href="https://github.com/DeliciousBuding/dida-cli/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/DeliciousBuding/dida-cli/ci.yml?branch=main&label=ci"></a>
  <a href="https://github.com/DeliciousBuding/dida-cli/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/DeliciousBuding/dida-cli"></a>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white">
  <img alt="JSON" src="https://img.shields.io/badge/JSON-agent--ready-2ea44f">
</p>

<p align="center">
  <a href="README.zh-CN.md">中文 README</a> ·
  <a href="https://deliciousbuding.github.io/dida-cli/">Website</a> ·
  <a href="#quickstart">Quickstart</a> ·
  <a href="#commands">Commands</a> ·
  <a href="#agent-workflows">Agent Workflows</a> ·
  <a href="docs/quickstart.md">Docs</a>
</p>

---

## Why DidaCLI

DidaCLI is a clean-room Go CLI for Dida365 / TickTick task operations. It is designed for humans, shell scripts, Hermes, Codex, and other agents that need a stable command surface, predictable JSON, and safe task automation.

The primary integration surface is the Dida365 Web API used by the official web app. That gives broader control than the public Open API while keeping the tool explicit, inspectable, and easy to disable.

| For operators | For agents | For developers |
|---|---|---|
| Browser cookie login, readable commands, no token printing | Stable JSON envelopes, bounded reads, dry-run support | Small Go codebase, unit tests, CI, documented Web API assumptions |

## Features

- Web API first: sync, settings, projects, folders, tags, filters, columns, comments, tasks, completed history, closed history, trash, search, Pomodoro, habits, sharing metadata, calendar metadata, statistics, templates, and raw GET probes.
- Agent-safe JSON: every `--json` response uses a consistent envelope.
- Ergonomic writes: create, update, complete, move, and parent operations run directly; destructive actions still require explicit confirmation.
- Browser login: visible Dida365 login captures only the `t` cookie into `~/.dida-cli/`.
- Dual-channel direction: Web API for broad coverage, official MCP for a cleaner token-based integration surface.
- Three-channel direction: Web API for breadth, official MCP for token-based tool access, and official OpenAPI for OAuth-based REST integration.
- Safety guardrails: cookie arguments disabled by default, `--dry-run` previews, bounded list output, and no raw write tunnel.

## Quickstart

### One-Line Install

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

Both installers download the latest GitHub Release, verify `checksums.txt`,
install to the user-local bin directory, then run:

```bash
dida version
dida doctor --json
```

Optional installer environment variables:

| Variable | Purpose |
| --- | --- |
| `DIDA_VERSION` | Install a specific release tag, for example `v0.1.0`. |
| `DIDA_INSTALL_DIR` | Override the install directory. |
| `DIDA_REPO` | Override the GitHub repository, for example for forks. |

### Login And First Read

```bash
dida auth login --browser --json
dida doctor --json
dida agent context --json
dida schema list --json
```

### Official Channels

```bash
DIDA365_TOKEN=... dida official doctor --json
dida openapi doctor --json
```

Web API, official MCP, and official OpenAPI are separate auth channels. Do not
reuse cookies or tokens across channels.

### Agent Note

This section is optimized for LLM/Agent operators. Prefer JSON commands,
inspect `dida schema list --json` before writes, preview generated writes with
`--dry-run`, and never ask the user to paste cookies or tokens into chat.

### Install From Source

Use this path for development:

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git
cd dida-cli
go test ./...
go build -o bin/dida ./cmd/dida
```

Full guides:

- [Quickstart](docs/quickstart.md)
- [LLM / Agent Quickstart](docs/llm-quickstart.md)
- [Distribution](docs/distribution.md)

## Commands

### Read Commands

```bash
dida doctor --json
dida official doctor --json
dida official tools --limit 20 --json
dida official show list_projects --json
dida official call list_projects --json
dida official call list_undone_tasks_by_time_query --args-json "{\"query_command\":\"today\"}" --json
dida openapi doctor --json
dida openapi status --json
dida openapi login --json
dida openapi auth-url --json
dida openapi exchange-code --code <code> --json
dida openapi project list --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi habit list --json
dida schema list --json
dida agent context --json
dida auth status --verify --json
dida sync all --json
dida sync checkpoint <checkpoint> --json
dida settings get --json
dida settings get --include-web --json
dida project list --json
dida folder list --json
dida tag list --json
dida filter list --json
dida task today --json
dida task upcoming --days 14 --json
dida completed today --json
dida closed list --status 2 --from 2026-05-01 --to 2026-05-09 --json
dida trash list --cursor 20 --compact --json
dida search all --query "exam" --limit 20 --json
dida pomo stats --json
dida habit checkins --habit <habit-id> --json
dida stats general --json
dida template project list --limit 20 --json
dida user sessions --limit 10 --json
```

### Write Commands

```bash
dida task create --project <project-id> --title "Read paper" --dry-run --json
dida task create --project <project-id> --title "Read paper" --priority 3 --json
dida task update <task-id> --project <project-id> --title "Read paper carefully" --json
dida task complete <task-id> --project <project-id> --json
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
dida project create --name "New project" --dry-run --json
dida folder create --name "Work" --dry-run --json
dida tag create planning --dry-run --json
```

See [docs/commands.md](docs/commands.md) for the full command reference.

## Agent Workflows

Recommended first-pass context:

```bash
dida doctor --json
dida schema list --json
dida agent context --json
dida auth status --verify --json
```

Recommended write flow:

```bash
dida task create --project <project-id> --title "Agent-created task" --dry-run --json
dida task create --project <project-id> --title "Agent-created task" --json
```

Repo-local skill:

```text
skills/dida-cli/SKILL.md
```

Install notes for Codex, Claude Code, OpenClaw, and Hermes Agent are in [docs/skill-installation.md](docs/skill-installation.md).

## Web API Scope

DidaCLI currently covers a broad slice of the observed Dida365 Web API, including:

- sync: `/batch/check/...`
- settings: `/user/preferences/settings`
- tasks/projects/folders/tags/comments
- productivity: `/pomodoros...`, `/habit...`
- sharing/calendar/statistics/templates/search
- raw read-only probing

See [docs/web-api.md](docs/web-api.md), [docs/api-coverage.md](docs/api-coverage.md), and [docs/research/api-surfaces.md](docs/research/api-surfaces.md) for endpoint-level notes.

For channel comparison and future direction, see [docs/research/official-mcp-vs-webapi.md](docs/research/official-mcp-vs-webapi.md).
For the official OpenAPI OAuth channel, see [docs/research/official-openapi-guide.md](docs/research/official-openapi-guide.md).

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

CI runs `go test ./...`, `go vet ./...`, and `govulncheck` on push.

## License

MIT. See [LICENSE](LICENSE).
