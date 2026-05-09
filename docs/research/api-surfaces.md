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
- `GET /column/project/{projectId}` for named kanban columns with `id`, `name`, `sortOrder`, `etag`, and timestamps.
- `GET /project/all/completed?from=YYYY-MM-DD%20HH%3Amm%3Ass&to=YYYY-MM-DD%20HH%3Amm%3Ass&limit=N` for completed task queries. Date-only `from/to` produced server 500 in the observed CN Web API; full timestamp format worked.
- `POST /batch/task` for task operations.
- `POST /batch/taskParent` for subtask operations.
- `POST /batch/taskProject` for task moves.
- `POST /batch/project` for project operations.
- `POST /batch/projectGroup` for folder/group operations.
- `POST /batch/tag` and tag-specific endpoints for tag operations.
- Sync payload `filters` for read-only custom filter discovery.
- `POST /column` for column creation experiments.
- `GET /project/{projectId}/task/{taskId}/comments` for task comments.
- `POST /project/{projectId}/task/{taskId}/comment` with `{"title":"..."}` for comment create.
- `PUT /project/{projectId}/task/{taskId}/comment/{commentId}` with `{"title":"..."}` for comment update.
- `DELETE /project/{projectId}/task/{taskId}/comment/{commentId}` for comment delete.
- `POST /batch/columnProject` is present in the webapp bundle, but exact update/delete/order payload shapes are not yet verified.
- `POST /batch/filter` is present in the webapp bundle, but create/update/delete payload shapes are not yet verified.

Column probe notes:

- `POST /column` with `{projectId,name}` returned `{"id2error":{},"id2etag":{}}` for temporary projects, but `GET /column/project/{projectId}` still returned an empty list after a short delay.
- The same result occurred for a temporary project created with `viewMode:"kanban"`.
- Because the create response did not expose a column id and no column appeared in the read endpoint, column update/delete/order must stay unimplemented until a real webapp network trace shows the full project/column preconditions and payload.

Observed negative probes on the CN Web API:

- `GET /project/{projectId}/data` returned 404.
- `GET /project/{projectId}/columns` returned 404.
- `GET /project/{projectId}` returned 405.

The working project task read endpoint remains `GET /project/{projectId}/tasks`.

### Client Layers

- `webapi.Client`: HTTP transport, auth headers, endpoint path construction, error decoding.
- `internal/webapi/sync.go`: full sync, settings, and completed task queries.
- `internal/webapi/tasks.go`: task create/update/complete/delete.
- `internal/webapi/resources.go`: task move/subtask, project CRUD, folder CRUD, tag CRUD, column list, and experimental column create.
- `internal/webapi/comments.go`: task comment list/create/update/delete without attachments.
- `internal/cli/*.go`: stable command envelope, dry-run previews, and destructive confirmation gates.

The old Doris/OpenClaw TypeScript implementation may be kept locally under ignored `data/private/reference/dida365-ai-tools/` for comparison. Do not commit that reference tree or copied live payloads.

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
- Client id and client secret can come from environment variables or `dida
  openapi client set --id <client-id> --secret-stdin`.
- Access token and refresh token are stored by `dida openapi login`.
- This channel is separate from Web API cookie auth and official MCP
  `DIDA365_TOKEN` auth.

Implemented command families:

- `dida openapi doctor`
- `dida openapi status`
- `dida openapi auth-url`
- `dida openapi listen-callback`
- `dida openapi exchange-code`
- `dida openapi login`
- `dida openapi project list/get/data`
- `dida openapi task get/create/update/complete/delete/move/completed/filter`
- `dida openapi focus get/list/delete`
- `dida openapi habit list/get/create/update/checkin/checkins`

Current blocker:

- Full live OAuth verification still needs a configured developer app redirect
  URL and real browser approval on the current account.

## Safety Rules

- Do not print cookies or bearer tokens.
- Write commands must support `--dry-run` before live writes.
- Raw escape hatch starts as read-only GET.
- Capture redacted fixtures before implementing broad parsers.
