# Dida API Surfaces

## Strategy

DidaCLI is Web API first because the old Doris/OpenClaw integration already proved that `api/v2` exposes the account-wide sync surface needed for automation.

Official Open API remains useful, but it should not define the first CLI architecture because it does not cover enough of the product surface for task-agent workflows.

## Web Private API

Primary implementation target.

Use only for the operator's own account where official API lacks coverage.

Observed from the previous Doris setup:

- Base URL: `https://api.dida365.com/api/v2`
- Auth: `Cookie: t=<browser cookie>`
- Token setup equivalent: copy the `t` cookie from a logged-in browser session.
- Existing Node tool stored it as `~/.dida365/token.json`.

Important existing endpoints:

- `GET /batch/check/0` for full sync payload. Current Dida365 CN response uses `projectProfiles` and `syncTaskBean.add/update`; older tools may expect top-level `projects/tasks`. It also returns `checkPoint`, `checks`, `filters`, `syncOrderBean`, `syncTaskOrderBean`, and reminder changes.
- `GET /batch/check/{checkpoint}` for incremental sync. When no changes exist, most top-level resource arrays are `null` and counts should normalize to zero.
- `GET /user/preferences/settings` for settings. The CN response currently includes both `nlpEnabled` and `nlpenabled`; parse with a normal JSON object in Go and avoid case-insensitive key assumptions in shell tooling.
- `GET /project/{projectId}/tasks` for project task lists.
- `GET /project/all/completed?from=YYYY-MM-DD%20HH%3Amm%3Ass&to=YYYY-MM-DD%20HH%3Amm%3Ass&limit=N` for completed task queries. Date-only `from/to` produced server 500 in the observed CN Web API; full timestamp format worked.
- `POST /batch/task` for task operations.
- `POST /batch/taskParent` for subtask operations.
- `POST /batch/taskProject` for task moves.
- `POST /batch/project` for project operations.
- `POST /batch/projectGroup` for folder/group operations.
- `POST /batch/tag` and tag-specific endpoints for tag operations.
- `POST /column` for column creation experiments.

### Client Layers

- `webapi.Client`: HTTP transport, auth headers, endpoint path construction, error decoding.
- `webapi.SyncService`: full sync and settings.
- `webapi.CompletedService`: completed task queries.
- `webapi.TaskService`: create/update/complete/delete/move/subtask.
- `webapi.ProjectService`: project/folder/column operations.
- `webapi.TagService`: tag list/create/rename/merge/delete.
- `webapi.RawService`: GET-only endpoint probe.

### Header Compatibility

The Doris tool used browser-like private headers:

- `Cookie: t=<token>`
- `User-Agent: Mozilla/5.0 ...`
- `x-device: {"platform":"web",...}`

DidaCLI should generate the `x-device` value centrally and keep it stable enough for normal web compatibility without copying browser fingerprints from unrelated sessions.

### Data Model Policy

Private API payloads are not guaranteed stable. Parsers should:

- normalize common fields into internal models,
- keep raw JSON available under `--json --raw` later,
- ignore unknown fields by default,
- fixture-test representative sync payloads.

## Official Open API

Use for stable long-term operations where feature coverage is enough.

Expected auth:

- OAuth2 app from the Dida365/TickTick developer portal.
- Access token and refresh token stored under `~/.dida-cli/oauth.json`.

Initial commands:

- `dida auth oauth start`
- `dida project list`
- `dida task create/update/complete/delete`

## Safety Rules

- Do not print cookies or bearer tokens.
- Write commands must support `--dry-run` before live writes.
- Raw escape hatch starts as read-only GET.
- Capture redacted fixtures before implementing broad parsers.
