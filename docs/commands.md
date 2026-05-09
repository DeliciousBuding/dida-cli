# Command Reference

This page documents the stable command surface. Prefer `--json` when calling DidaCLI from agents or scripts.

Use `--compact` or `--brief` on task-heavy reads when an agent only needs IDs, titles, dates, priority, status, column, and tags. Compact output omits large text, checklist, reminder, and raw fields.

## Auth

```bash
dida auth login --browser --json
dida auth login --json
dida auth cookie set --token-stdin
dida auth status --verify --json
dida auth logout --json
```

Credentials are stored in `~/.dida-cli/`. Full cookie values are never printed.
Cookie import through `--token` is disabled by default; use `--token-stdin` so
the cookie does not enter shell history.

## Schema

```bash
dida schema list --json
dida schema show task.create --json
dida schema show column.create --json
```

Schema output is local and does not require auth. It exposes command IDs, command
strings, status, Web API endpoints, `--dry-run` support, destructive
confirmation requirements, and compact-output support for agents that need to
choose the safest command without reading the full README.

## Agent Context

```bash
dida agent context --json
dida agent context --days 30 --limit 100 --json
dida agent context --full --json
```

`agent context` performs one full sync and returns a compact context pack:
projects, folders, tags, filters, today's tasks, upcoming tasks, and quadrant
buckets. It is the preferred first read for automation because it avoids several
separate sync calls. Compact mode is on by default; use `--full` only when large
task text, checklist, reminder, and raw fields are needed.

## Discovery

```bash
dida project list --json
dida official doctor --json
dida official tools --limit 20 --json
dida official show list_projects --json
dida official call list_projects --json
dida official call list_undone_tasks_by_time_query --args-json "{\"query_command\":\"today\"}" --json
dida openapi doctor --json
dida openapi auth-url --json
dida openapi exchange-code --code <code> --json
dida openapi project list --json
dida project tasks <project-id> --limit 50 --compact --json
dida project columns <project-id> --json
dida folder list --json
dida tag list --json
dida filter list --json
dida column list <project-id> --json
dida comment list --project <project-id> --task <task-id> --json
dida settings get --json
dida settings get --include-web --json
dida attachment quota --json
dida reminder daily --json
dida share contacts --json
dida share recent-users --json
dida share project shares <project-id> --json
dida share project quota <project-id> --json
dida share project invite-url <project-id> --json
dida calendar subscriptions --json
dida calendar archived --json
dida calendar third-accounts --json
dida stats general --json
dida template project list --limit 50 --json
dida search all --query "计算机" --limit 20 --json
dida search all --query "计算机" --limit 20 --full --json
dida user status --json
dida user profile --json
dida user sessions --limit 10 --json
```

## Tasks

```bash
dida +today --json
dida task today --compact --json
dida task list --filter today --limit 20 --compact --json
dida task list --filter all --limit 50 --compact --json
dida task get <task-id> --json
dida task search --query <text> --limit 20 --compact --json
dida task upcoming --days 14 --limit 50 --compact --json
dida task due-counts --json
dida quadrant list --json
dida quadrant view Q2 --json
```

## Task Writes

```bash
dida task create --project <project-id> --title <title> --dry-run --json
dida task create --project <project-id> --title <title> --content <text> --priority 3 --json
dida task create --project <project-id> --title <title> --tag work --item "Checklist item" --json
dida task update <task-id> --project <project-id> --title <title> --json
dida task update <task-id> --project <project-id> --priority 0 --json
dida task update <task-id> --project <project-id> --tags work,deep --column <column-id> --json
dida task complete <task-id> --project <project-id> --json
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
dida task move <task-id> --from <project-id> --to <project-id> --dry-run --json
dida task move <task-id> --from <project-id> --to <project-id> --json
dida task parent <task-id> --parent <parent-task-id> --project <project-id> --dry-run --json
```

Priority values follow Dida365 Web API conventions:

| Value | Meaning |
|---:|---|
| `0` | none |
| `1` | low |
| `3` | medium |
| `5` | high |

Task create/update support these Web API fields:

