# Agent Plugin Design

Last verified: 2026-10-23.

The DidaCLI plugin is intentionally small: one workflow skill and the official
remote Dida365 MCP server. It does not bundle a local MCP process, OAuth token,
browser cookie, task export, hook, or privileged helper.

## Package Layout

```text
plugins/dida-cli/
  plugin.json                 Agent Plugins 1.0.0 package manifest
  mcp.json                    Agent Plugins 1.0.0 remote MCP declaration
  .claude-plugin/plugin.json  Claude Code manifest adapter
  .mcp.json                   Claude Code native MCP adapter
  skills/dida-cli/SKILL.md    shared workflow instructions
  README.md                   install and trust notes
```

Both clients share the same skill and remote endpoint, but use their own
transport vocabulary. The portable root files target the closed Agent Plugins
1.0.0 schemas; the Claude adapter keeps Claude Code's native `.mcp.json`
format. The validator checks both contracts rather than assuming they are
byte-identical.

The repository also exposes:

- `.agents/plugins/marketplace.json` for Codex
- `.claude-plugin/marketplace.json` for Claude Code
- `scripts/install-agent-plugin.py` for native two-command installation
- `scripts/validate-agent-plugins.py` for offline manifest checks

## MCP Design Rules

The plugin consumes the official MCP server as-is. DidaCLI does not rewrite the
server's tool schemas or hide write-capable tools behind a parallel local
server.

- Tool discovery uses the live `tools/list` result.
- New tools can be previewed with `dida official call <tool> --dry-run`.
- Tool names and schemas remain the upstream contract.
- Tool annotations are treated as untrusted hints, not authorization.
- The CLI prefers MCP `2026-07-28` stateless discovery and falls back to the
  legacy initialize handshake only when the server does not advertise the
  current protocol.
- `official call` has a fixed read-only allowlist. Writes and unrecognized
  future tools require DidaCLI's `--yes` boundary.
- Browser cookies, MCP tokens, and OpenAPI personal/OAuth credentials never
  cross channels.
- The portable declaration uses Agent Plugins 1.0.0 `streamable-http`; the
  Claude adapter uses Claude Code's native `http` transport.

This follows the MCP tool model of model-controlled discovery with a human in
the loop and JSON Schema-backed inputs rather than a bespoke protocol dialect.

Reference points:

- [MCP 2026-07-28 tools](https://modelcontextprotocol.io/specification/2026-07-28/server/tools)
- [MCP 2026-07-28 discovery](https://modelcontextprotocol.io/specification/2026-07-28/server/discover)
- [Agent Plugins 1.0.0 specification](https://agent-plugins.org/specification)
- [Agent Plugins MCP configuration](https://agent-plugins.org/plugin-authors/mcp-servers)
- [Codex plugin packaging](https://developers.openai.com/plugins/build/plugins)
- [Claude Code plugin manifest](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code plugin installation](https://code.claude.com/docs/en/plugins/install)
- [Claude Code marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)

## Authentication

The bundled remote server is `https://mcp.dida365.com`.

- Codex/Claude Code complete the official MCP OAuth or bearer-token flow.
- The plugin contains no credential.
- DidaCLI's Official MCP token store remains a separate local path.
- OpenAPI remains a CLI-only channel because it uses a different personal-token
  or OAuth credential model and endpoint set.

## Installation

Native clients remain the installation authority:

```bash
codex plugin marketplace add DeliciousBuding/dida-cli --sparse .agents/plugins --sparse plugins/dida-cli
codex plugin add dida-cli@dida-cli

claude plugin marketplace add DeliciousBuding/dida-cli --sparse .claude-plugin plugins/dida-cli
claude plugin install dida-cli@dida-cli
```

From a checkout, `python3 scripts/install-agent-plugin.py codex` or `claude`
installs the marketplace from GitHub `main` by default, installs the CLI first
when it is missing, and checks that the client reports the plugin as enabled.
Pass `--local` to test the current checkout or `--source` to select another Git
marketplace. Windows uses `python` instead of `python3`.

The local install path was verified on 2026-10-23 with Codex CLI `0.160.1` and
Claude Code `2.1.292`; both clients resolved the remote MCP declaration,
installed the plugin, and reported it enabled.

## Maintenance Rules

- Update the plugin version when skills or bundled MCP behavior changes.
- Keep the portable `mcp.json` on the declared Agent Plugins schema and the
  Claude `.mcp.json` on Claude Code's native schema.
- Keep both marketplace names and plugin names equal to `dida-cli`.
- Do not add lifecycle hooks unless a concrete workflow requires them.
- Validate both manifests on every change.
- Re-check the live Official MCP catalog when the upstream tool count or tool
  names change; the live catalog wins over dated research snapshots.
