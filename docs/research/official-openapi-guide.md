# Official OpenAPI Guide

This document is a repo-authored summary of the official Dida365 OpenAPI page.
It is intended for implementation planning and channel comparison inside DidaCLI.

Source material was reviewed from the live developer documentation and the
public documentation URL:

- `https://developer.dida365.com/docs/index.html#/openapi`
- `https://developer.dida365.com/docs/openapi.md`

The expanded resource mapping was rechecked on 2026-10-23 against
`@suibiji/dida-cli@0.1.14`, the DIDA CLI package linked by the official
Dida365 help center. DidaCLI still treats the developer page and live API
responses as the primary contract.

## Purpose

The official OpenAPI is a REST-style interface for Dida365 account data.
Compared with the private Web API, it is narrower but better documented.
Compared with the official MCP channel, it is lower-level and OAuth-oriented.

## Auth Model

The official OpenAPI is not based on the browser cookie `t`.
It accepts either a personal API token or an OAuth access token as a Bearer
credential. DidaCLI keeps the personal-token environment variable separate from
the MCP token channel.

Observed auth options:

1. For single-account automation, create a personal API token in the web app
   and send it as `Authorization: Bearer <personal-token>`.
2. For app authorization, register an app in the Dida365 developer center to
   obtain a `client_id` and `client_secret`.
3. Redirect the user to the Dida365 authorization page:
   `https://dida365.com/oauth/authorize`
4. Receive an authorization `code` on the configured `redirect_uri`.
5. Exchange that `code` for an OAuth access token at:
   `https://dida365.com/oauth/token`
6. Call OpenAPI endpoints with:
   `Authorization: Bearer <access_token>`
7. Revoke a saved OAuth access token at:
   `https://dida365.com/oauth/revoke`

Important implications for DidaCLI:

- this channel is the documented REST third channel; OAuth is one credential
  mode and personal API tokens are the lightweight mode
- it should not be modeled as a drop-in replacement for Web API cookie auth
- personal OpenAPI tokens use `DIDA365_OPENAPI_TOKEN`, separate from the MCP
  `DIDA365_TOKEN=dp_...` flow

## Verified Runtime Findings

The following behavior was verified directly during implementation work:

- the provided `client_id` and `client_secret` are accepted by the OAuth token endpoint as a valid client pair
- posting an invalid authorization code returns:
  - `400 invalid_grant`
  - `Invalid authorization code: ...`
- attempting the authorization URL without a signed-in user returns a redirect to:
  - `/signin?dest=...`
- calling `/open/v1/...` directly with that same secret as a bearer token returns:
  - `401 invalid_token`
  - `WWW-Authenticate: Bearer realm="oauth"`

Practical conclusion:

- `client_id` and `client_secret` are not themselves an OpenAPI access token
- an OpenAPI call still needs either a personal API token or a real OAuth access
  token
- DidaCLI now has local OAuth plumbing for this flow:
  `openapi client set`, `openapi auth-url`, `openapi login`,
  `openapi listen-callback`, `openapi exchange-code`, `openapi status`, and
  `openapi logout`
- `openapi login` is designed for local callback URLs such as
  `http://127.0.0.1:17890/callback`; non-loopback redirect hosts are rejected
  before browser launch so agents get a fast JSON error instead of a long wait
- `DIDA365_OPENAPI_TOKEN` bypasses OAuth setup for single-account use and is
  reported as the active source by `openapi status`
- `openapi logout` revokes a saved OAuth token through
  `POST https://dida365.com/oauth/revoke` before clearing local state; the
  endpoint accepts form-encoded `token` values and returns an idempotent `200`
  for an already-invalid token

## Main Resource Areas

The current mapping spans these resource areas:

1. Task
2. Project
3. Project group and column
4. Task comment and assignment
5. Tag
6. Focus
7. Habit
8. Countdown and preference

It also includes a Definitions section for the main response models.

## Task Endpoints

Documented task paths include:

- `GET /open/v1/project/{projectId}/task/{taskId}`
- `POST /open/v1/task`
- `POST /open/v1/task/{taskId}`
- `POST /open/v1/project/{projectId}/task/{taskId}/complete`
- `POST /open/v1/task/completeTasks`
- `DELETE /open/v1/project/{projectId}/task/{taskId}`
- `POST /open/v1/task/move`
- `POST /open/v1/task/assign`
- `POST /open/v1/task/unassign`
- `POST /open/v1/task/completed`
- `POST /open/v1/task/filter`
- `POST /open/v1/task/search`
- `POST /open/v1/task/undone`
- `GET /open/v1/project/{projectId}/task/{taskId}/comments`
- `POST /open/v1/project/{projectId}/task/{taskId}/comment`
- `DELETE /open/v1/project/{projectId}/task/{taskId}/comment/{id}`

The task model in the page covers the familiar fields:

