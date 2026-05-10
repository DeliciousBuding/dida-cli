# OpenAPI Setup

This guide is the operator-facing setup path for the official Dida365 OpenAPI
channel used by DidaCLI.

Use this guide when you need to:

- register or configure a Dida365 developer app
- save `client_id` and `client_secret` locally for DidaCLI
- configure the OAuth `redirect_uri` correctly
- complete OAuth login once and then reuse the saved access token

This guide is not for the Web API cookie channel and not for the official MCP
`DIDA365_TOKEN` channel.

## What You Need

Before you start, have these ready:

- a Dida365 developer app with a valid `client_id`
- the matching `client_secret`
- local access to the machine that will receive the OAuth callback

For the default local login flow in DidaCLI, the callback URL is:

```text
http://127.0.0.1:17890/callback
```

Use that exact value unless you intentionally change the callback host or port.

## Developer Console Setup

In the Dida365 developer console:

1. Open your app settings.
2. Find the OAuth redirect URL or callback URL field.
3. Register this exact URL:

```text
http://127.0.0.1:17890/callback
```

Important details:

- Use `127.0.0.1`, not `localhost`, unless your CLI command also uses
  `localhost`.
- Include the full path `/callback`.
- Include the port `17890`.
- Do not register only the domain or host.
- If you change the callback port locally, update the developer console to the
  same exact URL.

## Save Client ID And Secret

Use stdin for the secret so it does not enter shell history:

```bash
dida openapi client set --id <client-id> --secret-stdin --json
```

Then verify the local setup:

```bash
dida openapi client status --json
dida openapi doctor --json
```

`dida openapi doctor --json` should show:

- `default_redirect_uri`
- `default_scope`
- ordered `next` actions

If `default_redirect_uri` does not match the value registered in the developer
console, stop and fix that first.

## Login Paths

There are two supported login paths.

### Path 1: Browser Shortcut

Use this when automatic browser launch works on your machine:

```bash
dida openapi login --browser --json
```

What it does:

- starts a local callback listener
- tries to open the browser
- waits for the OAuth callback
- exchanges the returned `code`
- saves the access token locally

If browser launch fails, the JSON error includes:

- `error.details.authorization_url`
- `error.details.redirect_uri`
- `error.details.next`

### Path 2: Manual Copy-Paste Flow

Use this when you want the most predictable setup path or when GUI/browser
launch is unreliable.

Step 1:

```bash
dida openapi listen-callback --json
```

Keep that terminal running.

Step 2:

```bash
dida openapi auth-url --json
```

Copy the returned `authorization_url` into a browser and complete login and
authorization.

Step 3:

After the browser redirects back, the first terminal should output:

- `code`
- `state`
- `redirect_uri`

Step 4:

```bash
dida openapi exchange-code --code <code> --json
```

## Verify The Saved Token

After either login path succeeds, verify the saved access token:

```bash
dida openapi status --json
```

Then verify a real API read:

```bash
dida openapi project list --json
```

If `project list` succeeds, the OpenAPI channel is live on that machine.

## Common Errors

### `invalid_request`: redirect URI must be registered

Meaning:

- the Dida365 developer app does not have a matching redirect URL configured

Fix:

- register the exact `default_redirect_uri` from `dida openapi doctor --json`

### `unsupported_response_type`

Meaning:

- the authorization URL is incomplete or hand-edited

Fix:

- do not hand-build the URL
- use `dida openapi auth-url --json` or `dida openapi login --browser --json`

### `ERR_CONNECTION_REFUSED` on `127.0.0.1`

Meaning:

- the browser reached the callback URL, but no local listener was running

Fix:

- start `dida openapi listen-callback --json` first
- then open a newly generated `authorization_url`
- do not reuse an old callback URL from a previous session

### Token saved but later reads say token is missing

Meaning:

- usually a different process or environment was used during a mixed test path

Fix:

- prefer the installed `dida` binary for follow-up checks:

```bash
dida openapi status --json
dida openapi project list --json
```

## Safety Notes

- Do not paste `client_secret`, access tokens, or refresh tokens into chat,
  issues, or committed files.
- Do not commit local callback URLs, secrets, or token files into the repo.
- Keep OpenAPI separate from:
  - Web API cookie auth
  - official MCP `DIDA365_TOKEN`

## Recommended Minimal Setup

If you want the shortest reliable path:

```bash
dida openapi client set --id <client-id> --secret-stdin --json
dida openapi doctor --json
dida openapi listen-callback --json
dida openapi auth-url --json
```

Then:

1. Copy the `authorization_url` into a browser.
2. Complete authorization.
3. Exchange the returned `code`.
4. Verify with `dida openapi project list --json`.
