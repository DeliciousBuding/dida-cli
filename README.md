---
purpose: 面向滴答清单 Dida365 和 TickTick 的 Agent 友好型 Go CLI，提供任务、项目、标签、番茄统计等读写命令与稳定 JSON 输出。
status: active
next_steps:
  - 持续维护 Web API、官方 MCP、OpenAPI 三条认证通道的命令覆盖面和文档
capabilities:
  - dida365-cli
  - ticktick-cli
  - task-management-cli
  - agent-friendly-json-cli
  - mcp-openapi-integration
---
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

## Commands

<details>
<summary>Read</summary>

```bash
dida task today --json
dida task upcoming --days 14 --json
dida task search --query "exam" --json
dida project list --json
dida tag list --json
dida completed today --json
dida pomo stats --json
```
</details>

<details>
<summary>Write</summary>

```bash
dida task create --project <id> --title "New task" --json
dida task update <id> --project <p> --title "Updated" --json
dida task complete <id> --project <p> --json
dida task move <id> --from <p> --to <p> --json
dida task delete <id> --project <p> --yes --json
```
</details>

<details>
<summary>Official channels (MCP & OpenAPI)</summary>

```bash
# MCP (token)
DIDA365_TOKEN=dp_xxx dida official project list --json

# OpenAPI (OAuth)
dida openapi client set --id <id> --secret-stdin --json
dida openapi login --browser --json
dida openapi project list --json
```
</details>

Full reference: [docs/commands.md](docs/commands.md)

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

## License

[MIT](LICENSE)

---

DidaCLI is an independent open-source CLI for [Dida365](https://dida365.com) / [TickTick](https://ticktick.com)-compatible workflows. Use it with accounts and automations you control, subject to the upstream services' terms.
