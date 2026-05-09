# Official MCP Tool Crosswalk

This file maps the official dida365 MCP tools to the current DidaCLI command
surface and highlights which official tools are most valuable to wrap as
first-class commands.

## Project Tools

| Official MCP tool | Current DidaCLI equivalent | Assessment |
| --- | --- | --- |
| `list_projects` | `project list` | Overlap; official is cleaner but not broader |
| `get_project_by_id` | no exact first-class equivalent | Worth wrapping |
| `get_project_with_undone_tasks` | `project tasks` plus extra reads | Worth wrapping as a bundle read |
| `create_project` | `project create` | Overlap |
| `update_project` | `project update` | Overlap |

## Task Write Tools

| Official MCP tool | Current DidaCLI equivalent | Assessment |
| --- | --- | --- |
| `create_task` | `task create` | Overlap |
| `update_task` | `task update` | Overlap |
| `complete_task` | `task complete` | Overlap |
| `complete_tasks_in_project` | none | Worth wrapping |
| `move_task` | `task move` | Overlap |
| `batch_add_tasks` | none | Worth wrapping |
| `batch_update_tasks` | none | Worth wrapping |

## Task Read / Query Tools

| Official MCP tool | Current DidaCLI equivalent | Assessment |
| --- | --- | --- |
| `get_task_in_project` | `task get` | Overlap; official is stricter |
| `get_task_by_id` | `task get` | Overlap |
| `fetch` | `task get` or `raw get` depending on use | Useful but overlapping |
| `search` | `search all` | Overlap; official may be narrower and cleaner |
| `search_task` | `task search` / `search all` | Useful candidate |
| `filter_tasks` | no exact equivalent | Worth wrapping |
| `list_undone_tasks_by_date` | no exact structured equivalent | Worth wrapping |
| `list_undone_tasks_by_time_query` | `task today`, `task list --filter ...` | Worth wrapping behind a friendlier query command |
| `list_completed_tasks_by_date` | `completed list` | Overlap |

## Habit Tools

| Official MCP tool | Current DidaCLI equivalent | Assessment |
| --- | --- | --- |
| `list_habits` | `habit list` | Overlap |
| `list_habit_sections` | `habit sections` | Overlap |
| `get_habit` | `official habit get` | Wrapped |
| `create_habit` | `official habit create` | Wrapped |
| `update_habit` | `official habit update` | Wrapped |
| `upsert_habit_checkins` | `official habit checkin` | Wrapped for one check-in at a time |
| `get_habit_checkins` | `habit checkins` | Overlap |

## Focus Tools

| Official MCP tool | Current DidaCLI equivalent | Assessment |
| --- | --- | --- |
| `get_focus` | `official focus get` | Wrapped |
| `get_focuses_by_time` | `official focus list` | Wrapped |
| `delete_focus` | `official focus delete` | Wrapped |

## Preference Tool

| Official MCP tool | Current DidaCLI equivalent | Assessment |
| --- | --- | --- |
| `get_user_preference` | `settings get` | Overlap, but official contract is narrower and cleaner |

## Best Official MCP Candidates For First-Class Wrapping

These are the official MCP tools with the strongest justification for dedicated
commands instead of only `official call`.

1. `get_project_by_id`
2. `get_project_with_undone_tasks`
3. `complete_tasks_in_project`
4. `batch_add_tasks`
5. `batch_update_tasks`
6. `filter_tasks`
7. `list_undone_tasks_by_date`
8. `search_task`
9. `get_habit` - wrapped as `official habit get`
10. `create_habit` - wrapped as `official habit create`
11. `update_habit` - wrapped as `official habit update`
12. `upsert_habit_checkins` - wrapped as `official habit checkin`
13. `get_focus` - wrapped as `official focus get`
14. `get_focuses_by_time` - wrapped as `official focus list`
15. `delete_focus` - wrapped as `official focus delete`

## Notable Overlap Pattern

The official MCP channel is not a replacement for the Web API channel.

What it is best at:

- well-defined task and project operations
- official token-based auth
- clean typed schemas
- batch task operations
- habit write support
- focus support

What the Web API still does better:

- comments
- tags
- folders
- sharing
- calendar metadata
- templates
- web-only settings
- account/session metadata
- attachment quota and broader account surfaces

## Current Recommendation

- Keep `official call` as the generic power-user / exploration surface.
- Promote only the high-value official tools into first-class commands.
- Avoid cloning the entire official MCP catalogue into parallel command trees
  unless there is a real UX or capability gain.
