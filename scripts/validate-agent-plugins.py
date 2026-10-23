#!/usr/bin/env python3
"""Validate the portable Codex and Claude Code plugin package without network access."""

from __future__ import annotations

import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
PLUGIN = ROOT / "plugins" / "dida-cli"


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


def main() -> int:
    portable = load(PLUGIN / "plugin.json")
    claude = load(PLUGIN / ".claude-plugin" / "plugin.json")
    portable_mcp = load(PLUGIN / "mcp.json")
    claude_mcp = load(PLUGIN / ".mcp.json")
    claude_marketplace = load(ROOT / ".claude-plugin" / "marketplace.json")
    codex_marketplace = load(ROOT / ".agents" / "plugins" / "marketplace.json")

    require(portable.get("name") == "dida-cli", "portable plugin name must be dida-cli")
    require(claude.get("name") == "dida-cli", "Claude plugin name must be dida-cli")
    require(portable_mcp == claude_mcp, "portable and Claude MCP configs must match")
    require(
        portable_mcp.get("mcpServers", {}).get("dida365", {}).get("url")
        == "https://mcp.dida365.com",
        "plugin must reference the official Dida365 MCP URL",
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
