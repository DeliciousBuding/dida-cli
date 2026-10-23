# DidaCLI agent plugin

This lightweight portable plugin packages:

- a Dida365/TickTick workflow skill under `skills/dida-cli/`
- the official remote MCP server `https://mcp.dida365.com`
- the same plugin as a Claude Code plugin via `.claude-plugin/plugin.json`

It does not bundle a local MCP process, OAuth token, browser cookie, or task
data. The official MCP server completes OAuth on first use.

The root `plugin.json` and `mcp.json` follow Agent Plugins 1.0.0. Claude Code
uses its native `.claude-plugin/plugin.json` and `.mcp.json` adapter, including
Claude's `http` transport name.

## Install

Run the installer from a checkout; by default it installs the plugin
marketplace from GitHub `main`:

```bash
python3 scripts/install-agent-plugin.py codex
python3 scripts/install-agent-plugin.py claude
```

On Windows, replace `python3` with `python`.

Pass `--local` to test the plugin files in the current checkout instead of the
GitHub marketplace. `--source <owner/repo[@ref]>` selects another Git source.

The installer installs `@delicious233/dida-cli` globally when `dida` is missing
and then runs its checksum-verified self-updater, so a lagging npm release does
not leave a fresh plugin install on an older CLI. It then adds the marketplace,
installs `dida-cli@dida-cli`, and verifies both the enabled plugin and the
client registration of `https://mcp.dida365.com`. Pass `--no-cli` when only the
plugin/MCP server should be installed.

Plugin installation does not grant account access. If the client reports that
the `dida365` server needs authentication, complete OAuth interactively with
`codex mcp login dida365` or
`claude mcp login plugin:dida-cli:dida365`.

Native manual install commands:

```bash
codex plugin marketplace add DeliciousBuding/dida-cli --sparse .agents/plugins --sparse plugins/dida-cli
codex plugin add dida-cli@dida-cli

claude plugin marketplace add DeliciousBuding/dida-cli --sparse .claude-plugin plugins/dida-cli
claude plugin install dida-cli@dida-cli -y
```

## Validate

```bash
python3 scripts/validate-agent-plugins.py
claude plugin validate ./plugins/dida-cli --strict
```
