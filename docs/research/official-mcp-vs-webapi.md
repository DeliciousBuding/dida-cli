# Official MCP vs Web API

This note records the current view of DidaCLI's two possible upstream channels.

## Summary

- Official MCP is cleaner, narrower, and token-based.
- Web API is broader, less stable, and cookie-based.
- For DidaCLI, the practical direction is `Web API first, official MCP supported where it is better`.

## Official MCP

- Endpoint: `https://mcp.dida365.com`
- Auth: `Bearer <DIDA365_TOKEN>`
- Shape: JSON-RPC over Streamable HTTP MCP
- Observed strengths:
  - official token flow
  - explicit tool catalogue
  - typed schemas for advertised tools
  - cleaner external integration story
- Observed limits:
  - smaller capability surface than the observed Web API
  - many account-level and Web-only surfaces are not present

## Web API

- Endpoint family: `https://api.dida365.com/api/v1|v2`
- Auth: browser cookie `t`
- Shape: direct resource endpoints
- Observed strengths:
  - broader coverage
  - closer to real web-app behavior
  - includes settings, comments, sharing metadata, templates, statistics, closed history, and other surfaces not exposed by official MCP
- Observed limits:
  - private and drift-prone
  - weaker official stability guarantees
  - auth is less clean than token-based official MCP

## Practical Direction

- Keep the main CLI on the Web API for coverage.
- Add a minimal official channel for:
  - health checks
  - tool catalogue introspection
  - future selective wrapping of official tools where they are cleaner than Web API equivalents
- Avoid duplicating the entire CLI surface twice unless there is a real behavioral or stability gain.

## TickTick

- The current CLI is Dida365-first.
- TickTick support is reasonable as a future compatibility target, but only after:
  - command contracts remain stable
  - the auth model is clearly separated
  - endpoint drift between CN and global products is documented
