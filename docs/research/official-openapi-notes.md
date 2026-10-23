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

An OAuth client secret from the developer app settings was intentionally tested
as a bearer token and rejected by the official OpenAPI with:

```text
401 invalid_token
```

and `WWW-Authenticate: Bearer realm="oauth"`.

## Interpretation

- An OAuth client secret is not an access token.
- The current official OpenAPI documentation separately supports a personal API
  token created in the web app, plus OAuth access tokens for app authorization.
- DidaCLI maps personal OpenAPI tokens to `DIDA365_OPENAPI_TOKEN` and keeps them
  separate from the MCP `DIDA365_TOKEN`.

## Practical Conclusion

There are currently three distinct channels:

1. Web API
   - auth: browser cookie `t`
   - best coverage

2. Official MCP
   - auth: `DIDA365_TOKEN=dp_...`
   - official task/project/habit/focus tools

3. Official OpenAPI
   - auth: personal API token or OAuth access token
   - client secrets are never accepted as bearer tokens

## Recommendation

- Keep Web API as the main coverage channel.
- Keep official MCP as the official task-tool channel.
- Use official OpenAPI with `DIDA365_OPENAPI_TOKEN` for lightweight
  single-account REST automation, or use OAuth for multi-user app flows.
