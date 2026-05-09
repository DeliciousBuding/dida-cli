# Web API Notes

DidaCLI uses the Dida365 Web API because it exposes a broader account surface than the public Open API. The integration is intended for the operator's own account and should be treated as private API compatibility work.

For command-by-command implementation status, see [api-coverage.md](api-coverage.md).

## Base

```text
https://api.dida365.com/api/v2
```

Required headers:

```text
Cookie: t=<browser-cookie>
User-Agent: browser-like user agent
x-device: browser-like Dida device descriptor
```

## Read Endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/batch/check/0` | Full sync |
| `GET` | `/batch/check/{checkpoint}` | Incremental sync |
| `GET` | `/user/preferences/settings` | User settings |
| `GET` | `/project/{projectId}/tasks` | Project task list |
| `GET` | `/column/project/{projectId}` | Kanban column list with names and sort order |
| `GET` | `/project/all/completed?...` | Completed tasks |
| `GET` | `/user/preferences/pomodoro` | Pomodoro preferences |
| `GET` | `/pomodoros?from={millis}&to={millis}` | Pomodoro records |
| `GET` | `/pomodoros/timing?from={millis}&to={millis}` | Pomodoro timing records |
| `GET` | `/user/preferences/habit?platform=web` | Habit preferences |
| `GET` | `/habits` | Habits |
| `GET` | `/habitSections` | Habit sections |

Observed CN full sync shape:

```text
projectProfiles
syncTaskBean.add
syncTaskBean.update
syncTaskBean.delete
projectGroups
tags
checkPoint
checks
filters
```

Completed query timestamps should use full datetime values:

```text
from=YYYY-MM-DD HH:mm:ss
to=YYYY-MM-DD HH:mm:ss
```

Date-only `from/to` values have produced HTTP 500 on the observed CN Web API.

Pomodoro range endpoints expect millisecond timestamps, not formatted datetime
strings. `dida pomo list` and `dida pomo timing` accept `YYYY-MM-DD` and convert
to the required millisecond range.

## Write Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/batch/task` | Create, update, complete, delete tasks |
| `POST` | `/batch/taskProject` | Move tasks between projects |
| `POST` | `/batch/taskParent` | Set task parent/subtask relationship |
| `POST` | `/batch/project` | Create, update, delete projects |
| `POST` | `/batch/projectGroup` | Create, update, delete project folders |
| `POST` | `/batch/tag` | Create and update tags |
| `PUT` | `/tag/rename` | Rename a tag |
| `PUT` | `/tag/merge` | Merge one tag into another |
| `DELETE` | `/tag?name=...` | Delete a tag |
| `POST` | `/column` | Create a kanban column; experimental |
| `POST` | `/project/{projectId}/task/{taskId}/comment` | Create task comment |
| `PUT` | `/project/{projectId}/task/{taskId}/comment/{commentId}` | Update task comment |
| `DELETE` | `/project/{projectId}/task/{taskId}/comment/{commentId}` | Delete task comment |

Task operation shapes:

```json
{"add":[{"id":"...","projectId":"...","title":"..."}]}
{"update":[{"id":"...","projectId":"...","title":"..."}]}
{"update":[{"id":"...","projectId":"...","status":2}]}
{"delete":[{"taskId":"...","projectId":"..."}]}
```

Task create/update exposes the observed Web API fields `content`, `desc`, `allDay`, `startDate`, `dueDate`, `timeZone`, `reminders`, `repeat`, `repeatFrom`, `repeatFlag`, `priority`, `columnId`, `tags`, `items`, and `isFloating`.

`priority` is represented internally as an optional field so `--priority 0` is sent explicitly and can clear an existing priority.

Incremental sync preserves `syncTaskBean.add`, `syncTaskBean.update`, `syncTaskBean.delete`, `syncOrderBean`, `syncTaskOrderBean`, and observed reminder delta containers in `sync checkpoint --json`.

Resource operation shapes:

```json
{"add":[{"id":"...","name":"...","viewMode":"list","kind":"TASK"}]}
{"update":[{"id":"...","name":"..."}]}
{"delete":["project-or-folder-id"]}
{"add":[{"name":"tag-name","color":"#147d4f"}]}
{"name":"old-tag","newName":"new-tag"}
{"from":"old-tag","to":"new-tag"}
{"projectId":"...","name":"Doing"}
```

Comment operation shape:

```json
{"id":"...","createdTime":"YYYY-MM-DDTHH:mm:ss.000+0000","taskId":"...","projectId":"...","title":"comment text","userProfile":{"isMyself":true},"mentions":[],"isNew":true}
```

The webapp client generates `id`, `createdTime`, `taskId`, `projectId`, `userProfile`, empty `mentions`, and `isNew` before sending a create request. The bundle also shows optional reply and attachment fields, but DidaCLI only exposes plain text comments until the attachment upload flow is verified.

Task relationship shapes:

```json
[{"taskId":"...","fromProjectId":"...","toProjectId":"..."}]
[{"taskId":"...","parentId":"...","projectId":"..."}]
```

Column update/delete/order endpoints are not yet documented in this CLI. The webapp bundle references `POST /batch/columnProject`, but DidaCLI should not expose first-class update/delete/order commands until payload shapes are observed and covered by tests.

Observed tag merge behavior: the endpoint can return success while the source tag remains listed. Treat merge and delete as separate operations.

## Compatibility Rules

- Normalize known fields and ignore unknown fields.
- Keep raw response access behind `raw get`.
- Do not commit full live payload dumps.
- Add tests for request shapes before adding new write commands.
- Prefer first-class resource commands over exposing arbitrary raw writes.
