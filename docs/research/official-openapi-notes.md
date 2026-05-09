# Official OpenAPI Notes

This note records the current findings for the official dida365 OpenAPI channel.

## What Was Tested

- Developer portal: `https://developer.dida365.com/docs#/openapi`
- OpenAPI base candidates:
  - `https://api.dida365.com/open/v1/project`
  - `https://api.dida365.com/open/v1/project/all`
  - `https://api.dida365.com/open/v1/task`
  - `https://api.dida365.com/open/v1/user/preferences`

## Result

The provided token:

```text
85r59DpD7RmUiStIJ7z7610cDbFlHXv5
```

was rejected by the official OpenAPI with:

```text
401 invalid_token
```

and `WWW-Authenticate: Bearer realm="oauth"`.

## Interpretation

- The official OpenAPI does not accept the above token as a direct bearer token.
- The official MCP channel does accept the `dp_...` API token through `DIDA365_TOKEN`.
- Therefore, the official OpenAPI is likely tied to an OAuth access-token flow, not the simple API-token flow used by official MCP.

## Practical Conclusion

There are currently three distinct channels:

1. Web API
   - auth: browser cookie `t`
   - best coverage

2. Official MCP
   - auth: `DIDA365_TOKEN=dp_...`
   - clean and officially exposed

3. Official OpenAPI
   - auth: likely OAuth access token
   - not directly usable with the tested non-OAuth token

## Recommendation

- Keep Web API as the main coverage channel.
- Keep official MCP as the official token-based channel.
- Treat official OpenAPI as a future third channel only after obtaining a real OAuth access token and documenting the full auth flow.
