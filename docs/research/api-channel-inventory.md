# API Channel Inventory

This document summarizes the three API channels currently studied or implemented
in DidaCLI.

## Channel 1: Web API

- Auth: browser cookie `t`
- Base: `api.dida365.com/api/v1|v2`
- Current role: main implementation channel
- Strengths:
  - broadest coverage
  - closest to real web app behavior
  - already covers most current DidaCLI commands
- Weaknesses:
  - private and drift-prone
  - some write surfaces still require reverse-engineering

## Channel 2: Official MCP

- Auth: `DIDA365_TOKEN=dp_...`
- Base: `https://mcp.dida365.com`
- Current role: official token-based tool channel
- Strengths:
  - officially exposed
  - typed tool schemas
  - clean resource domains for task/project/habit/focus
  - batch operations and focus/habit write support
- Weaknesses:
  - smaller surface than Web API
  - known-id habit/focus live smokes depend on disposable records in the
    current account

DidaCLI keeps generic `official tools/show/call` access, but high-value
project, task, habit, and focus reads plus safe task write wrappers now have
first-class commands.

## Channel 3: Official OpenAPI

- Auth: OAuth access token
- Base: `https://api.dida365.com/open/v1`
- Current role: official OAuth REST channel for project/task/focus/habit
  resources
- Strengths:
  - documented REST contract
  - clearer public semantics than private Web API
  - first-class DidaCLI wrappers exist for OAuth setup plus project, task,
    focus, and habit resources
- Weaknesses:
  - requires OAuth authorization-code flow
  - live resource smokes require a completed browser OAuth approval and saved
    access token
  - not directly usable with the official MCP `dp_...` token
  - narrower than the private Web API

## Current Maturity

| Channel | Research | Implemented | Live verified |
| --- | --- | --- | --- |
| Web API | deep | broad | yes |
| Official MCP | moderate to deep | generic tool call plus first-class project/task/habit/focus wrappers | yes for project/task reads and task writes; blocked for known-id habit/focus targets |
| Official OpenAPI | moderate to deep | OAuth helpers plus project/task/focus/habit wrappers | local/dry-run verified; live resource calls blocked on OAuth approval |

## Agent Channel Selection

Use this table before adding commands or choosing a command in an automation
run:

| Job | Preferred channel | Reason | Fallback |
| --- | --- | --- | --- |
| One-shot account context for an Agent | Web API `agent context --json` or `--outline` | One full sync gives projects, tags, filters, today, upcoming, quadrants, and task refs without multiple calls. | Web API `sync all --json` when raw sync payload is needed. |
| Normal task reads and writes | Web API `task ...` | Best coverage for the web app's current task model, including compact reads and dry-run previews. | Official MCP task wrappers when the operator has `DIDA365_TOKEN` and wants official token auth. |
| Project discovery | Web API `project list` | Fast and already part of the sync-derived model. | Official MCP `official project list` or OpenAPI `openapi project list` when validating official channels. |
| Habit and focus reads | Official MCP or OpenAPI | These are first-class official domains and should not be reverse-engineered from private Web API unless official coverage is missing. | Web API `habit ...` / `pomo ...` reads for account settings and web-app-only views. |
| OAuth REST validation | Official OpenAPI | Only OpenAPI verifies the public OAuth REST contract. | None; MCP `dp_...` tokens and Web API cookies are different auth models. |
| Web-app-only metadata | Web API | Settings, comments, sharing, calendar, templates, stats, trash, closed history, and search are not covered by current official channels. | Raw read-only probes when a first-class command is not ready. |
| Unknown private write flow | No command yet | New private writes need endpoint, payload, response, rollback, and live evidence before promotion. | Keep evidence in `docs/research/webapi-probe-log.md` and use `raw get` only for read-only checks. |

## Do Not Mix

- Do not send the Web API cookie `t` to Official MCP or OpenAPI commands.
- Do not send the Official MCP `DIDA365_TOKEN` / `dp_...` token to Web API or
  OpenAPI commands.
- Do not treat an OpenAPI OAuth access token as a browser cookie or MCP token.
- Do not implement a new first-class command only because `official call` or a
  private endpoint can technically reach it; first-class commands need a stable
  job, documented safety behavior, and tests.

## Blocker Exit Criteria

Current blockers are intentionally narrow. A future Agent should only mark them
resolved when the matching evidence exists:

| Blocker | Evidence needed to unblock |
| --- | --- |
| OpenAPI live resource calls | `dida openapi login --browser --json` saves an OAuth token, then at least one read such as `dida openapi project list --json` succeeds. |
| Official MCP known-id habit/focus reads | The account has a disposable habit or focus record, and `official habit get` or `official focus get --type 0|1` succeeds against that ID. |
| Official MCP habit/focus writes | A disposable target exists and the command has already been previewed with `--dry-run`; destructive focus delete also requires `--yes`. |
| Web API task activity | A Pro-entitled account or browser trace returns a successful `GET /task/activity/{taskId}` response with pagination semantics. |
| Task-level attachment commands | A reversible trace shows upload, task association, read-back/download or preview, quota behavior, and cleanup for uploaded-but-not-attached files. |
| Filter, column update/delete/order, collaboration writes | Real traffic captures request bodies, response shapes, ordering semantics, permissions, and rollback paths. |

## Practical Architecture

The current repo direction is:

1. Web API first for breadth
2. Official MCP for clean token-based official access
3. Official OpenAPI for documented OAuth REST coverage where the operator has
   completed browser authorization

That means DidaCLI should not force a single-channel worldview. It should use
the strongest channel for each resource area while keeping the auth and command
boundaries explicit.
