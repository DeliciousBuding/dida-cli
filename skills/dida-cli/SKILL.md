---
name: dida-cli
description: Use DidaCLI to operate Dida365/TickTick from agents. Trigger whenever the user asks to read, create, update, complete, delete, move, organize, or audit Dida365 tasks, projects, folders, tags, completed history, trash, or kanban columns. Prefer this skill over browser automation because DidaCLI provides stable JSON, auth checks, dry-run writes, and explicit destructive confirmations.
---

# DidaCLI Agent Skill

Use the `dida` command for Dida365/TickTick operations. It is built for agents: every `--json` response has a stable envelope, writes support `--dry-run`, and destructive actions require `--yes`.

## First Checks

Run these before doing useful work:

```bash
dida doctor --json
dida schema list --json
dida agent context --json
dida auth status --verify --json
```

If auth is missing or expired, ask the operator to run:

```bash
dida auth login --browser --json
```

Never ask the user to paste cookies, browser tokens, or raw `t=` values into chat. If manual cookie import is unavoidable, tell the operator to run `dida auth cookie set --token-stdin` locally.

Use the local schema command when selecting a command or checking safety flags:

```bash
dida schema show task.create --json
dida schema show comment.delete --json
```

Check `authRequired`, `dryRun`, and `confirmationRequired` in schema output
before choosing a command. If `authRequired` is true, verify the matching
channel auth first; do not assume a Web API cookie can satisfy official MCP or
OpenAPI commands.

## Read Context

Prefer the one-call context pack:

```bash
dida agent context --json
```

Use separate bounded JSON commands when you need a narrower response:

```bash
dida project list --json
dida folder list --json
dida tag list --json
dida filter list --json
dida column list <project-id> --json
dida comment list --project <project-id> --task <task-id> --json
dida +today --compact --json
dida task upcoming --days 14 --limit 50 --compact --json
dida task due-counts --json
dida quadrant list --json
dida completed today --compact --json
dida trash list --cursor 20 --compact --json
dida pomo list --limit 10 --json
dida pomo task --project <project-id> --task <task-id> --json
dida habit list --json
dida habit checkins --habit <habit-id> --json
dida attachment quota --json
dida reminder daily --json
dida share project shares <project-id> --json
dida calendar subscriptions --json
dida stats general --json
dida pomo timeline --limit 10 --json
dida template project list --limit 10 --json
dida search all --query <text> --limit 20 --json
```

Use exact IDs from read commands for writes. Do not guess project IDs, folder IDs, or task IDs from names if the command output is available.

Use `--compact` or `--brief` on broad task reads to reduce agent context. Compact task output keeps operational fields and drops large text, checklist, reminder, and raw fields. Fetch full task JSON only when those larger fields are needed.

## Task Writes

Preview generated task writes first:

```bash
dida task create --project <project-id> --title "Example" --dry-run --json
dida task update <task-id> --project <project-id> --title "New title" --dry-run --json
dida task move <task-id> --from <project-id> --to <project-id> --dry-run --json
dida task parent <task-id> --parent <parent-task-id> --project <project-id> --dry-run --json
```

Use first-class task field flags instead of raw payloads: `--content`, `--desc`, `--start`, `--due`, `--timezone`, `--tag`, `--tags`, `--item`, `--column`, `--reminder`, `--repeat`, `--repeat-from`, `--repeat-flag`, `--all-day`, and `--floating`.

Execute narrow writes only after the preview matches intent:

```bash
dida task create --project <project-id> --title "Example" --json
dida task update <task-id> --project <project-id> --title "New title" --json
dida task complete <task-id> --project <project-id> --json
```

Deletes require explicit confirmation:

```bash
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
```

## Resource Writes

Use high-level resource commands before raw API probes:

```bash
dida project create --name "Inbox review" --dry-run --json
dida project update <project-id> --name "New name" --dry-run --json
dida project delete <project-id> --yes --json

dida folder create --name "Work" --dry-run --json
dida folder update <folder-id> --name "Work archive" --dry-run --json
dida folder delete <folder-id> --yes --json

dida tag create planning --color "#147d4f" --dry-run --json
dida tag rename planning planning-next --dry-run --json
dida tag merge old-tag new-tag --yes --json
dida tag delete old-tag --yes --json
```

After `tag merge`, list tags again if the goal is to retire the source tag. The observed private endpoint can leave the source tag object present, so explicit `tag delete` may still be needed.

Column support is intentionally conservative:

```bash
dida column list <project-id> --json
dida column create --project <project-id> --name "Doing" --dry-run --json
```

Do not claim column update/delete support unless the CLI exposes those commands.

## Comments

Use comment commands for task discussion. Preview writes before execution:

```bash
dida comment list --project <project-id> --task <task-id> --json
dida comment create --project <project-id> --task <task-id> --text "Example" --dry-run --json
dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text "Updated" --dry-run --json
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --dry-run --json
```

Deletes require explicit confirmation:

```bash
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --yes --json
```

Do not use comment attachments yet; the CLI intentionally does not expose the multipart upload flow.

## Official MCP And OpenAPI

Keep official channels separate from browser-cookie Web API commands.

Use official MCP when the operator has configured `DIDA365_TOKEN`:

```bash
dida official doctor --json
dida official show get_focuses_by_time --json
dida official project data <project-id> --json
dida official task search --query "today" --json
dida official task query --query "today" --json
dida official task filter --project <project-id> --status 0 --json
dida official task batch-add --args-json "{\"tasks\":[{\"title\":\"Example\"}]}" --dry-run --json
dida official task complete-project --project <project-id> --task <task-id> --dry-run --json
dida official habit get <habit-id> --json
dida official habit checkin <habit-id> --date 2026-05-09 --value 1 --json
dida official focus list --start-time 2026-05-01T00:00:00+08:00 --end-time 2026-05-09T23:59:59+08:00 --json
```

Use official OpenAPI only through `dida openapi ...`. It is OAuth-based and
does not accept MCP `dp_...` tokens or Web API cookies as bearer tokens.

```bash
dida openapi doctor --json
dida openapi client status --json
dida openapi project list --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi habit list --json
dida openapi habit checkin <habit-id> --args-json "{\"stamp\":20260407,\"value\":1}" --dry-run --json
```

If OpenAPI client config is missing, ask the operator to run this locally:

```bash
dida openapi client set --id <client-id> --secret-stdin --json
```

Do not delete focus records unless the operator has identified a disposable
record; both `dida official focus delete` and `dida openapi focus delete`
require `--yes`.

## Account Metadata

Use these read-only commands for operational context around uploads, reminders,
sharing, and calendar integrations:

```bash
dida attachment quota --json
dida reminder daily --json
dida share contacts --json
dida share recent-users --json
dida share project shares <project-id> --json
dida share project quota <project-id> --json
dida share project invite-url <project-id> --json
dida calendar subscriptions --json
```

Do not create invite links, delete invite links, invite users, or change
sharing permissions unless the CLI exposes a first-class command with dry-run or
confirmation behavior.

## Error Handling

If a JSON command returns `ok: false`, surface `error.message` and `error.hint` to the operator. Do not retry destructive operations automatically.

Example error shape:

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

## Raw Escape Hatch

Use raw reads only when a high-level command is missing:

```bash
dida raw get /batch/check/0 --json
dida raw get /user/preferences/settings --json
dida raw get /attachment/isUnderQuota --api-version v1 --json
```

Raw writes are intentionally unavailable. Add a first-class command with tests instead of tunneling writes through raw calls.
