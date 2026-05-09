# Command Reference

This page documents the stable command surface. Prefer `--json` when calling DidaCLI from agents or scripts.

## Auth

```bash
dida auth login --browser --json
dida auth login --json
dida auth cookie set --token-stdin
dida auth status --verify --json
dida auth logout --json
```

Credentials are stored in `~/.dida-cli/`. Full cookie values are never printed.

## Discovery

```bash
dida project list --json
dida project tasks <project-id> --json
dida project columns <project-id> --json
dida settings get --json
```

## Tasks

```bash
dida +today --json
dida task today --json
dida task list --filter today --limit 20 --json
dida task list --filter all --limit 50 --json
dida task get <task-id> --json
dida task search --query <text> --limit 20 --json
dida task upcoming --days 14 --limit 50 --json
```

## Task Writes

```bash
dida task create --project <project-id> --title <title> --dry-run --json
dida task create --project <project-id> --title <title> --content <text> --priority 3 --json
dida task update <task-id> --project <project-id> --title <title> --json
dida task complete <task-id> --project <project-id> --json
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
```

Priority values follow Dida365 Web API conventions:

| Value | Meaning |
|---:|---|
| `0` | none |
| `1` | low |
| `3` | medium |
| `5` | high |

## Completed History

```bash
dida completed today --json
dida completed yesterday --json
dida completed week --json
dida completed list --from 2026-05-01 --to 2026-05-09 --limit 100 --json
```

## Sync

```bash
dida sync all --json
dida sync checkpoint <checkpoint> --json
```

`sync all` returns the latest checkpoint in `meta.checkpoint`.

## Raw Read-Only Probe

```bash
dida raw get /batch/check/0 --json
dida raw get /user/preferences/settings --json
```

Raw calls are intentionally GET-only.
