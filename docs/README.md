# DidaCLI docs

This directory is the maintained documentation entrypoint for DidaCLI.

## Start here

| Audience | Document | Purpose |
| --- | --- | --- |
| Users | [quickstart.md](quickstart.md) | Install, verify, authenticate, and run common commands. |
| Chinese users | [quickstart.zh-CN.md](quickstart.zh-CN.md) | 中文快速开始。 |
| Attachment download (中文) | [attachment-download.md](attachment-download.md) | `dida attachment download` 功能说明与用法。 |
| OpenAPI operators | [openapi-setup.md](openapi-setup.md) | Configure `client_id`, `client_secret`, redirect URL, and OAuth login. |
| 中文 OpenAPI 配置 | [openapi-setup.zh-CN.md](openapi-setup.zh-CN.md) | 官方 OpenAPI 设置与回调指南。 |
| LLM / Agent operators | [llm-quickstart.md](llm-quickstart.md) | Short JSON-first command path for agents. |
| Contributors | [commands.md](commands.md) | Command reference and safety rules. |
| Maintainers | [distribution.md](distribution.md) | Release and package-manager distribution plan. |

## API channels

DidaCLI exposes three separate API channels:

| Channel | Auth | Best for | Status |
| --- | --- | --- | --- |
| Web API | Browser cookie `t` | Web API resources outside the official channels | Primary working channel. |
| Official MCP | `DIDA365_TOKEN` or configured token store | Official task/project/habit/focus tools | Token-based project/task reads and task writes are live-smoked; habit list, sections, and focus range reads are verified. |
| Official OpenAPI | OAuth access token | Official REST project/task/focus/habit endpoints | OAuth login and project reads are live-verified; write smokes and known-id habit/focus paths need disposable targets. |

## Coverage and evidence

- [api-coverage.md](api-coverage.md) tracks command-level Web API coverage.
- [web-api.md](web-api.md) records private Web API endpoint notes.
- [research/api-channel-inventory.md](research/api-channel-inventory.md) explains channel boundaries.
- [research/official-channel-validation-matrix.md](research/official-channel-validation-matrix.md) tracks official MCP and OpenAPI verification.
- [research/roadmap-completion-audit.md](research/roadmap-completion-audit.md) maps the active goal to evidence and blockers.

## Current blockers

- Official OpenAPI write smokes need disposable project/task targets; habit and focus known-id checks need test records.
- Additional Web API upload smokes need an account with attachment quota available; the verified comment attachment path remains implemented.
- Web API task-level attachment mutation remains research-only until association, download/preview, file-limit, and cleanup behavior are traced with a disposable task.
- Web API task activity detail needs a Pro account or a browser trace that confirms the response shape.
- Official MCP habit get/checkins need known habit IDs; focus get/delete need focus records and disposable targets.

Do not commit cookies, tokens, OAuth secrets, local absolute paths, or full private payload dumps.
