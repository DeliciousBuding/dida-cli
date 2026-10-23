# DidaCLI agent plugin

This lightweight portable plugin packages:

- a Dida365/TickTick workflow skill under `skills/dida-cli/`
- the official remote MCP server `https://mcp.dida365.com`
- the same plugin as a Claude Code plugin via `.claude-plugin/plugin.json`

It does not bundle a local MCP process, OAuth token, browser cookie, or task
data. The official MCP server completes OAuth on first use.

## Install

From a checkout:

```bash
python3 scripts/install-agent-plugin.py codex
python3 scripts/install-agent-plugin.py claude
```

On Windows, replace `python3` with `python`.

The installer installs `@delicious233/dida-cli` globally when `dida` is missing,
then adds the repository marketplace, installs `dida-cli@dida-cli`, and verifies
that the native client reports the plugin as enabled. Pass `--no-cli` when only
the plugin/MCP server should be installed.

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
