# OpenAPI Live Validation Log

This log tracks the official OAuth OpenAPI channel. It must not contain OAuth
client secrets, access tokens, refresh tokens, local secret paths, or full
private response payloads.

## Auth Model

- Authorization endpoint: `https://dida365.com/oauth/authorize`
- Token endpoint: `https://dida365.com/oauth/token`
- API base: `https://api.dida365.com/open/v1`
- Required API request header: `Authorization: Bearer <oauth_access_token>`

The OpenAPI channel does not accept the official MCP `dp_...` token and does
not accept the OAuth client secret as a bearer token.

## Validation Events

| Check | Result | Evidence | Next action |
| --- | --- | --- | --- |
| Generate authorization URL | passed | `openapi auth-url` builds the expected OAuth URL. | Complete browser approval. |
| Developer client authentication | passed | Token endpoint returned an OAuth `invalid_grant` response for an intentionally invalid code, which confirms the client-auth path reaches the OAuth server. | Exchange a real callback code. |
| Direct bearer with non-OAuth credential | failed as expected | `/open/v1/...` returned `401 invalid_token` and OAuth bearer challenge. | Do not use non-OAuth credentials as access tokens. |
| Interactive login command | implemented, not fully live-verified | CLI has `openapi login` with callback listener, callback `code` validation, callback `state` validation, and token exchange. | Run full browser authorization. |
| Current local OAuth state | blocked | `openapi status --json` reports no saved token; `openapi doctor --json` reports no client env in the current shell. | Provide env locally, then run browser authorization. |
| Project list | implemented, not fully live-verified | `openapi project list` exists. | Run after token persistence succeeds. |
| Project get/data | implemented, not fully live-verified | `openapi project get` and `openapi project data` exist. | Run after project list succeeds. |
| Task endpoint family | implemented, not fully live-verified | `openapi task get/create/update/complete/delete/move/completed/filter` exist. | Run read smoke after project list succeeds; write smoke only with disposable task. |
| Focus endpoint family | documented only | Official docs define focus get/list/delete. | Implement after OAuth project list succeeds. |
| Habit endpoint family | documented only | Official docs define habit CRUD and check-ins. | Implement after OAuth project list succeeds. |

## Safe Live Test Plan

1. Run `dida openapi doctor --json`.
2. Run `dida openapi login --json` and complete browser approval.
3. Confirm `dida openapi status --json` reports a saved access token without
   printing token material.
4. Run `dida openapi project list --json`.
5. Only after project list succeeds, test read-only OpenAPI `project get/data`,
   `task get/filter/completed`, `focus list`, and `habit list`.
6. Test write commands only on disposable tasks, projects, habits, or focus
   records with a clear cleanup action.
