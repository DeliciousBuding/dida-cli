#!/usr/bin/env python3
"""Validate the portable Codex and Claude Code plugin package without network access."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from urllib.parse import urlparse


ROOT = Path(__file__).resolve().parent.parent
PLUGIN = ROOT / "plugins" / "dida-cli"
PLUGIN_SCHEMA = "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"
MCP_SCHEMA = "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json"
DIDA_MCP_URL = "https://mcp.dida365.com"


def load(path: Path) -> dict:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise SystemExit(f"missing {path.relative_to(ROOT)}") from None
    except json.JSONDecodeError as exc:
        raise SystemExit(f"invalid JSON in {path.relative_to(ROOT)}: {exc}") from None
    if not isinstance(value, dict):
        raise SystemExit(f"{path.relative_to(ROOT)} must contain a JSON object")
    return value


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(message)


def require_keys(value: dict, expected: set[str], label: str) -> None:
    actual = set(value)
    require(actual == expected, f"{label} fields must be {sorted(expected)}, got {sorted(actual)}")


def require_https_url(value: object, label: str) -> None:
    require(isinstance(value, str), f"{label} must be a string")
    parsed = urlparse(value)
    require(parsed.scheme == "https" and bool(parsed.netloc), f"{label} must be an absolute HTTPS URL")


def main() -> int:
    portable = load(PLUGIN / "plugin.json")
    claude = load(PLUGIN / ".claude-plugin" / "plugin.json")
    portable_mcp = load(PLUGIN / "mcp.json")
    claude_mcp = load(PLUGIN / ".mcp.json")
    claude_marketplace = load(ROOT / ".claude-plugin" / "marketplace.json")
    codex_marketplace = load(ROOT / ".agents" / "plugins" / "marketplace.json")

    require(portable.get("$schema") == PLUGIN_SCHEMA, "portable plugin schema is incorrect")
    require(portable.get("name") == "dida-cli", "portable plugin name must be dida-cli")
    require(claude.get("name") == "dida-cli", "Claude plugin name must be dida-cli")
    require_keys(portable_mcp, {"$schema", "mcpServers"}, "portable mcp.json")
    require(portable_mcp.get("$schema") == MCP_SCHEMA, "portable MCP schema is incorrect")
    portable_server = portable_mcp.get("mcpServers", {}).get("dida365", {})
    require(isinstance(portable_server, dict), "portable Dida365 MCP server must be an object")
    require_keys(portable_server, {"type", "url"}, "portable Dida365 MCP server")
    require(
        portable_server.get("type") == "streamable-http",
        "portable Dida365 MCP transport must be streamable-http",
    )
    require_https_url(portable_server.get("url"), "portable Dida365 MCP URL")
    require(
        portable_server.get("url") == DIDA_MCP_URL,
        "portable plugin must reference the official Dida365 MCP URL",
    )
    require_keys(claude_mcp, {"mcpServers"}, "Claude .mcp.json")
    claude_server = claude_mcp.get("mcpServers", {}).get("dida365", {})
    require(isinstance(claude_server, dict), "Claude Dida365 MCP server must be an object")
    require_keys(claude_server, {"type", "url"}, "Claude Dida365 MCP server")
    require(claude_server.get("type") == "http", "Claude Dida365 MCP transport must be http")
    require_https_url(claude_server.get("url"), "Claude Dida365 MCP URL")
    require(
        claude_server.get("url") == DIDA_MCP_URL,
        "Claude plugin must reference the official Dida365 MCP URL",
    )
    require(
        portable.get("extensions", {}).get("com.openai", {}).get("onboardingSkill")
        == "./skills/dida-cli/SKILL.md",
        "Codex onboarding skill path is incorrect",
    )
    require((PLUGIN / "skills" / "dida-cli" / "SKILL.md").is_file(), "plugin skill is missing")
    skill = (PLUGIN / "skills" / "dida-cli" / "SKILL.md").read_text(encoding="utf-8")
    require(skill.startswith("---\n"), "plugin skill must start with YAML frontmatter")
    require("name: dida-cli" in skill, "plugin skill name must be dida-cli")
    require("dida official call" in skill, "plugin skill must document official call")
    require("dida openapi call" in skill, "plugin skill must document openapi call")

    require(claude_marketplace.get("name") == "dida-cli", "Claude marketplace name is incorrect")
    require(codex_marketplace.get("name") == "dida-cli", "Codex marketplace name is incorrect")
    require(
        claude_marketplace.get("plugins", [{}])[0].get("source") == "./plugins/dida-cli",
        "Claude marketplace source is incorrect",
    )
    codex_source = codex_marketplace.get("plugins", [{}])[0].get("source", {})
    require(
        codex_source.get("path") == "./plugins/dida-cli",
        "Codex marketplace source is incorrect",
    )
    print("agent plugin manifests valid")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
