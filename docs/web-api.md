# Web API Notes

DidaCLI uses the Dida365 Web API because it exposes a broader account surface than the public Open API. The integration is intended for the operator's own account and should be treated as private API compatibility work.

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
| `GET` | `/project/all/completed?...` | Completed tasks |

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

## Write Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/batch/task` | Create, update, complete, delete tasks |

Task operation shapes:

```json
{"add":[{"id":"...","projectId":"...","title":"..."}]}
{"update":[{"id":"...","projectId":"...","title":"..."}]}
{"update":[{"id":"...","projectId":"...","status":2}]}
{"delete":[{"taskId":"...","projectId":"..."}]}
```

## Compatibility Rules

- Normalize known fields and ignore unknown fields.
- Keep raw response access behind `raw get`.
- Do not commit full live payload dumps.
- Add tests for request shapes before adding new write commands.
- Prefer first-class resource commands over exposing arbitrary raw writes.
