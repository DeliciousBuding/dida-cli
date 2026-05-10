# Official OpenAPI Guide

This document is a repo-authored summary of the official Dida365 OpenAPI page.
It is intended for implementation planning and channel comparison inside DidaCLI.

Source material was reviewed from a locally saved copy of the developer page
and the public documentation URL:

- `https://developer.dida365.com/docs/index.html#/openapi`

## Purpose

The official OpenAPI is a REST-style interface for Dida365 account data.
Compared with the private Web API, it is narrower but better documented.
Compared with the official MCP channel, it is lower-level and OAuth-oriented.

## Auth Model

The official OpenAPI is not based on the browser cookie `t`.
It is also not the same token model as the official MCP `dp_...` token flow.

Observed auth flow:

1. Register an app in the Dida365 developer center to obtain a `client_id` and `client_secret`.
2. Redirect the user to the Dida365 authorization page:
   `https://dida365.com/oauth/authorize`
3. Receive an authorization `code` on the configured `redirect_uri`.
4. Exchange that `code` for an OAuth access token at:
   `https://dida365.com/oauth/token`
5. Call OpenAPI endpoints with:
   `Authorization: Bearer <access_token>`

Important implication for DidaCLI:

- this channel is the OAuth-based third channel for documented REST resources
- it should not be modeled as a drop-in replacement for Web API cookie auth
- it should not be conflated with the official MCP `DIDA365_TOKEN=dp_...` flow

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
- a real OAuth authorization code flow is still required before the OpenAPI can be used as a live third channel
- DidaCLI now has local OAuth plumbing for this flow:
  `openapi client set`, `openapi auth-url`, `openapi login`,
  `openapi listen-callback`, `openapi exchange-code`, `openapi status`, and
  `openapi logout`
- `openapi login` is designed for local callback URLs such as
  `http://127.0.0.1:17890/callback`; non-loopback redirect hosts are rejected
  before browser launch so agents get a fast JSON error instead of a long wait

## Main Resource Areas

The reviewed page groups the official OpenAPI into four main resource areas:

1. Task
2. Project
3. Focus
4. Habit

It also includes a Definitions section for the main response models.

## Task Endpoints

Documented task paths include:

- `GET /open/v1/project/{projectId}/task/{taskId}`
- `POST /open/v1/task`
- `POST /open/v1/task/{taskId}`
- `POST /open/v1/project/{projectId}/task/{taskId}/complete`
- `DELETE /open/v1/project/{projectId}/task/{taskId}`
- `POST /open/v1/task/move`
- `POST /open/v1/task/completed`
- `POST /open/v1/task/filter`

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
- `POST /open/v1/project`
- `POST /open/v1/project/{projectId}`
- `DELETE /open/v1/project/{projectId}`

The project-related models include:

- Project
- Column
- ProjectData

This is notable because the official OpenAPI exposes both project CRUD and
project-with-data responses with column shapes, even though the command surface
is still much smaller than the private Web API we are currently using.

## Focus Endpoints

Documented focus paths include:

- `GET /open/v1/focus/{focusId}`
- `GET /open/v1/focus`
- `DELETE /open/v1/focus/{focusId}`

The page documents an `OpenFocus` model and related task-brief structures.
This is now wrapped in DidaCLI as `openapi focus get/list/delete`, pending live
OAuth validation.

## Habit Endpoints

Documented habit paths include:

- `GET /open/v1/habit/{habitId}`
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

Compared with the private Web API currently used by DidaCLI, the reviewed page
does not give us the broader surface we already rely on for:

- comments
- tags
- project groups / folders
- share metadata
- calendar subscription metadata
- template listing
- broad account settings and Web-specific preferences
- raw search across multiple resource types
- closed history and many web-only operational views

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
