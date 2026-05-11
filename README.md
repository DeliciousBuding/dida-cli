<p align="center">
  <img src="assets/logo-icon.svg" alt="DidaCLI" width="100">
</p>

<h1 align="center">DidaCLI</h1>

<p align="center">
  <b>Agent-friendly task automation for <a href="https://dida365.com">Dida365</a> / <a href="https://ticktick.com">TickTick</a></b>
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
  <a href="docs/quickstart.md">Quick Start</a> ·
  <a href="docs/commands.md">Commands</a> ·
  <a href="CONTRIBUTING.md">Contributing</a>
</p>

---

DidaCLI is a single Go binary that talks to Dida365/TickTick's **Web API**, **official MCP**, and **official OpenAPI** — designed for both humans and AI agents that need structured, predictable task automation.

```bash
$ dida task today --json
{
  "ok": true,
  "command": "task today",
  "meta": { "count": 3 },
  "data": {
    "tasks": [
      { "title": "写周报", "project": "工作" },
      { "title": "Review PR #42", "project": "工作" },
      { "title": "买菜", "project": "生活" }
    ]
  }
}
```

## Why DidaCLI?

- **Agent-native JSON** — Every response uses a stable envelope `{ ok, command, meta, data, error }`. No HTML parsing, no fragile scraping.
- **Three auth channels** — Web API (browser cookie), Official MCP (token), Official OpenAPI (OAuth). Never mixed.
- **Dry-run writes** — All write commands support `--dry-run` to preview payloads before execution.
- **Zero dependencies** — Single static binary, pure Go stdlib, no CGO.
- **Six platforms** — Windows / Linux / macOS on amd64 + arm64. Apple Silicon native.
- **30+ commands** — Tasks, projects, folders, tags, columns, comments, habits, Pomodoro, trash, search, stats, and more.

## Features

| Category | Capabilities |
|---|---|
| **Tasks** | Create, update, complete, delete, move, batch operations, comments, attachments |
| **Projects & Folders** | List, create, manage hierarchies, columns, sort orders |
| **Tags & Filters** | Create, assign, filter by tag, saved filters |
| **Search & Queries** | Full-text search, upcoming tasks, completed history |
| **Habits & Pomodoro** | Check-in tracking, Pomodoro stats and timer |
| **Agent Context** | Outline mode, schema introspection, context packs |
| **Auth & Security** | Multi-channel auth, `doctor` diagnostics, token management |
| **Dry-run & Schema** | Preview every write, inspect API schemas |

## Supported Platforms

| Platform | Architecture | Archive | Install |
|---|---|---|---|
| **macOS** | Apple Silicon (arm64) | `.tar.gz` | `curl ... \| sh` / Homebrew |
| **macOS** | Intel (amd64) | `.tar.gz` | `curl ... \| sh` / Homebrew |
| **Linux** | x86_64 (amd64) | `.tar.gz` | `curl ... \| sh` |
| **Linux** | ARM64 (arm64) | `.tar.gz` | `curl ... \| sh` |
| **Windows** | x86_64 (amd64) | `.zip` | Scoop / PowerShell |
| **Windows** | ARM64 (arm64) | `.zip` | Scoop / PowerShell |

## Install

### npm (recommended)

```bash
npm install -g @delicious233/dida-cli
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

### Homebrew (macOS / Linux) — coming soon

```bash
brew install dida
```

### Windows (Scoop) — coming soon

```powershell
scoop install dida
```

### Windows (PowerShell)

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

### Go

```bash
go install github.com/DeliciousBuding/dida-cli/cmd/dida@latest
```

<details>
<summary><b>Pin a specific version</b></summary>

```bash
# macOS / Linux
DIDA_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh

# Windows PowerShell
$env:DIDA_VERSION="v0.2.0"; iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```
</details>

After install, verify:

```bash
dida version
dida doctor --json
```

## Quick Start

```bash
# 1. Login — open dida365.com in your browser, sign in, then:
dida auth cookie set --token-stdin --json
# Paste your browser cookie named "t" when prompted.

# 2. Verify everything works
dida doctor --verify --json

# 3. View today's tasks
dida task today --json

# 4. List your projects to get a project ID
dida project list --json

# 5. Create a task (preview first with --dry-run)
dida task create --project <project-id> --title "Ship v1" --dry-run --json
dida task create --project <project-id> --title "Ship v1" --json

