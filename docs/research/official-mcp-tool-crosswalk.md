# Official MCP Tool Crosswalk

Live snapshot: **57 tools**, verified against `https://mcp.dida365.com` on
2026-10-23 with `dida official tools --full --json`.

The live `tools/list` response is the authority. This page records the current
surface and the intended command strategy; it is not a hand-maintained schema.

## Coverage Strategy

DidaCLI uses two layers instead of cloning every upstream tool into a parallel
command tree:

1. **First-class wrappers** for stable, high-value reads and writes with local
   validation and `--dry-run` previews.
2. **`official call <tool>`** for newly published or low-frequency tools.
   Preview with `dida official call <tool> --args-json '...' --dry-run --json`
   before executing. A fixed read-only tool allowlist runs directly; writes and
   unrecognized future tools require `--yes`.

The live server advertised MCP `2026-07-28` stateless discovery and per-request
metadata. DidaCLI negotiates that protocol first and retains the legacy
initialize handshake only as a fallback.

The 2026-10-23 response included `outputSchema` and annotations for all 57
tools. No current input schema used `x-mcp-header`; the client still treats
annotations as untrusted discovery hints rather than authorization metadata.

The live catalog can be inspected with:

```bash
dida official doctor --json
dida official tools --limit 100 --json
dida official show <tool-name> --json
```

## Projects And Organization

| Official tool | DidaCLI surface | Status |
| --- | --- | --- |
| `list_projects` | `official project list` | first-class |
| `get_project_by_id` | `official project get` | first-class |
| `get_project_with_undone_tasks` | `official project data` | first-class |
| `list_project_members` | `official call list_project_members` | generic |
| `create_project` | `official call create_project` | generic; Web API wrapper also exists |
| `update_project` | `official call update_project` | generic; Web API wrapper also exists |
| `list_project_groups` | `official call list_project_groups` | generic |
| `create_project_group` | `official call create_project_group` | generic |
| `update_project_group` | `official call update_project_group` | generic |
| `delete_project_group` | `official call delete_project_group --yes` | generic; destructive |
| `list_columns` | `official call list_columns` | generic |
| `create_column` | `official call create_column` | generic |
| `update_column` | `official call update_column` | generic |

## Tasks

| Official tool | DidaCLI surface | Status |
| --- | --- | --- |
| `get_task_by_id` | `official task get` | first-class |
| `get_task_in_project` | `official task get --project <id>` | first-class |
| `search_task` | `official task search` | first-class |
| `list_undone_tasks_by_time_query` | `official task query` | first-class |
| `list_undone_tasks_by_date` | `official task undone` | first-class |
| `filter_tasks` | `official task filter` | first-class |
| `batch_add_tasks` | `official task batch-add` | first-class, dry-run |
| `batch_update_tasks` | `official task batch-update` | first-class, dry-run |
| `complete_tasks_in_project` | `official task complete-project` | first-class, dry-run |
| `create_task` | `official call create_task` | generic; Web API wrapper also exists |
| `update_task` | `official call update_task` | generic; Web API wrapper also exists |
| `complete_task` | `official call complete_task` | generic |
| `complete_checklist_item` | `official call complete_checklist_item` | generic |
| `delete_task` | `official call delete_task --yes` | generic; destructive |
| `move_task` | `official call move_task` | generic |
| `assign_task` | `official call assign_task` | generic |
| `unassign_task` | `official call unassign_task` | generic |
| `list_completed_tasks_by_date` | `official call list_completed_tasks_by_date` | generic |
| `search` | `official call search` | generic, complement to `search_task` |
| `fetch` | `official call fetch` | generic, ID-based full task fetch |

## Comments, Tags, Countdowns, And Calendar

| Official tool | DidaCLI surface | Status |
| --- | --- | --- |
| `get_comment` | `official call get_comment` | generic |
| `add_comment` | `official call add_comment` | generic; write-capable |
| `delete_comment` | `official call delete_comment --yes` | generic; destructive |
| `list_tags` | `official call list_tags` | generic |
| `create_tag` | `official call create_tag` | generic |
| `rename_tag` | `official call rename_tag` | generic |
| `update_tag` | `official call update_tag` | generic |
| `delete_tag` | `official call delete_tag --yes` | generic; destructive |
| `list_countdowns` | `official call list_countdowns` | generic |
| `list_calendar_accounts` | `official call list_calendar_accounts` | generic |
| `query_calendar_events` | `official call query_calendar_events` | generic; bounded read |

## Habits, Focus, And Preferences

| Official tool | DidaCLI surface | Status |
| --- | --- | --- |
| `list_habits` | `official habit list` | first-class |
| `list_habit_sections` | `official habit sections` | first-class |
| `get_habit` | `official habit get` | first-class |
| `create_habit` | `official habit create` | first-class, dry-run |
| `update_habit` | `official habit update` | first-class, dry-run |
| `upsert_habit_checkins` | `official habit checkin` | first-class, dry-run |
| `get_habit_checkins` | `official habit checkins` | first-class |
| `get_focus` | `official focus get` | first-class |
| `get_focuses_by_time` | `official focus list` | first-class |
| `delete_focus` | `official focus delete --yes` | first-class, destructive |
| `create_focus` | `official call create_focus` | generic; required `startTime`, `endTime`, `type` |
| `update_focus` | `official call update_focus` | generic; type required |
| `get_user_preference` | `official call get_user_preference` | generic |

## What Changed In The 2026-10-23 Snapshot

- The official catalog now exposes task assignment, checklist completion,
  comments, tag mutations, project groups/columns, countdowns, calendar access,
  and focus creation/update.
- The former claim that comments, tags, and folders are Web-API-only is stale.
- The catalog remains narrower than the private Web API for sharing, templates,
  account sessions, broad settings, and some calendar metadata.

## Tool Contract Rules

- Use the live catalog and schemas, not an embedded copy, for discovery.
- Treat the MCP `inputSchema` as the argument authority. Current MCP schemas use
  JSON Schema 2020-12 with an object root.
- Treat annotations as hints; host approval and DidaCLI confirmation remain the
  authorization boundary.
- Do not expose a generic write without a dry-run preview or explicit `--yes`.
- Preserve deterministic tool ordering and do not rewrite upstream results.
