# LLM / Agent Quickstart

This page is for LLMs and automation agents. It is not a human tutorial.
Prefer copy-pasteable JSON commands and do not ask users to paste secrets into
chat.

## First Commands

```bash
dida doctor --json
dida schema list --compact --json
dida channel list --json
dida agent context --outline --json
```

If auth is missing or expired, tell the user to run this locally:

```bash
dida auth login --browser --json
```

Then verify:

```bash
dida doctor --verify --json
dida auth status --verify --json
```

## Channel Boundaries

Do not mix these channels:

```bash
# Web API: browser cookie stored locally
dida agent context --json

# Official MCP: token from Dida365 account settings
DIDA365_TOKEN=... dida official doctor --json
dida official token status --json
dida official project list --json
dida official task get <task-id> --project <project-id> --json
dida official task query --query today --json
dida official habit list --json
dida official task batch-add --args-json '{"tasks":[{"title":"Agent task"}]}' --dry-run --json

# Official OpenAPI: OAuth REST
dida openapi doctor --json
dida openapi client status --json
dida openapi login --browser --json
dida openapi project list --json
dida openapi project create --args-json '{"name":"Project","viewMode":"list","kind":"TASK"}' --dry-run --json
dida openapi habit list --json
dida openapi habit checkin <habit-id> --args-json '{"stamp":20260407,"value":1}' --dry-run --json
```

For OpenAPI, read `default_redirect_uri` and `next` from `dida openapi doctor
--json`. Ask the user to configure that exact redirect URL in the Dida365
developer app before OAuth login. Do not treat the MCP `DIDA365_TOKEN` as an
OpenAPI bearer token.

## Discover Before Writing

```bash
dida schema list --compact --json
dida schema show task.create --json
dida schema show task.delete --json
```

Read `authRequired`, `dryRun`, and `confirmationRequired` before choosing a
command. Most commands under `official.*` need `DIDA365_TOKEN` or a saved
official token from `dida official token set --token-stdin`; OpenAPI resource
commands need a saved OAuth access token. Do not treat Web API cookie auth as
valid for either official channel.

## Safe Reads

```bash
dida project list --json
dida task today --compact --json
dida task upcoming --days 14 --limit 50 --compact --json
dida task search --query "keyword" --limit 20 --compact --json
dida completed today --compact --json
```

## Safe Writes

Preview generated writes first:

```bash
dida task create --project <project-id> --title "Agent task" --dry-run --json
dida task update <task-id> --project <project-id> --title "New title" --dry-run --json
dida official task batch-add --args-json '{"tasks":[{"title":"Agent task"}]}' --dry-run --json
dida official task complete-project --project <project-id> --task <task-id> --dry-run --json
```

Execute only after the preview matches the user's request:

```bash
dida task create --project <project-id> --title "Agent task" --json
```

Deletes require explicit confirmation:

```bash
dida task delete <task-id> --project <project-id> --dry-run --json
dida task delete <task-id> --project <project-id> --yes --json
```

## Never Do This

- Do not ask the user to paste cookies or tokens into chat.
- Do not use official MCP tokens as OpenAPI bearer tokens.
- Do not use Web API cookies as official OpenAPI tokens.
- Do not use `dida official call` for write-capable MCP tools unless the user
  explicitly approves that exact tool and payload; it has no dry-run layer.
- If Official MCP auth is missing, ask the user to run `dida official token set --token-stdin --json` locally.
- If OpenAPI client config is missing, ask the user to run `dida openapi client set --id <client-id> --secret-stdin --json` locally.
- Do not create live disposable habits just to test Official MCP/OpenAPI habit
  writes unless the user has approved a cleanup or archive procedure.
- Do not run destructive commands without `--yes`.
- Do not use raw probes for writes; DidaCLI intentionally supports only raw
  read probes.