| Flag | Purpose |
|---|---|
| `--content <text>` | Plain task content |
| `--desc <markdown>` | Rich description field |
| `--start <time>` / `--due <time>` | Start and due timestamps |
| `--timezone <zone>` | IANA timezone, for example `Asia/Shanghai` |
| `--tag <name>` / `--tags a,b` | Tags |
| `--item <title>` | Checklist item; repeatable |
| `--column <id>` | Kanban column id |
| `--reminder <value>` | Reminder value; repeatable |
| `--repeat <rule>` / `--repeat-from <value>` / `--repeat-flag <value>` | Repeat metadata |
| `--all-day` / `--not-all-day` | All-day toggle |
| `--floating` / `--not-floating` | Floating task toggle |

## Project, Folder, Tag, and Column Writes

```bash
dida project create --name "New project" --dry-run --json
dida project create --name "New project" --group <folder-id> --json
dida project update <project-id> --name "Renamed project" --json
dida project delete <project-id> --dry-run --json
dida project delete <project-id> --yes --json

dida folder create --name "Work" --dry-run --json
dida folder update <folder-id> --name "Work archive" --json
dida folder delete <folder-id> --dry-run --json
dida folder delete <folder-id> --yes --json

dida tag create planning --color "#147d4f" --dry-run --json
dida tag update planning --color "#147d4f" --json
dida tag rename planning planning-next --json
dida tag merge old-tag new-tag --dry-run --json
dida tag merge old-tag new-tag --yes --json
dida tag delete old-tag --dry-run --json
dida tag delete old-tag --yes --json

dida column create --project <project-id> --name "Doing" --dry-run --json
dida column create --project <project-id> --name "Doing" --json
```

Column list is backed by `GET /column/project/{projectId}`. Column create is backed by an observed private Web API endpoint. Column update/delete/order commands are intentionally not exposed until payloads are verified.

## Comments

```bash
dida comment list --project <project-id> --task <task-id> --json
dida comment create --project <project-id> --task <task-id> --text "Looks good" --dry-run --json
dida comment create --project <project-id> --task <task-id> --text "Looks good" --json
dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text "Updated" --dry-run --json
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --dry-run --json
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --yes --json
```

Comment attachments are intentionally not exposed until the multipart upload and attach flow is verified.

## Filters

```bash
dida filter list --json
```

Filters are read from the sync payload. Filter writes are intentionally not exposed until `/batch/filter` payloads are verified from a real webapp trace.

Observed merge behavior: `tag merge` moves associations through the private endpoint but may leave the source tag object present. Delete the source tag explicitly when the intended outcome is full retirement.

## Completed History

```bash
dida completed today --json
dida completed yesterday --json
dida completed week --json
dida completed list --from 2026-05-01 --to 2026-05-09 --limit 100 --compact --json
dida closed list --status 2 --from 2026-05-01 --to 2026-05-09 --limit 50 --json
```

## Pomodoro And Habits

```bash
dida pomo preferences --json
dida pomo list --from 2026-05-01 --to 2026-05-09 --limit 20 --json
dida pomo timing --from 2026-05-01 --to 2026-05-09 --limit 20 --json
dida pomo task --project <project-id> --task <task-id> --json
dida pomo stats --json
dida pomo timeline --limit 20 --json

dida habit preferences --json
dida habit list --json
dida habit sections --json
dida habit checkins --habit <habit-id> --after-stamp <millis> --json
```

Pomodoro range commands convert `YYYY-MM-DD` flags into the millisecond
timestamps expected by the private Web API. The default range is the last 30
days and the default output limit is 50.

## Account Metadata

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

These commands are read-only. Collaboration writes such as creating invite
links, deleting invite links, inviting users, or changing permissions are not
exposed until their payloads and rollback paths are traced.

## Sync

```bash
dida sync all --json
dida sync checkpoint <checkpoint> --json
```

`sync all` returns the latest checkpoint in `meta.checkpoint`.

`sync checkpoint` returns both a normalized `data.view` and raw-compatible `data.deltas` for task additions, updates, deletes, order changes, and reminder changes.

## Raw Read-Only Probe

```bash
dida raw get /batch/check/0 --json
dida raw get /user/preferences/settings --json
dida raw get /attachment/isUnderQuota --api-version v1 --json
```

Raw calls are intentionally GET-only.
