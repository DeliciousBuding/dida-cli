# CLAUDE.md

Claude Code 专用指令文件。跨 agent 共享规则见 `AGENTS.md`。

## Architecture

DidaCLI is a single Go binary (`cmd/dida/main.go`) with zero external Go module dependencies (standard library only). It exposes three isolated auth channels, each in its own package:

| Channel | Auth | Package | Role |
| --- | --- | --- | --- |
| Web API | Browser cookie `t` | `internal/webapi/` | Primary: widest task/project/comment/sync coverage |
| Official MCP | `DIDA365_TOKEN=dp_...` or saved config | `internal/officialmcp/` | Token-based: project/task/habit/focus wrappers + generic `call` |
| Official OpenAPI | OAuth access token | `internal/openapi/` | Standard REST: project/task/focus/habit CRUD |

Shared packages: `internal/cli/` (command graph, output envelope, schema registry), `internal/auth/` (cookie + browser login), `internal/config/` (local paths), `internal/model/` (types and normalization).

The output contract: all JSON goes through `internal/cli/output.go` -- stable envelope with `ok`, `command`, `data`, `error`.

## Build / Run / Test

```bash
go build -o bin/dida ./cmd/dida          # build
./bin/dida doctor --json                  # smoke
go test ./...                             # all tests
go vet ./...                              # vet
make staticcheck                          # Staticcheck
bash scripts/check-private-state.sh       # secret leak check
```

Makefile targets: `make test`, `make build`, `make install-local`, `make coverage-cli`, `make staticcheck`.

## Key Files To Read When Starting Work

| File | Why |
| --- | --- |
| `ROADMAP.md` | Current direction, workstreams, and completion criteria |
| `AGENTS.md` | Cross-agent governance (branching, conventions, release) |
| `skills/dida-cli/SKILL.md` | Agent skill -- safety rules, channel selection, command patterns |
| `docs/commands.md` | User-facing command reference |
| `docs/api-coverage.md` | Web API coverage matrix with verification status |
| `docs/archives/release-governance-optimization/progress/MASTER.md` | Completed release-governance tracker |
| `internal/cli/cli_test.go` | Integration-level command tests |
| `cmd/dida/main.go` | Entrypoint and version injection |

## Rules

### Validation Before Auth

Before using any channel, verify auth is live: `dida doctor --verify --json` (Web API), `dida official doctor --json` (MCP), `dida openapi doctor --json` (OpenAPI). Do not assume auth works across sessions.

### Dry-Run For Writes

All write commands support `--dry-run`. Preview before executing. The agent skill (`skills/dida-cli/SKILL.md`) mandates: always dry-run first, show the payload, get confirmation, then execute. Destructive actions (delete, merge) also require `--yes`.

### No Secret Leaks

- Never commit `.env`, cookie files, token files, OAuth configs, or local absolute paths.
- `scripts/check-private-state.sh` is the enforcement gate -- run it before commits. CI runs it on every push.
- Cookie and token values are redacted in all error messages and output.
- `data/private/` is gitignored for local research evidence. Never commit its contents.

### Channel Isolation

The three channels must never share credentials, HTTP clients, or token state. Web API cookie `t` cannot authenticate MCP or OpenAPI calls. MCP `dp_...` tokens cannot authenticate Web API or OpenAPI calls. OpenAPI OAuth tokens are channel-specific.

## Memory

No active spec-driven tracker is open. The completed release-governance tracker is archived at `docs/archives/release-governance-optimization/progress/MASTER.md`.

When starting a new spec-driven run, create a fresh tracker under `docs/progress/MASTER.md` and archive it when the run completes. Do not edit archived trackers for new work.