- `id`
- `projectId`
- `title`
- `content`
- `desc`
- `startDate`
- `dueDate`
- `timeZone`
- `isAllDay`
- `priority`
- `status`
- `items`
- `tags`
- `columnId`
- `parentId`

## Project Endpoints

Documented project paths include:

- `GET /open/v1/project`
- `GET /open/v1/project/{projectId}`
- `GET /open/v1/project/{projectId}/data`
- `GET /open/v1/project/{projectId}/members`
- `POST /open/v1/project`
- `POST /open/v1/project/{projectId}`
- `DELETE /open/v1/project/{projectId}`
- `GET /open/v1/project/group`
- `POST /open/v1/project/group`
- `POST /open/v1/project/group/{projectGroupId}`
- `DELETE /open/v1/project/group/{projectGroupId}`
- `GET /open/v1/project/{projectId}/column`
- `POST /open/v1/project/{projectId}/column`
- `POST /open/v1/project/{projectId}/column/{columnId}`

The project-related models include:

- Project
- Column
- ProjectData

The official OpenAPI exposes both project CRUD and project-with-data responses
with column shapes. Its command surface remains smaller than the private Web API
used for broad account coverage.

## Focus Endpoints

Documented focus paths include:

- `GET /open/v1/focus/{focusId}`
- `GET /open/v1/focus`
- `POST /open/v1/focus`
- `POST /open/v1/focus/{focusId}`
- `DELETE /open/v1/focus/{focusId}`

The page documents an `OpenFocus` model and related task-brief structures.
This is now wrapped in DidaCLI as `openapi focus get/list/delete`, pending live
OAuth validation.

## Habit Endpoints

Documented habit paths include:

- `GET /open/v1/habit/{habitId}`
- `GET /open/v1/habit/sections`
- `GET /open/v1/habit`
- `POST /open/v1/habit`
- `POST /open/v1/habit/{habitId}`
- `POST /open/v1/habit/{habitId}/checkin`
- `GET /open/v1/habit/checkins`

The page also documents the habit data models:

- `OpenHabit`
- `OpenHabitCheckinData`
- `OpenHabitCheckin`

This is now wrapped in DidaCLI as
`openapi habit list/get/create/update/checkin/checkins`, pending live OAuth
validation.

## Tag, Preference, And Countdown Endpoints

Additional documented and mapped paths include:

- `GET /open/v1/tag`
- `POST /open/v1/tag`
- `PUT /open/v1/tag`
- `PUT /open/v1/tag/rename?scope=...`
- `DELETE /open/v1/tag/delete?scope=...`
- `POST /open/v1/preference`
- `GET /open/v1/countdown`

These do not yet have dedicated first-class wrappers. Use the constrained
generic command:

```bash
dida openapi call GET /tag --json
dida openapi call POST /preference --args-json '{}' --json
dida openapi call POST /tag --args-json '{"label":"work"}' --dry-run --json
```

`GET` and the known read-only POST paths `/preference`, `/task/completed`,
`/task/filter`, `/task/search`, and `/task/undone` execute without `--yes`.
Other writes require
`--yes`.

## Definitions Covered By The Page

The reviewed page includes definitions for at least:

- `ChecklistItem`
- `Task`
- `Project`
- `Column`
- `ProjectData`
- `OpenPomodoroTaskBrief`
- `OpenFocus`
- `OpenHabit`
- `OpenHabitCheckinData`
- `OpenHabitCheckin`

These definitions are useful when comparing:

- official field names
- Web API field names
- MCP tool schemas
- DidaCLI normalization choices

## What The Official OpenAPI Does Not Cover Well

Compared with the private Web API currently used by DidaCLI, the official
mapping still does not replace:

- share metadata
- calendar subscription metadata
- template listing
- broad account settings and Web-specific preferences
- closed history and many web-only operational views
- session and device metadata

So even with OAuth support in DidaCLI, the official OpenAPI still should not
replace the Web API as the broadest coverage channel.

## Practical DidaCLI Direction

The most defensible architecture remains:

1. Web API first for breadth of coverage
2. Official MCP as the clean token-based official channel
3. Official OpenAPI as the documented OAuth REST channel for selected resources

The current first-class official OpenAPI command surface covers:

- project get/list/data/create/update/delete
- task get/create/update/complete/delete/move/completed/filter
- focus get/list/delete
- habit list/get/create/update/checkin/checkins

The `openapi call` escape hatch now covers newly documented endpoints without a
first-class wrapper while preserving the OAuth host boundary and a write
confirmation gate.

The least urgent official OpenAPI work is broad account-level functionality,
because that is where the private Web API is currently much more capable.

## Notes For Future Implementation

- Do not assume `dp_...` MCP tokens work for the OpenAPI.
- Keep OpenAPI commands behind the OAuth access-token boundary. The login
  plumbing exists, but live resource verification still requires a saved OAuth
  token from browser authorization.
- Keep channel boundaries explicit in CLI docs and schema metadata.
- Prefer resource areas where the official contract is clearly better than the
  private Web API, not just because an official endpoint exists.
