# API Coverage Matrix

This matrix tracks the Dida365 Web API surfaces that DidaCLI intentionally supports. It separates verified command coverage from experimental or unknown private endpoints.

## Coverage Status

| Area | Web API surface | CLI command | Status | Verification |
|---|---|---|---|---|
| Auth | `Cookie: t=<token>` | `auth login`, `auth cookie set`, `auth status --verify`, `auth logout` | Stable | Unit tests and live auth verify |
| Full sync | `GET /batch/check/0` | `sync all` | Stable | Unit tests and live read |
| Incremental sync | `GET /batch/check/{checkpoint}` | `sync checkpoint` | Stable | Unit tests preserve add/update/delete/order/reminder deltas |
| Settings | `GET /user/preferences/settings` | `settings get` | Stable | Live read |
| Completed tasks | `GET /project/all/completed` | `completed today/yesterday/week/list` | Stable | Unit query test and live read |
| Project task list | `GET /project/{projectId}/tasks` | `project tasks <project-id> --limit N` | Stable | Unit endpoint test and live read |
| Project columns | `GET /column/project/{projectId}` | `project columns`, `column list` | Stable read | Unit endpoint test and live read |
| Task CRUD | `POST /batch/task` | `task create/update/complete/delete` | Stable | Unit request tests and live reversible smoke |
| Task advanced fields | `/batch/task` fields | `--content`, `--desc`, `--start`, `--due`, `--timezone`, `--tag`, `--tags`, `--item`, `--column`, `--reminder`, `--repeat*`, `--all-day`, `--floating`, `--priority 0` | Stable | Unit request tests and live reversible smoke |
| Task due activity counts | `POST /task/activity/count/all` | `task due-counts` | Stable read | Unit endpoint test and live read |
| Task move | `POST /batch/taskProject` | `task move` | Stable | Unit request tests and live reversible smoke |
| Task parent/subtask | `POST /batch/taskParent` | `task parent` | Stable | Unit request tests and live reversible smoke |
| Project CRUD | `POST /batch/project` | `project create/update/delete` | Stable | Unit request tests and live reversible smoke |
| Folder CRUD | `POST /batch/projectGroup` | `folder create/update/delete` | Stable | Unit request tests and live reversible smoke |
| Tag create/update | `POST /batch/tag` | `tag create/update` | Stable | Unit request tests and live reversible smoke |
| Tag rename | `PUT /tag/rename` | `tag rename` | Stable | Unit request tests and live reversible smoke |
| Tag merge | `PUT /tag/merge` | `tag merge --yes` | Stable with caveat | Unit request tests and live smoke; source tag may remain |
| Tag delete | `DELETE /tag?name=...` | `tag delete --yes` | Stable | Unit URL escaping test and live reversible smoke |
| Filters | sync payload `filters` | `filter list` | Stable read | Unit sync-view test and live read |
| Column create | `POST /column` | `column create` | Experimental | Unit request test; live write avoided because delete endpoint is unknown |
| Task comments | `GET/POST/PUT/DELETE /project/{projectId}/task/{taskId}/comment(s)` | `comment list/create/update/delete` | Stable without attachments | Unit request tests and reversible live smoke |
| Attachment quota | `GET /api/v1/attachment/isUnderQuota`, `GET /api/v1/attachment/dailyLimit` | `attachment quota` | Stable read | Unit endpoint test and live read |
| Daily reminder preferences | `GET /user/preferences/dailyReminder` | `reminder daily` | Stable read | Unit endpoint test and live read |
| Sharing contacts | `GET /share/shareContacts`, `GET /project/share/recentProjectUsers` | `share contacts`, `share recent-users` | Stable read | Unit endpoint test and live read |
| Project share state | `GET /project/{projectId}/shares`, `GET /project/{projectId}/share/check-quota`, `GET /project/{projectId}/collaboration/invite-url` | `share project shares/quota/invite-url` | Stable read | Unit endpoint test and live read |
| Calendar subscriptions | `GET /calendar/subscription` | `calendar subscriptions` | Stable read | Unit endpoint test and live read |
| Calendar archived and accounts | `GET /calendar/archivedEvent`, `GET /calendar/third/accounts` | `calendar archived`, `calendar third-accounts` | Stable read | Unit endpoint test and live read |
| Statistics | `GET /statistics/general` | `stats general` | Stable read | Unit endpoint test and live read |
| Project templates | `GET /projectTemplates/all?timestamp=...` | `template project list` | Stable read | Unit endpoint test and live read |
| Search | `GET /search/all?keywords=...` | `search all` | Stable read | Unit endpoint test and live read |
| User metadata | `GET /user/status`, `GET /user/profile`, `GET /user/sessions?lang=...` | `user status`, `user profile`, `user sessions` | Stable read | Unit endpoint test and live read |
| Pomodoro preferences | `GET /user/preferences/pomodoro` | `pomo preferences` | Stable read | Live read |
| Pomodoro records | `GET /pomodoros`, `GET /pomodoros/timing` | `pomo list`, `pomo timing` | Stable read | Live read |
| Pomodoro statistics and timeline | `GET /pomodoros/statistics/generalForDesktop`, `GET /pomodoros/timeline` | `pomo stats`, `pomo timeline` | Stable read | Unit endpoint test and live read |
| Task Pomodoro records | `GET /pomodoros/task?projectId=...&taskId=...` | `pomo task` | Stable read | Unit endpoint test and live read |
| Habit preferences | `GET /user/preferences/habit?platform=web` | `habit preferences` | Stable read | Live read |
| Habits and sections | `GET /habits`, `GET /habitSections` | `habit list`, `habit sections` | Stable read | Live read |
| Habit check-ins | `POST /habitCheckins/query` | `habit checkins` | Stable read | Unit endpoint test and live read with empty account |
| Raw read | any GET path under base URL | `raw get` | Stable read-only | Live reads |
| Quadrant view | derived from sync | `quadrant list/view` | Stable derived command | Unit classifier test and live read |

