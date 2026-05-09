# Official Channel Validation Matrix

This matrix separates the two official channels:

- `official mcp`: token-based MCP server accessed with `DIDA365_TOKEN`.
- `official openapi`: OAuth REST API under `/open/v1`.

They are intentionally not interchangeable.

## Status Legend

| Status | Meaning |
| --- | --- |
| `live-verified` | A real request succeeded against the current account. |
| `implemented` | Command exists and tests pass, but live verification may need a safe target. |
| `documented` | Official docs or schemas exist, but DidaCLI has not wrapped it yet. |
| `blocked` | Requires user OAuth approval, a safe write target, or missing request evidence. |
| `deferred` | Intentionally not wrapped because generic tooling is better for now. |

## Official MCP

| Area | Upstream tool or action | DidaCLI surface | Status | Evidence / next action |
| --- | --- | --- | --- | --- |
| Health | `initialize`, `tools/list` | `official doctor` | live-verified | Tool count and sample tools were verified through the MCP server. |
| Discovery | `tools/list` | `official tools` | live-verified | Compact output is default; `--full` preserves raw schemas. |
| Schema | `tools/list` by name | `official show <tool>` | live-verified | Used to inspect tool contracts before wrapping. |
| Generic call | `tools/call` | `official call <tool>` | live-verified | Works for read tools such as project list and time-query tasks. |
| Habit read | `get_habit` | `official habit get` | implemented | Needs a known habit id for live read smoke. |
| Habit create | `create_habit` | `official habit create` | implemented | Needs a reversible live habit create/update/delete plan; no delete wrapper exists yet. |
| Habit update | `update_habit` | `official habit update` | implemented | Requires a known test habit. |
| Habit check-in | `upsert_habit_checkins` | `official habit checkin` | implemented | Requires a known test habit and reversible check-in date. |
| Focus read | `get_focus` | `official focus get` | implemented | Needs a known focus id. |
| Focus range | `get_focuses_by_time` | `official focus list` | implemented | Safe read; should be live-smoked with a bounded range. |
| Focus delete | `delete_focus` | `official focus delete --yes` | implemented | Destructive; only test on a disposable focus record. |
| Project detail | `get_project_by_id` | `official project get` | implemented | Needs a known project id for live read smoke. |
| Project data | `get_project_with_undone_tasks` | `official project data` | implemented | Needs a known project id for live read smoke. |
| Batch task add | `batch_add_tasks` | none | documented | Needs wrapper design with clear payload file support. |
| Batch task update | `batch_update_tasks` | none | documented | Needs dry-run-like local request preview. |
| Task filtering | `filter_tasks` | none | documented | Good candidate for compact output. |

## Official OpenAPI

| Area | Endpoint / action | DidaCLI surface | Status | Evidence / next action |
| --- | --- | --- | --- | --- |
| OAuth URL | `GET https://dida365.com/oauth/authorize` | `openapi auth-url` | implemented | Generates authorization URL from env-provided client credentials. |
| Callback listener | local HTTP callback | `openapi listen-callback` | implemented | Needs end-to-end browser approval smoke. |
| Token exchange | `POST https://dida365.com/oauth/token` | `openapi exchange-code` | implemented | Invalid-code test confirmed client authentication path; real code still needed. |
| Interactive login | OAuth URL + listener + token exchange | `openapi login` | implemented | Needs full live OAuth approval and persisted token verification. |
| Project list | `GET /open/v1/project` | `openapi project list` | implemented | Requires saved OAuth access token for final live smoke. |
| Project get/data | `GET /open/v1/project/{id}`, `/data` | none | documented | Should follow after OAuth is live. |
| Task CRUD | `/open/v1/task...` | none | documented | Implement after OAuth read path is verified. |
| Focus | `/open/v1/focus...` | none | documented | Implement after OAuth read path is verified. |
| Habit | `/open/v1/habit...` | none | documented | Implement after OAuth read path is verified. |

## Current Priority

1. Live-verify `openapi login` with a real authorization code.
2. Live-smoke `official focus list` because it is a bounded read.
3. Live-smoke MCP project wrappers because they improve agent context without
   private Web API risk.
4. Keep OpenAPI task/focus/habit wrappers blocked until OAuth token persistence
   is proven on the current account.
