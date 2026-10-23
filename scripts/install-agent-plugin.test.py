#!/usr/bin/env python3
"""Unit tests for the native Codex and Claude Code plugin installer."""

from __future__ import annotations

import importlib.util
import sys
import unittest
from pathlib import Path
from unittest import mock


ROOT = Path(__file__).resolve().parent.parent
INSTALLER_PATH = ROOT / "scripts" / "install-agent-plugin.py"
SPEC = importlib.util.spec_from_file_location("install_agent_plugin", INSTALLER_PATH)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError(f"cannot load {INSTALLER_PATH}")
INSTALLER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(INSTALLER)


def inventory_for(command: list[str]) -> object:
    if command[0] == "codex":
        if command[1:3] == ["mcp", "list"]:
            return [
                {
                    "name": "dida365",
                    "enabled": True,
                    "transport": {"url": INSTALLER.DIDA_MCP_URL},
                    "auth_status": "not_logged_in",
                }
            ]
        return {
            "installed": [
                {
                    "pluginId": "dida-cli@dida-cli",
                    "enabled": True,
                }
            ]
        }
    return [
        {
            "id": "dida-cli@dida-cli",
            "enabled": True,
            "mcpServers": {
                "dida365": {"type": "http", "url": INSTALLER.DIDA_MCP_URL}
            },
        }
    ]


def invoke(args: list[str]) -> tuple[int, list[list[str]]]:
    commands: list[list[str]] = []

    def record(command: list[str]) -> None:
        commands.append(command)

    with (
        mock.patch.object(sys, "argv", ["install-agent-plugin.py", *args]),
        mock.patch.object(
            INSTALLER.shutil,
            "which",
            side_effect=lambda command: f"/usr/bin/{command}",
        ),
        mock.patch.object(INSTALLER, "run", side_effect=record),
        mock.patch.object(INSTALLER, "run_json", side_effect=inventory_for),
    ):
        code = INSTALLER.main()
    return code, commands


class InstallAgentPluginTests(unittest.TestCase):
    def test_codex_remote_is_default(self) -> None:
        code, commands = invoke(["codex", "--no-cli"])
        self.assertEqual(code, 0)
        self.assertEqual(
            commands[0],
            [
                "codex",
                "plugin",
                "marketplace",
                "add",
                "DeliciousBuding/dida-cli",
                "--sparse",
                ".agents/plugins",
                "--sparse",
                "plugins/dida-cli",
                "--json",
            ],
        )

    def test_codex_local_uses_checkout_without_sparse(self) -> None:
        code, commands = invoke(["codex", "--local", "--no-cli"])
        self.assertEqual(code, 0)
        self.assertEqual(commands[0][2:5], ["marketplace", "add", str(ROOT)])
        self.assertNotIn("--sparse", commands[0])

    def test_claude_remote_uses_client_sparse_shape(self) -> None:
        code, commands = invoke(["claude", "--no-cli"])
        self.assertEqual(code, 0)
        self.assertEqual(
            commands[0],
            [
                "claude",
                "plugin",
                "marketplace",
                "add",
                "DeliciousBuding/dida-cli",
                "--sparse",
                ".claude-plugin",
                "plugins/dida-cli",
                "--json",
            ],
        )

    def test_custom_source(self) -> None:
        code, commands = invoke(
            ["claude", "--source", "example/plugins@v1.2.3", "--no-cli"]
        )
        self.assertEqual(code, 0)
        self.assertIn("example/plugins@v1.2.3", commands[0])

    def test_fresh_npm_bootstrap_updates_cli(self) -> None:
        commands: list[list[str]] = []
        which_results = iter(
            [None, "/usr/bin/npm", "/usr/bin/dida", "/usr/bin/codex"]
        )

        def record(command: list[str]) -> None:
            commands.append(command)

        with (
            mock.patch.object(
                sys, "argv", ["install-agent-plugin.py", "codex"]
            ),
            mock.patch.object(
                INSTALLER.shutil, "which", side_effect=lambda _: next(which_results)
            ),
            mock.patch.object(INSTALLER, "run", side_effect=record),
            mock.patch.object(INSTALLER, "run_json", side_effect=inventory_for),
        ):
            code = INSTALLER.main()

        self.assertEqual(code, 0)
        self.assertEqual(
            commands[:2],
            [
                ["npm", "install", "-g", "@delicious233/dida-cli"],
                ["dida", "upgrade", "--json"],
            ],
        )

    def test_codex_rejects_missing_mcp_registration(self) -> None:
        def responses(command: list[str]) -> object:
            if command[1:3] == ["mcp", "list"]:
                return []
            return inventory_for(command)

        with (
            mock.patch.object(
                sys, "argv", ["install-agent-plugin.py", "codex", "--no-cli"]
            ),
            mock.patch.object(
                INSTALLER.shutil, "which", side_effect=lambda _: "/usr/bin/client"
            ),
            mock.patch.object(INSTALLER, "run"),
            mock.patch.object(INSTALLER, "run_json", side_effect=responses),
            self.assertRaises(SystemExit),
        ):
            INSTALLER.main()

    def test_claude_rejects_wrong_mcp_endpoint(self) -> None:
        def responses(_: list[str]) -> object:
            return [
                {
                    "id": "dida-cli@dida-cli",
                    "enabled": True,
                    "mcpServers": {
                        "dida365": {
                            "type": "http",
                            "url": "https://example.com/mcp",
                        }
                    },
                }
            ]

        with (
            mock.patch.object(
                sys, "argv", ["install-agent-plugin.py", "claude", "--no-cli"]
            ),
            mock.patch.object(
                INSTALLER.shutil, "which", side_effect=lambda _: "/usr/bin/client"
            ),
            mock.patch.object(INSTALLER, "run"),
            mock.patch.object(INSTALLER, "run_json", side_effect=responses),
            self.assertRaises(SystemExit),
        ):
            INSTALLER.main()


if __name__ == "__main__":
    unittest.main()
