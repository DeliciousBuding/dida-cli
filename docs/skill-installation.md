# Agent Skill Installation

DidaCLI ships a repo-local skill at:

```text
skills/dida-cli/SKILL.md
```

The skill is plain Markdown with standard front matter, so it can be copied or symlinked into agents that support local skills.

## Codex

```powershell
New-Item -ItemType Directory -Force $env:USERPROFILE\.codex\skills | Out-Null
Copy-Item -Recurse .\skills\dida-cli $env:USERPROFILE\.codex\skills\dida-cli -Force
```

Or keep it linked to the repo:

```powershell
New-Item -ItemType SymbolicLink `
  -Path $env:USERPROFILE\.codex\skills\dida-cli `
  -Target (Resolve-Path .\skills\dida-cli)
```

## Claude Code

Use the same skill folder if your Claude Code setup reads local skills:

```powershell
New-Item -ItemType Directory -Force $env:USERPROFILE\.claude\skills | Out-Null
Copy-Item -Recurse .\skills\dida-cli $env:USERPROFILE\.claude\skills\dida-cli -Force
```

If your setup uses a project-local skill directory, keep `skills/dida-cli/SKILL.md` in this repository and point Claude Code at the repo root.

## OpenClaw

Copy or mount the folder into OpenClaw's configured skills directory:

```bash
cp -R skills/dida-cli ~/.openclaw/skills/dida-cli
```

Then restart the OpenClaw process so it reloads skill metadata.

## Hermes Agent

Copy or mount the folder into Hermes' skills directory:

```bash
mkdir -p ~/.hermes/skills
cp -R skills/dida-cli ~/.hermes/skills/dida-cli
```

Then restart Hermes or the Hermes gateway session that should use the skill.

## Runtime Requirement

All agents need the `dida` binary on `PATH`:

```bash
dida doctor --json
dida auth status --verify --json
```

Authentication stays local in `~/.dida-cli/`. Do not paste cookies or token values into agent chat.
