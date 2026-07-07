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
| Saved client config auth URL | passed | Redacted client-config flow generated an OAuth authorization URL on 2026-05-10. | Configure redirect URL and complete browser approval. |
| Developer client authentication | passed | Token endpoint returned an OAuth `invalid_grant` response for an intentionally invalid code, which confirms the client-auth path reaches the OAuth server. | Exchange a real callback code. |
| Direct bearer with non-OAuth credential | failed as expected | `/open/v1/...` returned `401 invalid_token` and OAuth bearer challenge. | Do not use non-OAuth credentials as access tokens. |
| Interactive login command | live-verified | Redacted OAuth callback flow completed on 2026-05-10. | Keep this flow stable; continue improving the browser/manual guidance. |
| Interactive login callback validation | passed | 2026-05-10 CLI smoke confirmed `openapi login --redirect-uri http://example.com:17890/callback --json` returns one JSON validation error and rejects non-loopback callback hosts. Unit tests cover `--browser`, custom redirect URI parsing, callback normalization, and invalid callback shapes. | Keep browser OAuth blocked only on real authorization, not callback plumbing. |
| Interactive login short-timeout smoke | blocked by browser approval | Timeout smokes confirmed JSON errors include `authorization_url`, `redirect_uri`, `scope`, `state`, `listen_address`, and ordered `next` actions. A direct authorization URL probe returned a sign-in redirect, with no immediate redirect-URI rejection. | Complete sign-in and authorization in a browser session that can reach the local callback listener. |
| OAuth token status | passed | Redacted OAuth exchange completed without committing token material. | Use a disposable account for follow-up reads and writes. |
| Project endpoint family | partially live-verified | `openapi project list` returned a valid list shape; 2026-07-07 refresh returned 6 projects, and `project get/data` succeeded against a discovered project with count-only evidence. Create/update/delete still rely on local dry-run until a disposable project is chosen. | Run disposable project writes only after explicit cleanup approval. |
| Task endpoint family | partially live-verified | `openapi task get/create/update/complete/delete/move/completed/filter` exist. 2026-07-07 bounded `task completed` returned 25 items, and `task filter` with a valid `projectIds`/`status` payload returned 0 items. An invalid empty `projectId` probe produced an upstream 500 and is not treated as validation evidence. | Run write smoke only with a disposable task and cleanup plan. |
| Focus endpoint family | partially live-verified | `openapi focus get/list/delete` exist with `--dry-run` for delete. 2026-07-07 bounded `focus list` succeeded for type 0 and type 1 with zero records. | Known-id `focus get` and destructive delete need a disposable focus record. |
| Habit endpoint family | partially live-verified | `openapi habit list/get/create/update/checkin/checkins` exist with `--dry-run` for writes. 2026-07-07 `habit list` succeeded with zero habits. | Known-id reads and writes need disposable habit/check-in data. |
| Local dry-run without OAuth token | passed | `openapi project create --dry-run`, `openapi project update --dry-run`, `openapi project delete --dry-run`, `openapi task create --dry-run`, `openapi focus delete --dry-run`, and `openapi habit checkin --dry-run` return JSON previews without a saved token. | Keep this behavior so agents can preview writes before asking the user to complete OAuth. |

## Safe Live Test Plan

1. Run `dida openapi doctor --json`.
2. Configure the developer app OAuth redirect URL to the reported
   `default_redirect_uri`.
3. Run `dida openapi login --browser --json` and complete browser approval, or
   use `auth-url` plus `listen-callback` plus `exchange-code`.
4. Confirm `dida openapi status --json` reports a saved access token without
   printing token material.
5. Run `dida openapi project list --json`.
6. Test remaining read-only OpenAPI `project get/data`, `task get/filter/completed`,
   `focus list`, and `habit list`.
7. Test write commands only on disposable tasks, projects, habits, or focus
   records with a clear cleanup action.