## Explicit Non-Goals Until Verified

These surfaces are visible in payloads or product behavior but do not yet have verified command coverage:

- Column update/delete/order: exact `/batch/columnProject` endpoint is visible in the webapp bundle, but add/update/delete/order body shapes and rollback behavior are not verified.
- Legacy named column probes: read-only probes for `/project/{id}/data` and `/project/{id}/columns` returned 404 on the observed CN Web API. Use `GET /column/project/{projectId}` instead.
- Attachments/media upload: payloads expose attachment metadata, but upload and attach flow has not been mapped.
- Task activity: `GET /api/v1/task/activity/{taskId}` is visible in the webapp bundle, but returned HTTP 500 against the observed CN account/task without the exact cursor context; keep it in raw probing until a successful trace is captured.
- Comment attachments: comment CRUD is mapped, but multipart upload and attachment body flow are not exposed yet.
- Collaboration/team permission writes: read-only share metadata is mapped, but invite creation/deletion and user permission changes are not exposed until multi-user behavior and rollback paths are mapped.
- Filter writes: `/batch/filter` is visible in the webapp bundle, but create/update/delete payloads are not mapped.
- Arbitrary raw writes: intentionally unavailable for safety. Add a first-class command and tests instead.

## Safety Policy

- All writes support `--dry-run`.
- Delete and merge operations require `--yes`.
- Normal create/update/move operations execute directly when `--dry-run` is omitted, matching Lark CLI-style operator ergonomics.
- Agent docs and the repo skill instruct agents to preview generated writes first.
- The CLI should reject unsupported flags instead of silently ignoring them.

## Adding New Coverage

Before adding another Web API write command:

1. Capture the endpoint, method, request body, response shape, and rollback path.
2. Add a request-shape unit test.
3. Add a CLI dry-run test.
4. Run a reversible live smoke test when possible.
5. Update this matrix, `docs/web-api.md`, `docs/commands.md`, and `skills/dida-cli/SKILL.md`.
