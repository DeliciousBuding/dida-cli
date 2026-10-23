---
name: dida-cli
description: Use DidaCLI and the official Dida365 MCP server to inspect or manage tasks, projects, habits, focus records, countdowns, tags, and calendar data. Dida365/TickTick task and reminder workflows.
---

# DidaCLI plugin

Use the installed `dida` command as the command and schema authority. Start with:

```bash
dida --help
dida schema list --compact --json
dida channel list --json
```

If `dida` is missing, ask before installing it globally:

```bash
npm install -g @delicious233/dida-cli
```

## Channel selection

| Need | Preferred surface |
| --- | --- |
| Broad Web API reads and account metadata | `dida <command>` |
| Official MCP task, project, habit, focus, countdown, calendar, comment, tag, or group tools | `dida official ...` or the bundled `dida365` MCP server |
| Official OpenAPI REST reads and writes | `dida openapi ...` |
| A newly documented endpoint without a first-class wrapper | `dida official call ... --dry-run` or `dida openapi call ... --dry-run` |

Do not mix browser cookies, Official MCP tokens, and OpenAPI personal/OAuth
credentials.

## Safety

- Preview every write with `--dry-run --json`.
- Require `--yes` for destructive operations and for generic `official call`
  tools outside the fixed read-only allowlist.
- Read a target back after writes when the channel supports it.
- Use `dida official tools --json` to discover the live MCP tool catalog. The live server is authoritative.
- Never print or ask the user to paste cookies, bearer tokens, OAuth secrets, private task exports, or raw sensitive response bodies.
- `dida task update` is a full-object write. Do not use it to patch a single omitted field without first reading the task and planning the full payload.

## Official MCP discovery

```bash
dida official doctor --json
dida official tools --limit 100 --json
dida official show <tool-name> --json
dida official call <tool-name> --args-json '{"...":"..."}' --dry-run --json
```

Use first-class wrappers for common operations. Use `official call --dry-run`
for newly published tools, then execute writes or unrecognized future tools
only after review with `--yes`. MCP annotations are discovery hints, not
authorization.

## Official OpenAPI discovery

Set `DIDA365_OPENAPI_TOKEN` for single-account use, or complete OAuth login.
The two credential modes remain separate from the MCP token.

```bash
dida openapi doctor --json
dida openapi status --json
dida openapi call GET /tag --json
dida openapi call GET /countdown --json
dida openapi call POST /preference --args-json '{}' --json
dida openapi call POST /tag --args-json '{"label":"work"}' --dry-run --json
```

`GET` and known read-only `POST` requests execute directly. Other writes require
`--yes`.
