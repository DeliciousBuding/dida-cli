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
| Focus range | `get_focuses_by_time` | `official focus list` | live-verified | Bounded range smoke succeeded on 2026-05-10; output was treated as summary-only. |
| Focus delete | `delete_focus` | `official focus delete --yes` | implemented | Destructive; only test on a disposable focus record. |
| Project detail | `get_project_by_id` | `official project get` | implemented | `official project data` was live-smoked with a known project; run `project get` when exact metadata-only verification is needed. |
| Project data | `get_project_with_undone_tasks` | `official project data` | live-verified | Live smoke succeeded on 2026-05-10 using a project id discovered through `list_projects`; private project data was not committed. |
| Time-query task read | `list_undone_tasks_by_time_query` | `official task query` | live-verified | `today` query succeeded on 2026-05-10 and returned an empty array for the current account. |
| Batch task add | `batch_add_tasks` | `official task batch-add` | implemented | Local `--dry-run` preview works without token; live write needs disposable targets and schema confirmation. |
| Batch task update | `batch_update_tasks` | `official task batch-update` | implemented | Local `--dry-run` preview works without token; live write needs disposable targets and schema confirmation. |
| Complete tasks in project | `complete_tasks_in_project` | `official task complete-project` | implemented | Local `--dry-run` preview works without token; live write needs known disposable tasks. |
| Task search | `search_task` | `official task search` | live-verified | No-result query smoke succeeded on 2026-05-10. |
| Task undone by date | `list_undone_tasks_by_date` | `official task undone` | live-verified | Same-day bounded range smoke succeeded on 2026-05-10. |
| Task filtering | `filter_tasks` | `official task filter` | live-verified | `--status 0` smoke succeeded on 2026-05-10; output was summarized by count only. |

## Official OpenAPI

| Area | Endpoint / action | DidaCLI surface | Status | Evidence / next action |
| --- | --- | --- | --- | --- |
| Client config | local OAuth app credentials | `openapi client status/set/clear` | implemented | Stores client id/secret locally for OAuth commands; secret is accepted through stdin and not printed. |
| OAuth URL | `GET https://dida365.com/oauth/authorize` | `openapi auth-url` | implemented | Generates authorization URL from env or saved client credentials. |
| Callback listener | local HTTP callback | `openapi listen-callback` | implemented | Needs end-to-end browser approval smoke. |
| Token exchange | `POST https://dida365.com/oauth/token` | `openapi exchange-code` | implemented | Invalid-code test confirmed client authentication path; real code still needed. |
| Interactive login | OAuth URL + listener + token exchange | `openapi login` | implemented | Needs full live OAuth approval and persisted token verification. |
| Project list | `GET /open/v1/project` | `openapi project list` | implemented | Requires saved OAuth access token for final live smoke. |
| Project get/data | `GET /open/v1/project/{id}`, `/data` | `openapi project get/data` | implemented | Requires saved OAuth access token for final live smoke. |
| Task CRUD | `/open/v1/task...` | `openapi task get/create/update/complete/delete/move` | implemented | Requires saved OAuth access token and disposable live task for write smoke. |
| Task query | `/open/v1/task/completed`, `/open/v1/task/filter` | `openapi task completed/filter` | implemented | Requires saved OAuth access token for read smoke. |
| Focus | `/open/v1/focus...` | `openapi focus get/list/delete` | implemented | Requires saved OAuth access token for read smoke; delete only a disposable record. |
| Habit | `/open/v1/habit...` | `openapi habit list/get/create/update/checkin/checkins` | implemented | Requires saved OAuth access token; writes need disposable habit/check-in data. |

## Current Priority

1. Live-verify `openapi login` with a real authorization code.
2. Live-smoke known-id MCP habit/focus reads when safe IDs exist.
3. Live-smoke MCP writes only after disposable project/task/habit targets exist.
4. Live-smoke OpenAPI project/task/focus/habit wrappers once OAuth token
   persistence is proven on the current account.
