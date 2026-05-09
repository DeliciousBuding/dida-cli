# Official Channel Validation Matrix

This matrix separates the two official channels:

- `official mcp`: token-based MCP server accessed with `DIDA365_TOKEN` or
  saved local official token config.
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
| Project list | `list_projects` | `official project list` | live-verified | Live smoke succeeded on 2026-05-10; promoted after evidence because it is the official-channel project discovery read. |
| Habit list | `list_habits` | `official habit list` | live-verified | Live smoke succeeded on 2026-05-10; current account returned an empty list. |
| Habit sections | `list_habit_sections` | `official habit sections` | live-verified | Live smoke succeeded on 2026-05-10; output was summarized by count only. |
| Habit read | `get_habit` | `official habit get` | blocked | Command exists, but the 2026-05-10 token smoke found zero habits on the current account, so no safe known habit id is available for live read smoke. |
| Habit create | `create_habit` | `official habit create` | implemented with dry-run | Local dry-run works without a token. Live smoke needs a reversible live habit create/update plan; no delete wrapper exists yet. |
| Habit update | `update_habit` | `official habit update` | implemented with dry-run | Local dry-run works without a token. Live update requires a known test habit. Current account has no habits. |
| Habit check-in | `upsert_habit_checkins` | `official habit checkin` | implemented with dry-run | Local dry-run works without a token. Live write requires a known test habit and reversible check-in date. Current account has no habits. |
| Habit check-ins | `get_habit_checkins` | `official habit checkins` | implemented | Requires known habit ids and a bounded date stamp range. Current account has no habits. |
| Focus read | `get_focus` | `official focus get --type 0|1` | blocked | Command exists, but 2026-05-10 token smokes found no type 0 or type 1 focus records in the current account, including a 365-day range. |
| Focus range | `get_focuses_by_time` | `official focus list --from-time ... --to-time ... --type 0|1` | live-verified | Bounded type 0 and type 1 range smokes succeeded on 2026-05-10; repeat 365-day smokes also succeeded and returned empty lists. |
| Focus delete | `delete_focus` | `official focus delete --type 0|1 --yes` | implemented with dry-run | Local dry-run works without a token. Destructive live delete requires `--yes` and a disposable focus record. |
| Project detail | `get_project_by_id` | `official project get` | live-verified | Live smoke succeeded on 2026-05-10 using a project id discovered through `list_projects`; private project data was not committed. |
| Project data | `get_project_with_undone_tasks` | `official project data` | live-verified | Live smoke succeeded on 2026-05-10 using a project id discovered through `list_projects`; private project data was not committed. |
| Time-query task read | `list_undone_tasks_by_time_query` | `official task query` | live-verified | `today` query succeeded on 2026-05-10 and returned an empty array for the current account. |
| Task detail | `get_task_by_id`, `get_task_in_project` | `official task get` | live-verified | Project-scoped task detail smoke succeeded on 2026-05-10; private task payload was not committed. |
| Batch task add | `batch_add_tasks` | `official task batch-add` | live-verified | Local `--dry-run` preview works without token; 2026-05-10 live smoke created a disposable task and cleaned it up after verification. |
| Batch task update | `batch_update_tasks` | `official task batch-update` | live-verified | Local `--dry-run` preview works without token; 2026-05-10 live smoke updated a disposable task title/content/priority, verified it through `official task get`, and cleaned it up. |
| Complete tasks in project | `complete_tasks_in_project` | `official task complete-project` | live-verified | Local `--dry-run` preview works without token; 2026-05-10 live smoke completed the disposable task created through `official task batch-add`. |
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
2. Create or identify disposable MCP habit/focus records before known-id read
   or destructive smoke tests.
3. Live-smoke remaining MCP habit/focus writes only after disposable targets exist.
4. Live-smoke OpenAPI project/task/focus/habit wrappers once OAuth token
   persistence is proven on the current account.
