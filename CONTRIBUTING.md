# Contributing to DidaCLI

Thanks for your interest in contributing. This document explains how to set up your development environment, what conventions to follow, and how to submit changes.

## Getting Started

```bash
git clone https://github.com/DeliciousBuding/dida-cli.git
cd dida-cli
go test ./...
go build -o bin/dida ./cmd/dida
```

**Requirements:** Go 1.26+, Python 3 with Playwright (only for browser login testing)

## Development Workflow

1. Fork the repository and create a branch from `main`
2. Make your changes
3. Run the verification suite:
   ```bash
   go test ./...
   go vet ./...
   go run golang.org/x/vuln/cmd/govulncheck@latest ./...
   ```
4. Commit with a [conventional commit](#commit-messages) message
5. Open a pull request

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>: <description>
```

| Type | When to use |
|---|---|
| `feat` | New command or feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `chore` | Build, CI, packaging, non-functional |
| `refactor` | Code restructuring without behavior change |
| `test` | Adding or updating tests |

Examples:
- `feat: add task batch-complete command`
- `fix: prevent help flags from triggering writes`
- `docs: update API coverage matrix`

## Pull Request Checklist

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] New/changed commands have schema entries in `internal/cli/schema_cmd.go`
- [ ] New/changed commands have help text in `internal/cli/help.go`
- [ ] Docs updated (`docs/commands.md`, `docs/api-coverage.md`, etc.)
- [ ] CHANGELOG.md updated for user-facing changes
- [ ] No cookies, tokens, or private response dumps committed

## Command Design Principles

- **Resource commands over generic tunnels:** prefer `task create` over raw POST.
- **Reads run directly:** no confirmation needed.
- **Writes support `--dry-run`:** preview payloads before executing.
- **Destructive writes require `--yes`:** delete, remove, etc.
- **JSON errors include context:** `type`, `message`, and a useful `hint`.
- **Bounded output:** list commands default to reasonable limits.
- **Token safety:** never print full tokens, redact by default.

## Project Layout

```
cmd/dida/           CLI entrypoint (main.go)
internal/auth/      Browser cookie capture, storage, redaction
internal/cli/       Command dispatch, JSON envelopes, help text, schemas
internal/config/    Config directory resolution
internal/model/     Normalized Task/Project/Column models
internal/webapi/    Dida365 Web API HTTP client
internal/officialmcp/ Official MCP protocol client
internal/openapi/   Official OpenAPI OAuth client
docs/               User and API documentation
skills/dida-cli/    Repo-local agent skill
packaging/          Homebrew, Scoop, winget templates
npm/                npm installer skeleton
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/model/...
go test ./internal/auth/...

# Verbose output
go test -v ./internal/webapi/...
```

Live API tests require a valid Dida365 session. Cookie-based tests should use disposable accounts and clean up after themselves.

## Questions?

Open a [discussion](https://github.com/DeliciousBuding/dida-cli/discussions) or check the [documentation](docs/README.md).
