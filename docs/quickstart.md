# Quickstart

This guide is for humans and operators who want DidaCLI installed, verified,
authenticated, and ready for daily task automation.

## Install

Unix-like systems:

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

Optional environment variables:

| Variable | Purpose |
| --- | --- |
| `DIDA_VERSION` | Install a specific release tag, for example `v0.1.0`. |
| `DIDA_INSTALL_DIR` | Override the install directory. |
| `DIDA_REPO` | Override the GitHub repository, for example for forks. |

## Verify

```bash
dida version
dida doctor --json
```

## Web API Login

The Web API channel uses the browser session cookie captured locally.

```bash
dida auth login --browser --json
dida auth status --verify --json
```

Do not paste cookies into chat or issue trackers. Manual cookie import, when
needed, should use stdin:

```bash
dida auth cookie set --token-stdin
```

## Agent First Read

```bash
dida agent context --json
```

This gives agents a compact account context: projects, folders, tags, filters,
today, upcoming, and quadrants.

## Schema Discovery

```bash
dida schema list --json
dida schema show task.create --json
```

Inspect schema before generated writes. It tells agents which commands support
`--dry-run`, `--yes`, and compact output.

## Official MCP

Official MCP is the token-based official channel. It is separate from Web API
cookie auth.

```bash
DIDA365_TOKEN=... dida official doctor --json
dida official tools --limit 20 --json
dida official project data <project-id> --json
dida official task filter --project <project-id> --status 0 --json
```

## Official OpenAPI

Official OpenAPI is OAuth-based REST. It is separate from Web API and official
MCP.

```bash
dida openapi doctor --json
dida openapi login --json
dida openapi project list --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi habit list --json
dida openapi habit checkin <habit-id> --args-json '{"stamp":20260407,"value":1}' --dry-run --json
```

## Common Reads

```bash
dida project list --json
dida task today --compact --json
dida task upcoming --days 14 --limit 50 --compact --json
dida task search --query "paper" --limit 20 --compact --json
dida completed today --compact --json
dida trash list --cursor 20 --compact --json
dida stats general --json
```

## Common Writes

Preview generated writes first:

```bash
dida task create --project <project-id> --title "Read paper" --dry-run --json
dida task update <task-id> --project <project-id> --title "Read paper carefully" --dry-run --json
```

Execute after the preview is correct:

```bash
dida task create --project <project-id> --title "Read paper" --json
dida task complete <task-id> --project <project-id> --json
```

Destructive operations require `--yes`:

```bash
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
```

## Agent Note

This section is optimized for LLM/Agent operators. Prefer JSON commands,
inspect `dida schema list --json` before writes, preview generated writes with
`--dry-run`, and never ask the user to paste cookies or tokens into chat.