# 6. Get context for AI agents
dida agent context --outline --json
```

> **Tip:** `dida auth login --browser --json` can capture the cookie automatically if Python is installed. Otherwise use the manual flow above.

## Commands

<details>
<summary><b>Reading data</b></summary>

```bash
dida task today --json                       # Today's tasks
dida task upcoming --days 14 --json          # Next two weeks
dida task search --query "exam" --json       # Search tasks
dida project list --json                     # All projects
dida folder list --json                      # All folders
dida tag list --json                         # All tags
dida completed today --json                  # Completed today
dida pomo stats --json                       # Pomodoro stats
dida stats general --json                    # Account stats
```
</details>

<details>
<summary><b>Writing data</b></summary>

```bash
dida task create --project <id> --title "New task" --json
dida task update <task-id> --project <id> --title "Updated" --json
dida task complete <task-id> --project <id> --json
dida task move <task-id> --project <id> --to-project <dest> --json
dida task delete <task-id> --project <id> --yes --json
dida project create --name "New project" --json
dida tag create my-tag --json
```
</details>

<details>
<summary><b>Official channels (MCP & OpenAPI)</b></summary>

```bash
# Official MCP (token-based)
DIDA365_TOKEN=dp_xxx dida official doctor --json
dida official project list --json
dida official task query --query "today" --json

# Official OpenAPI (OAuth-based)
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
```
</details>

Full command reference: [docs/commands.md](docs/commands.md)

## Auth Channels

| | Web API | Official MCP | Official OpenAPI |
|---|---|---|---|
| **Auth method** | Browser cookie | Token | OAuth app |
| **Coverage** | Broadest (private endpoints) | MCP tool-based | Standard REST |
| **Write safety** | Dry-run + confirm | Dry-run | Dry-run |
| **Setup** | One browser login | Get token | Register OAuth app |

Use Web API for maximum coverage, OpenAPI for standard REST integration, MCP for official tool access. They use separate auth channels — never mixed.

## Agent Integration

DidaCLI is built for AI agent workflows. Agents can:

1. **Discover commands** — `dida schema list --compact --json`
2. **Build context** — `dida agent context --outline --json`
3. **Preview writes** — `--dry-run` before executing
4. **Parse responses** — stable JSON envelope

| Agent | Install |
|---|---|
| Claude Code | Copy [`skills/dida-cli/SKILL.md`](skills/dida-cli/SKILL.md) to your skills directory |
| Codex | See [docs/skill-installation.md](docs/skill-installation.md) |
| Hermes | See [docs/skill-installation.md](docs/skill-installation.md) |

### Agent Safety

When using DidaCLI with AI agents, **you are responsible for reviewing and approving all write operations**. Key safety boundaries:

- **Always preview first** — Agents should run `--dry-run` before any write. Review the generated payload before removing `--dry-run`.
- **Destructive operations require confirmation** — `task delete`, `project delete`, `tag merge` etc. require `--yes`. Never pass `--yes` blindly.
- **Agent mistakes are your responsibility** — If an agent creates, modifies, or deletes tasks/projects incorrectly, DidaCLI and its authors are not liable. You control the agent; the agent controls the CLI.
- **Token security** — Never share cookies or tokens with agents or in chat. DidaCLI stores tokens locally only; it does not transmit them to any third party.
- **Scope your agent's access** — Consider using a dedicated Dida365 account for agent experimentation, not your primary account.

## Documentation

- [Quick Start](docs/quickstart.md) — Running in 2 minutes
- [Commands Reference](docs/commands.md) — Every command, every flag
- [Agent Usage](docs/agent-usage.md) — Using DidaCLI with AI agents
- [API Coverage](docs/api-coverage.md) — Endpoint coverage map
- [OpenAPI Setup](docs/openapi-setup.md) — OAuth channel configuration
- [Distribution](docs/distribution.md) — Build from source, packaging

## Contributing

Contributions welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git
cd dida-cli
go test ./...
go build -o bin/dida ./cmd/dida
```

## License

[MIT](LICENSE)

## Disclaimer

DidaCLI is an independent, third-party open-source project. It is **not** affiliated with, endorsed by, or connected to [Dida365](https://dida365.com) / [TickTick](https://ticktick.com) (杭州随笔记网络技术有限公司 / Hangzhou Suibiji Network Technology Co., Ltd.). Provided "as is" for personal learning and research purposes only. The author assumes no responsibility for any consequences arising from the use of this tool.

**AI Agent usage:** When DidaCLI is operated by an AI agent (Claude, Codex, Hermes, etc.), the human operator is solely responsible for all actions performed. Always review agent-generated write operations before execution. Use `--dry-run` to preview. The CLI authors are not liable for data loss, account issues, or unintended modifications caused by agent actions.
