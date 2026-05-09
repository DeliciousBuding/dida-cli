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
| `get_habit` | `official habit get` | Habit detail is not covered by the Web API read surface. |
| `create_habit` | `official habit create` | Official channel supports habit writes with schema-backed payloads. |
| `update_habit` | `official habit update` | Official channel supports habit writes with schema-backed payloads. |
| `upsert_habit_checkins` | `official habit checkin` | Common action can be expressed as one safe, predictable command. |
| `get_focus` | `official focus get` | Focus detail is an official-only capability. |
| `get_focuses_by_time` | `official focus list` | Bounded read that is useful for daily review agents. |
| `delete_focus` | `official focus delete` | Officially supported destructive operation; use only with known IDs. |

## Next Promotion Candidates

1. `get_project_by_id`
2. `get_project_with_undone_tasks`
3. `filter_tasks`
4. `list_undone_tasks_by_date`
5. `complete_tasks_in_project`
6. `batch_add_tasks`
7. `batch_update_tasks`

The project and read/filter wrappers should come before batch writes because
they improve agent context without requiring rollback planning.
