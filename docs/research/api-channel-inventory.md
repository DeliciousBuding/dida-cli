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

## Practical Architecture

The current repo direction is:

1. Web API first for breadth
2. Official MCP for clean token-based official access
3. Official OpenAPI for documented OAuth REST coverage where the operator has
   completed browser authorization

That means DidaCLI should not force a single-channel worldview. It should use
the strongest channel for each resource area while keeping the auth and command
boundaries explicit.
