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
dida folder list --json
dida tag list --json
dida +today --json
dida task upcoming --days 14 --limit 50 --json
dida quadrant list --json
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

The same pattern applies to resources:

```bash
dida project create --name "Agent staging" --dry-run --json
dida folder create --name "Agent staging" --dry-run --json
dida tag create agent-staging --dry-run --json
dida project delete <project-id> --yes --json
dida folder delete <folder-id> --yes --json
dida tag delete agent-staging --yes --json
```

Use `dida column create` only when the operator accepts that column support is based on an experimental private endpoint. The CLI does not expose column update/delete yet.

Use `dida sync checkpoint <checkpoint> --json` when an agent needs deletions, order deltas, or reminder deltas; those live under `data.deltas`.

## Repo Skill

This repository includes `skills/dida-cli/SKILL.md` for Codex, Claude Code, OpenClaw, and Hermes Agent. Install instructions are in [skill-installation.md](skill-installation.md).

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
