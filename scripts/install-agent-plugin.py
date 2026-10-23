#!/usr/bin/env python3
"""Install the DidaCLI plugin through the native Codex or Claude Code client."""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
from pathlib import Path


DEFAULT_MARKETPLACE_SOURCE = "DeliciousBuding/dida-cli"


def run(command: list[str]) -> None:
    print("+", " ".join(command))
    subprocess.run(command, check=True)


def run_json(command: list[str]) -> object:
    print("+", " ".join(command))
    completed = subprocess.run(command, check=True, text=True, capture_output=True)
    if completed.stderr:
        print(completed.stderr, end="", file=sys.stderr)
    try:
        return json.loads(completed.stdout)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"{command[0]} returned invalid JSON: {exc}") from None


def verify_codex_plugin() -> None:
    inventory = run_json(["codex", "plugin", "list", "--json"])
    if not isinstance(inventory, dict):
        raise SystemExit("codex plugin list returned an unexpected shape")
    installed = inventory.get("installed", [])
    if not any(item.get("pluginId") == "dida-cli@dida-cli" and item.get("enabled") for item in installed):
        raise SystemExit("Codex installed dida-cli but did not report it as enabled")


def verify_claude_plugin() -> None:
    inventory = run_json(["claude", "plugin", "list", "--json"])
    if not isinstance(inventory, list):
        raise SystemExit("Claude plugin list returned an unexpected shape")
    if not any(item.get("id") == "dida-cli@dida-cli" and item.get("enabled") for item in inventory):
        raise SystemExit("Claude Code installed dida-cli but did not report it as enabled")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("target", choices=("codex", "claude"))
    source = parser.add_mutually_exclusive_group()
    source.add_argument(
        "--source",
        help="Git marketplace source (owner/repo[@ref], URL, or local path)",
    )
    source.add_argument(
        "--local",
        action="store_true",
        help="Install the marketplace from this checkout instead of GitHub",
    )
    parser.add_argument(
        "--no-cli",
        action="store_true",
        help="Do not install @delicious233/dida-cli when dida is missing",
    )
    args = parser.parse_args()

    repo_root = Path(__file__).resolve().parent.parent
    marketplace_source = (
        str(repo_root)
        if args.local
        else args.source or DEFAULT_MARKETPLACE_SOURCE
    )
    if not args.no_cli and shutil.which("dida") is None:
        if shutil.which("npm") is None:
            print("dida is missing and npm is not available", file=sys.stderr)
            return 1
        run(["npm", "install", "-g", "@delicious233/dida-cli"])

    if args.target == "codex":
        if shutil.which("codex") is None:
            print("codex CLI is not installed or not on PATH", file=sys.stderr)
            return 1
        command = ["codex", "plugin", "marketplace", "add", marketplace_source]
        if not args.local:
            command.extend(
                ["--sparse", ".agents/plugins", "--sparse", "plugins/dida-cli"]
            )
        command.append("--json")
        run(command)
        run(["codex", "plugin", "add", "dida-cli@dida-cli", "--json"])
        verify_codex_plugin()
    else:
        if shutil.which("claude") is None:
            print("claude CLI is not installed or not on PATH", file=sys.stderr)
            return 1
        command = ["claude", "plugin", "marketplace", "add", marketplace_source]
        if not args.local:
            command.extend(["--sparse", ".claude-plugin", "plugins/dida-cli"])
        command.append("--json")
        run(command)
        run(["claude", "plugin", "install", "dida-cli@dida-cli", "-y", "--json"])
        verify_claude_plugin()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
