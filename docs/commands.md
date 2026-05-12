# Command Reference

This page documents the stable command surface. Prefer `--json` when calling DidaCLI from agents or scripts.

Use `--compact` or `--brief` on task-heavy reads when an agent only needs IDs, titles, dates, priority, status, column, and tags. Compact output omits large text, checklist, reminder, and raw fields.

## Auth

```bash
dida doctor --verify --json
dida auth login --browser --json
dida auth cookie set --token-stdin
dida auth status --verify --json
dida auth logout --json
```

Credentials are stored in `~/.dida-cli/`. Full cookie values are never printed.
Prefer `dida auth login --browser --json` for normal setup. `dida auth login
--json` remains a manual fallback that prints the login URL and then expects
cookie import through `dida auth cookie set --token-stdin`. Cookie import
through `--token` is disabled by default; use `--token-stdin` so the cookie does
not enter shell history.

## Schema

```bash
dida schema list --compact --json
dida schema show task.create --json
dida schema show column.create --json
dida schema show openapi.login --json
dida channel list --json
```

Schema output is local and does not require auth. Compact schema output exposes
command IDs, command strings, status, auth requirements, `--dry-run` support,
destructive confirmation requirements, and compact-output support. Use
`schema show` when you need Web API endpoints, notes, aliases, or full details.
Channel output is also local and explains when to use Web API, Official MCP, or
Official OpenAPI without mixing auth models.

## Agent Context

```bash
dida agent context --json
dida agent context --outline --json
dida agent context --days 30 --limit 100 --json
dida agent context --full --json
```

`agent context` performs one full sync and returns a compact context pack:
projects, folders, tags, filters, today's tasks, upcoming tasks, and quadrant
buckets. It is the preferred first read for automation because it avoids several
separate sync calls. Compact mode is on by default; use `--full` only when large
task text, checklist, reminder, and raw fields are needed.

Use `--outline` when token budget matters. It keeps project/filter metadata,
returns `today`, `upcoming`, and quadrant buckets as task id references, and
adds a deduplicated `taskIndex` with compact task objects.

## Discovery

```bash
dida project list --json
dida official doctor --json
dida official token status --json
dida official token set --token-stdin --json
dida official tools --limit 20 --json
dida official show list_projects --json
dida official project list --json
dida official project get <project-id> --json
dida official project data <project-id> --json
dida official task search --query "today" --json
dida official task query --query "today" --json
dida official task undone --start 2026-05-01T00:00:00+08:00 --end 2026-05-09T23:59:59+08:00 --json
dida official task filter --project <project-id> --status 0 --json
dida official task batch-add --args-json "{\"tasks\":[{\"title\":\"Task\"}]}" --dry-run --json
dida official task batch-update --args-json "{\"tasks\":[{\"id\":\"<task-id>\",\"title\":\"Task\"}]}" --dry-run --json
dida official task complete-project --project <project-id> --task <task-id> --dry-run --json
dida official habit list --json
dida official habit sections --json
dida official habit get <habit-id> --json
dida official habit create --args-json "{\"name\":\"Read\",\"type\":\"Boolean\"}" --dry-run --json
dida official focus list --from-time 2026-05-01T00:00:00+08:00 --to-time 2026-05-09T23:59:59+08:00 --type 1 --json
dida openapi doctor --json
dida openapi status --json
dida openapi logout --json
dida openapi client status --json
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi auth-url --json
dida openapi exchange-code --code <code> --json
dida openapi project list --json
dida openapi project get <project-id> --json
dida openapi project data <project-id> --json
dida openapi project create --args-json "{\"name\":\"Project\",\"viewMode\":\"list\",\"kind\":\"TASK\"}" --dry-run --json
dida openapi project update <project-id> --args-json "{\"name\":\"Renamed\"}" --dry-run --json
dida openapi project delete <project-id> --dry-run --json
dida openapi project delete <project-id> --yes --json
dida openapi task get --project <project-id> --task <task-id> --json
dida openapi task create --args-json "{\"projectId\":\"<project-id>\",\"title\":\"Task\"}" --dry-run --json
dida openapi task update <task-id> --args-json "{\"projectId\":\"<project-id>\",\"title\":\"Task\"}" --dry-run --json
dida openapi task complete --project <project-id> --task <task-id> --dry-run --json
dida openapi task delete --project <project-id> --task <task-id> --yes --json
dida openapi task move --args-json "[{\"fromProjectId\":\"<from>\",\"toProjectId\":\"<to>\",\"taskId\":\"<task-id>\"}]" --dry-run --json
dida openapi task completed --args-json "{\"projectIds\":[\"<project-id>\"]}" --json
dida openapi task filter --args-json "{\"projectIds\":[\"<project-id>\"],\"status\":[0]}" --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi habit list --json
dida openapi habit checkins --habit-ids <habit-id> --from 20260401 --to 20260407 --json
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

Use `dida official token clear --json` only when intentionally removing the
saved local official MCP token. Use `dida official call ...` primarily for
read-only exploration after checking `dida official show <tool-name> --json`;
prefer first-class `official project/task/habit/focus` wrappers for normal
work.

For OpenAPI OAuth setup, read `default_redirect_uri`, `default_scope`, and
`next` from `dida openapi doctor --json`. Configure the developer app redirect
URL to that exact URI before running `dida openapi login --browser --json`.

Use `official call` primarily for read-only exploration and schema-backed
probes. For writes, prefer first-class `official task/habit/focus ... --dry-run`
wrappers or get explicit operator approval; `official call` itself has no
dry-run layer.

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
dida comment create --project <project-id> --task <task-id> --text "See attachment" --file ./probe.png --dry-run --json
dida comment create --project <project-id> --task <task-id> --text "See attachment" --file ./probe.png --json
dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text "Updated" --dry-run --json
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --dry-run --json
dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --yes --json
```

