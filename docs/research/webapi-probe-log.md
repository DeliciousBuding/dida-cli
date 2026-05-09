# Web API Probe Log

This log records private Web API probes that should guide future
implementation. It is deliberately curated: do not commit raw cookies, full
payload dumps, or local browser exports here.

## Confirmed Working Surfaces

| Surface | Endpoint | Status | Notes |
| --- | --- | --- | --- |
| Full sync | `GET /batch/check/0` | working | Backbone for projects, tags, filters, tasks, and checkpoints. |
| Agent context | `GET /batch/check/0` derived view | working | Live-smoked on 2026-05-10; compact output returned projects, active tasks, today/upcoming, and quadrants without needing extra sync calls. |
| Incremental sync | `GET /batch/check/{checkpoint}` | working | Used by `sync checkpoint`. |
| Project tasks | `GET /project/{projectId}/tasks` | working | Used by `project tasks`. |
| Columns read | `GET /column/project/{projectId}` | working | Used by `column list` and `project columns`. |
| Task writes | `POST /batch/task` | working | Used by create/update/complete/delete with dry-run support. |
| Project writes | `POST /batch/project` | working | Used by project CRUD with confirmation for delete. |
| Folder writes | `POST /batch/projectGroup` | working | Used by folder CRUD. |
| Tag writes | `POST /batch/tag`, `PUT /tag/rename`, `PUT /tag/merge`, `DELETE /tag` | working with quirks | Merge may leave the source tag object present. |
| Comments | comment list/create/update/delete paths | working | Attachment fields are not exposed yet. A 2026-05-10 live read against an existing task returned an empty comment list with the expected envelope. |
| Completed history | `GET /project/all/completed` | working | Date-only inputs caused server errors; full datetime strings work. |
| Closed history | `GET /project/{projectIds|all}/closed` | working | Uses full datetime strings and status filters. |
| Trash pagination | `GET /project/all/trash/page?from={cursor}` | working | Live-smoked on 2026-05-10. First page returns `next=20`; `from=20` returns the next page and `next=40`. `type=task` returned HTTP 500 and should not be sent. |
| Search | Web indexed search endpoint | working | Compact mode avoids large content blobs by default. |
| Attachment quota | legacy v1 attachment quota endpoints | working | 2026-05-10 live read returned `underQuota=true` and a numeric daily limit. Upload flow still unmapped. |

## Failed Or Incomplete Probes

| Surface | Endpoint | Observed result | Current interpretation | Next evidence needed |
| --- | --- | --- | --- | --- |
| Task activity detail | v1 `GET /task/activity/{taskId}` with optional `skip` and `lastId` | 2026-05-10 raw CLI probes against a real task returned HTTP 500 with `errorCode=need_pro` in the raw probe body snippet for no query, `skip=0`, and `skip=0&lastId=`; v2-style route returned 404 | Bundle evidence still points to the legacy v1 client. The route is reachable, but the observed account lacks the entitlement required to learn the success shape. | Pro account live read or browser trace that captures response fields and pagination semantics. |
| Project data by id | `GET /project/{id}/data` | HTTP 404 on observed CN Web API | Not the active private endpoint for current web app. | Recheck only if bundle or network trace changes. |
| Project columns by id | `GET /project/{id}/columns` | HTTP 404 | Replaced by `/column/project/{projectId}`. | None unless webapp changes. |
| Project direct get | `GET /project/{id}` | HTTP 405 | Method/path mismatch for private Web API. | Prefer sync or official channels. |
| Column project batch | `POST /batch/columnProject` | visible but not mapped | Update/delete/reorder bodies unknown. | Capture real kanban column edit/delete/reorder traffic. |
| Filter batch | `POST /batch/filter` | visible but not mapped | Create/update/delete bodies unknown. | Capture real filter edit traffic. |
| Attachment upload | multipart/upload chain | not mapped | Requires multi-step upload and association flow. | Capture upload, attach, and download/reference requests. |
| Collaboration writes | invite and permission mutation paths | not mapped | Multi-user side effects and rollback unclear. | Trace with disposable project/user setup. |

## Implementation Rule

Working reads can become commands once output shape and pagination are clear.
Writes require a request-body builder test, dry-run or preview where practical,
and a rollback note. A direct `500` probe is evidence to keep researching, not
evidence to ship a command.
