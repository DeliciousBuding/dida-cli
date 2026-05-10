# Agent Usage Guide

Use this guide when DidaCLI is called from Hermes, Codex, Claude Code, or another automation agent.

## First Commands

```bash
dida doctor --json
dida schema list --json
dida agent context --outline --json
dida auth status --verify --json
```

If auth is missing, ask the operator to run:

```bash
dida auth login --browser --json
```

Do not ask the user to paste cookies into chat.

Use `dida schema show <id> --json` when you need the exact command contract for
a write or less common resource. The schema command is local, auth-free, and
lists whether `--dry-run`, `--yes`, or `--compact` applies.

## Channel Choice

Pick the channel by job:

| Job | Prefer | Notes |
| --- | --- | --- |
| First account read | `dida agent context --outline --json` | Web API context pack with compact task refs. |
| Normal task/project/folder/tag/comment work | Web API first-class commands | Broadest coverage, dry-run writes, explicit `--yes` deletes. |
| Official token-based task/project validation | `dida official ...` | Uses `DIDA365_TOKEN` or saved official token config, not browser cookies. |
| Habit/focus work | Official MCP or `dida openapi ...` | Prefer official surfaces; live write tests need disposable records. |
| Public OAuth REST validation | `dida openapi ...` | Requires OpenAPI OAuth access token. |
| Web-app-only metadata | Web API reads | Settings, sharing, calendar, templates, stats, trash, search, closed history. |
| Unknown private write flow | No command yet | Document endpoint, payload, response, rollback, and live evidence first. |

## Context Pack

Prefer the one-call context pack. Use outline mode first when an agent only
needs IDs and compact task fields:

```bash
dida agent context --outline --json
dida agent context --json
```

Use separate reads when you need a narrower or resource-specific response:

```bash
dida project list --json
dida folder list --json
dida tag list --json
dida filter list --json
dida column list <project-id> --json
dida comment list --project <project-id> --task <task-id> --json
dida +today --compact --json
dida task upcoming --days 14 --limit 50 --compact --json
dida quadrant list --json
dida completed today --compact --json
dida pomo list --limit 10 --json
dida habit list --json
dida official doctor --json
dida official focus list --from-time 2026-05-01T00:00:00+08:00 --to-time 2026-05-09T23:59:59+08:00 --type 1 --json
```

Prefer `--compact` for broad task reads. It keeps IDs, titles, dates, priority, status, columns, and tags while omitting large descriptions, checklist items, reminders, and raw payloads. Use full JSON only when you need those fields for a specific task.
Use `agent context --outline` when token budget matters; it replaces repeated
task objects in today/upcoming/quadrants with task id references and a
deduplicated `taskIndex`.

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

Use comment commands for task discussion. Preview comment writes, and require `--yes` for deletes:

```bash
dida comment create --project <project-id> --task <task-id> --text "Example" --dry-run --json
dida comment create --project <project-id> --task <task-id> --text "Example" --file ./image.png --dry-run --json
dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text "Updated" --dry-run --json
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --yes --json
```

For comment attachments, use the real project id from `dida agent context
--json`, not the logical `inbox` alias. Preview with `--dry-run` first. Task-level
attachment download/preview and task attachment mutation are not exposed yet.

Use `dida sync checkpoint <checkpoint> --json` when an agent needs deletions, order deltas, or reminder deltas; those live under `data.deltas`.

## Official Channels

Use `dida official ...` only for the official MCP channel. It requires
`DIDA365_TOKEN` or saved official token config and is separate from
browser-cookie Web API auth.

```bash
dida official tools --limit 20 --json
dida official project data <project-id> --json
dida official task search --query "today" --json
dida official task filter --project <project-id> --status 0 --json
dida official show get_focuses_by_time --json
dida official habit get <habit-id> --json
dida official habit checkin <habit-id> --date 2026-05-09 --value 1 --json
dida official focus list --from-time 2026-05-01T00:00:00+08:00 --to-time 2026-05-09T23:59:59+08:00 --type 1 --json
```

Use `dida openapi ...` only for the official OAuth OpenAPI channel. It requires
OAuth client credentials plus a saved OAuth access token. Do not try to use MCP
tokens as OpenAPI bearer tokens.

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
