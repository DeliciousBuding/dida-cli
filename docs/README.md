# DidaCLI Documentation

This directory is the maintained documentation entrypoint for DidaCLI.

## Start Here

| Audience | Document | Purpose |
| --- | --- | --- |
| Users | [quickstart.md](quickstart.md) | Install, verify, authenticate, and run common commands. |
| Chinese users | [quickstart.zh-CN.md](quickstart.zh-CN.md) | 中文快速开始。 |
| LLM / Agent operators | [llm-quickstart.md](llm-quickstart.md) | Short JSON-first command path for agents. |
| Contributors | [commands.md](commands.md) | Command reference and safety rules. |
| Maintainers | [distribution.md](distribution.md) | Release and package-manager distribution plan. |

## API Channels

DidaCLI intentionally keeps three channels separate:

| Channel | Auth | Best for | Status |
| --- | --- | --- | --- |
| Web API | Browser cookie `t` | Broadest Dida365 account coverage | Primary working channel. |
| Official MCP | `DIDA365_TOKEN` or saved local token config | Official task/project/habit/focus tools | Token-based reads and task writes are live-smoked; habit/focus known-id writes need disposable targets. |
| Official OpenAPI | OAuth access token | Official REST project/task/focus/habit endpoints | Implemented locally; live resource calls need completed browser OAuth. |

## Coverage And Evidence

- [api-coverage.md](api-coverage.md) tracks command-level Web API coverage.
- [web-api.md](web-api.md) records private Web API endpoint notes.
- [research/api-channel-inventory.md](research/api-channel-inventory.md) explains channel boundaries.
- [research/official-channel-validation-matrix.md](research/official-channel-validation-matrix.md) tracks official MCP and OpenAPI verification.
- [research/roadmap-completion-audit.md](research/roadmap-completion-audit.md) is the current completion audit.
- [research/prompt-to-artifact-checklist.md](research/prompt-to-artifact-checklist.md) maps the active goal to artifacts and evidence.

## Current External Blockers

- Official OpenAPI requires a completed browser OAuth approval before live resource smokes.
- Additional Web API upload smokes are blocked on the observed account while quota reports `underQuota=false`; the verified comment attachment path remains implemented.
- Web API task-level attachment mutation remains research-only until association, download/preview, file-limit, and cleanup behavior are traced with a disposable task.
- Web API task activity detail is blocked by `need_pro` on the observed account.
- Official MCP habit/focus known-id reads and destructive writes need disposable habit/focus targets.

Do not commit cookies, tokens, OAuth secrets, local absolute paths, or full private payload dumps.