Comment attachment create is verified for the Web API v1 multipart field `file`.
Use the real project id from `dida agent context --json`, not the logical
`inbox` alias. Task-level attachment download/preview and task attachment
mutation remain intentionally unexposed until separately verified. The CLI
checks `dida attachment quota --json` before uploading files.

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
dida trash list --cursor 20 --limit 20 --compact --json
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

## Official MCP Project, Habit, And Focus

```bash
dida official project list --json
dida official project get <project-id> --json
dida official project data <project-id> --json

dida official task get <task-id> --project <project-id> --json
dida official task search --query "today" --json
dida official task query --query "today" --json
dida official task undone --project <project-id> --start 2026-05-01T00:00:00+08:00 --end 2026-05-09T23:59:59+08:00 --json
dida official task filter --project <project-id> --priority 0,5 --tag work --status 0 --json
dida official task batch-add --args-json "{\"tasks\":[{\"title\":\"Task\"}]}" --dry-run --json
dida official task batch-update --args-json "{\"tasks\":[{\"id\":\"<task-id>\",\"title\":\"Task\"}]}" --dry-run --json
dida official task complete-project --project <project-id> --task <task-id> --dry-run --json

dida official habit list --json
dida official habit sections --json
dida official habit get <habit-id> --json
dida official habit create --args-json "{\"name\":\"Read\",\"type\":\"Boolean\",\"goal\":1}" --dry-run --json
dida official habit update <habit-id> --args-json "{\"name\":\"Read more\"}" --dry-run --json
dida official habit checkin <habit-id> --date 2026-05-09 --value 1 --dry-run --json
dida official habit checkins --habit-ids <habit-id> --from 20260501 --to 20260510 --json

dida official focus get <focus-id> --type 0 --json
dida official focus list --from-time 2026-05-01T00:00:00+08:00 --to-time 2026-05-09T23:59:59+08:00 --type 1 --json
dida official focus delete <focus-id> --type 0 --dry-run --json
```

These commands use the official MCP channel. `DIDA365_TOKEN` takes precedence,
or save the token locally with `dida official token set --token-stdin --json`.
Use `dida official show <tool-name> --json` to inspect the exact upstream
schema before passing larger `--args-json` payloads. Web API habit and Pomodoro
commands remain separate under `dida habit` and `dida pomo`. Official task
batch commands support local `--dry-run` previews without a token; remove
`--dry-run` only after checking the upstream schema and payload.

## Official OpenAPI Focus And Habit

```bash
dida openapi client status --json
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi client clear --json
dida openapi status --json
dida openapi logout --json

dida openapi focus get <focus-id> --type 0 --json
dida openapi focus list --from 2026-04-01T00:00:00+0800 --to 2026-04-02T00:00:00+0800 --type 1 --json
dida openapi focus delete <focus-id> --type 0 --dry-run --json
dida openapi focus delete <focus-id> --type 0 --yes --json

dida openapi habit list --json
dida openapi habit get <habit-id> --json
dida openapi habit create --args-json "{\"name\":\"Read\",\"type\":\"Boolean\",\"goal\":1}" --dry-run --json
dida openapi habit update <habit-id> --args-json "{\"name\":\"Read more\"}" --dry-run --json
dida openapi habit checkin <habit-id> --args-json "{\"stamp\":20260407,\"value\":1,\"goal\":1}" --dry-run --json
dida openapi habit checkins --habit-ids <habit-id> --from 20260401 --to 20260407 --json
```

These commands use the official OAuth OpenAPI channel and require a saved
OpenAPI access token from `dida openapi login`. Focus `--type` follows the
official API values: `0` for Pomodoro and `1` for Timing. Habit and focus write
commands support `--dry-run`; focus delete requires `--yes` when executed.
`dida openapi login --browser --json` emits one final JSON envelope after browser
authorization completes. For manual no-browser OAuth, use `dida openapi
auth-url --json` and `dida openapi listen-callback --json`.
`dida openapi client set` stores the OAuth client id and secret locally; use
`--secret-stdin` so the secret does not enter shell history. Environment
variables still take precedence.

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
dida raw get /path --api-version v1|v2 --json
dida raw get /batch/check/0 --json
dida raw get /user/preferences/settings --json
dida raw get /attachment/isUnderQuota --api-version v1 --json
```

Raw calls are intentionally GET-only. On JSON failures, `raw get` includes
`error.details.statusCode`, `error.details.path`, and a short
`error.details.bodySnippet` so private API probes can distinguish entitlement
errors, path mistakes, and server failures without enabling raw writes.

## Upgrade

```bash
dida upgrade --json              # Check for updates and self-upgrade
dida upgrade --check --json      # Only check, do not install
```

The updater queries GitHub Releases for the latest version, downloads the
platform-matched binary archive, verifies the SHA-256 checksum against
`checksums.txt`, extracts the binary, and replaces the current executable.
Requires write permission to the binary location. On Windows, uses a
rename-then-replace strategy to avoid file-lock issues.
