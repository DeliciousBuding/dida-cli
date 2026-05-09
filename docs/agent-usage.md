# Agent Usage Guide

Use this guide when DidaCLI is called from Hermes, Codex, Claude Code, or another automation agent.

## First Commands

```bash
dida doctor --json
dida auth status --verify --json
```

If auth is missing, ask the operator to run:

```bash
dida auth login --browser --json
```

Do not ask the user to paste cookies into chat.

## Context Pack

```bash
dida project list --json
dida +today --json
dida task upcoming --days 14 --limit 50 --json
dida completed today --json
```

## Safe Writes

For generated writes, preview first:

```bash
dida task create --project <project-id> --title "Example" --dry-run --json
```

Then execute:

```bash
dida task create --project <project-id> --title "Example" --json
```

Delete requires explicit confirmation:

```bash
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
```

## Error Handling

All JSON errors use:

```json
{
  "ok": false,
  "command": "task delete",
  "error": {
    "type": "confirmation_required",
    "message": "...",
    "hint": "..."
  }
}
```

Agents should surface `error.hint` to the operator instead of guessing.
