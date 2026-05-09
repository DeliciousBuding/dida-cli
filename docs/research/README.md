# Research Index

This directory tracks API evidence for DidaCLI. It is not a dumping ground for
raw captures, secrets, cookies, or local machine paths.

## Core Maps

- [API channel inventory](api-channel-inventory.md): the three-channel model.
- [API surfaces](api-surfaces.md): broad notes from earlier Web API and tool
  research.
- [Official channel validation matrix](official-channel-validation-matrix.md):
  what has been live-tested, documented only, or blocked.
- [Roadmap completion audit](roadmap-completion-audit.md): conservative
  objective-to-evidence checklist and remaining blockers.
- [Prompt-to-artifact checklist](prompt-to-artifact-checklist.md): detailed
  requirement-by-requirement evidence map for the active objective.

## Official Channels

- [Official MCP tool crosswalk](official-mcp-tool-crosswalk.md): MCP tool names
  mapped to DidaCLI commands and wrapping candidates.
- [Official MCP wrapping policy](official-mcp-wrapping-policy.md): rules for
  promoting MCP tools into first-class commands.
- [Official MCP vs Web API](official-mcp-vs-webapi.md): channel tradeoffs.
- [Official OpenAPI guide](official-openapi-guide.md): curated OpenAPI summary.
- [Official OpenAPI notes](official-openapi-notes.md): OAuth and token findings.
- [OpenAPI live validation log](openapi-live-validation-log.md): live OAuth and
  endpoint validation status.

## Web API

- [Web API gap catalog](webapi-gap-catalog.md): remaining private Web API gaps.
- [Web API probe log](webapi-probe-log.md): probe results and next evidence
  needed before implementation.

## Documentation Rules

- Commit curated summaries, not raw private payload dumps.
- Record exact endpoint names, request method, required identifiers, and observed
  status codes.
- Do not commit cookies, tokens, OAuth secrets, local secret paths, full user
  profiles, or full task exports.
- If a private endpoint returns `500`, do not wrap it as a command until the
  missing query/body shape is known.
