# Official MCP Wrapping Policy

`dida official call` is the permanent escape hatch. First-class MCP commands
exist only when they make common agent work easier than passing raw tool names
and JSON every time.

## Promote A Tool When

- It is frequently useful for agents.
- The command name is clearer than the upstream tool name.
- The wrapper can add bounded output, validation, or safer argument ergonomics.
- The wrapper does not hide the underlying tool identity.
- The payload shape is stable from `official show <tool> --json`.

## Do Not Promote A Tool When

- It is rarely used and `official call` is adequate.
- The wrapper would be only a one-to-one alias with no usability gain.
- The write semantics are unclear.
- The command would duplicate an existing Web API command without improving
  auth, safety, batch behavior, or output quality.

## Command Design Rules

- Keep all wrappers under `dida official ...`.
- Keep `DIDA365_TOKEN` as the only auth model for this channel.
- Mention the upstream MCP tool in schema docs.
- Use `--args-json` or `--args-file` when the official payload is broad.
- Prefer narrow flags only for common simple operations.
- Keep JSON envelopes stable and compact where the upstream payload is noisy.

## Current Promotion Decisions

| Upstream tool | Wrapper | Reason |
| --- | --- | --- |
| `get_project_by_id` | `official project get` | Official project detail read avoids private Web API endpoint ambiguity. |
| `get_project_with_undone_tasks` | `official project data` | Bundles project, columns, and undone tasks for agent context. |
| `search_task` | `official task search` | Official search is a narrow read with a simple query contract. |
| `list_undone_tasks_by_time_query` | `official task query` | Gives agents the official natural time-query read without raw tool JSON. |
| `list_undone_tasks_by_date` | `official task undone` | Bounded task reads are useful for planning agents. |
| `filter_tasks` | `official task filter` | Exposes official structured filters without private Web API guessing. |
| `complete_tasks_in_project` | `official task complete-project` | Batch completion uses explicit project/task IDs and supports local dry-run preview. |
| `batch_add_tasks` | `official task batch-add` | Batch create is schema-backed but broad, so the wrapper keeps payload JSON visible. |
| `batch_update_tasks` | `official task batch-update` | Batch update is schema-backed but broad, so the wrapper keeps payload JSON visible. |
| `get_habit` | `official habit get` | Habit detail is not covered by the Web API read surface. |
| `create_habit` | `official habit create` | Official channel supports habit writes with schema-backed payloads. |
| `update_habit` | `official habit update` | Official channel supports habit writes with schema-backed payloads. |
| `upsert_habit_checkins` | `official habit checkin` | Common action can be expressed as one safe, predictable command. |
| `get_focus` | `official focus get` | Focus detail is an official-only capability. |
| `get_focuses_by_time` | `official focus list` | Bounded read that is useful for daily review agents. |
| `delete_focus` | `official focus delete` | Officially supported destructive operation; use only with known IDs. |

## Next Promotion Candidates

No remaining official MCP task/project/habit/focus tool currently justifies a
new first-class wrapper without fresh usage evidence. Keep `official call` for
rare tools and use `official show <tool-name> --json` before broad payloads.
