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
| Project task list | `GET /project/{projectId}/tasks` | `project tasks <project-id>` | Stable | Unit endpoint test and live read |
| Project columns | inferred from task `columnId` | `project columns`, `column list` | Stable read | Live read; names unavailable from observed sync |
| Task CRUD | `POST /batch/task` | `task create/update/complete/delete` | Stable | Unit request tests and live reversible smoke |
| Task advanced fields | `/batch/task` fields | `--content`, `--desc`, `--start`, `--due`, `--timezone`, `--tag`, `--tags`, `--item`, `--column`, `--reminder`, `--repeat*`, `--all-day`, `--floating`, `--priority 0` | Stable | Unit request tests and live reversible smoke |
| Task move | `POST /batch/taskProject` | `task move` | Stable | Unit request tests and live reversible smoke |
| Task parent/subtask | `POST /batch/taskParent` | `task parent` | Stable | Unit request tests and live reversible smoke |
| Project CRUD | `POST /batch/project` | `project create/update/delete` | Stable | Unit request tests and live reversible smoke |
| Folder CRUD | `POST /batch/projectGroup` | `folder create/update/delete` | Stable | Unit request tests and live reversible smoke |
| Tag create/update | `POST /batch/tag` | `tag create/update` | Stable | Unit request tests and live reversible smoke |
| Tag rename | `PUT /tag/rename` | `tag rename` | Stable | Unit request tests and live reversible smoke |
| Tag merge | `PUT /tag/merge` | `tag merge --yes` | Stable with caveat | Unit request tests and live smoke; source tag may remain |
| Tag delete | `DELETE /tag?name=...` | `tag delete --yes` | Stable | Unit URL escaping test and live reversible smoke |
| Column create | `POST /column` | `column create` | Experimental | Unit request test; live write avoided because delete endpoint is unknown |
| Raw read | any GET path under base URL | `raw get` | Stable read-only | Live reads |
| Quadrant view | derived from sync | `quadrant list/view` | Stable derived command | Unit classifier test and live read |

## Explicit Non-Goals Until Verified

These surfaces are visible in payloads or product behavior but do not yet have verified command coverage:

- Column update/delete: endpoint and rollback behavior not verified.
- Named column list: read-only probes for `/project/{id}/data` and `/project/{id}/columns` returned 404 on the observed CN Web API; current column read is inferred from task `columnId`.
- Attachments/media upload: payloads expose attachment metadata, but upload and attach flow has not been mapped.
- Comments: payloads expose `commentCount`, but comment endpoints are not mapped.
- Collaboration/team permissions: project payloads expose team and permission fields, but multi-user behavior is not mapped.
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
