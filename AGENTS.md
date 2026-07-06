# AGENTS.md

## Project Overview

DidaCLI is a JSON-first CLI for Dida365 / TickTick, built as a single Go binary with zero external Go module dependencies. It exposes three explicitly separated upstream channels: Web API (browser cookie auth, widest coverage), Official MCP (token-based `dp_...` auth), and Official OpenAPI (OAuth access token). Every command returns a stable JSON envelope, writes support `--dry-run` preview, and destructive actions require `--yes` confirmation. The project targets both human operators and AI agents as first-class users.

## Branching Strategy

- `main` is the release branch. All commits on `main` must pass CI (`go test ./...`, `go vet ./...`, `govulncheck ./...`, `check-private-state.sh`).
- Feature work branches off `main` with descriptive names (e.g. `feature/task-attachment-download`, `fix/upgrade-windows-lock`). Merge back via PR when done.
- Tags are semver `vX.Y.Z` and trigger the release workflow (build, sign, publish to GitHub Releases + npm).
- Never force-push to `main`. Never rebase shared branches.

## Code Conventions

- **Zero external deps**: `go.mod` must list only `go 1.26.x`. No third-party Go modules -- not even for HTTP, JSON, or CLI flag parsing. Everything is built on the standard library.
- **File organization**: `cmd/dida/` (entrypoint), `internal/cli/` (command definitions and output formatting), `internal/webapi/` (Web API HTTP client), `internal/officialmcp/` (official MCP client), `internal/openapi/` (OAuth + OpenAPI client), `internal/auth/` (cookie + browser auth), `internal/config/` (paths and local state), `internal/model/` (shared data types and normalization).
- **Naming**: packages use lowercase single-word or short compound names. Files per resource area (e.g. `tasks.go` + `tasks_test.go`). Test files sit alongside their subjects.
- **Output contracts**: all JSON output goes through `internal/cli/output.go` to guarantee a stable envelope (`ok`, `command`, `data`, `error`).
- **Channel isolation**: the three auth channels (Web API, Official MCP, Official OpenAPI) must never share credentials, HTTP clients, or token state. Each channel has its own package and its own `doctor` / `status` path.

## Testing Requirements

- Every new CLI command must add a test in `internal/cli/cli_test.go` covering at minimum: the dry-run path, the JSON output shape, and flag parsing.
- Every new Web API endpoint must add a request-shape unit test in the corresponding `internal/webapi/*_test.go` file.
- Before claiming an API surface is done, run a reversible live smoke test when possible (create → read-back → delete).
- Pre-commit checklist: `go test ./...`, `go vet ./...`, `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`.
- CI runs all tests with `-count=1` (no cache) on every push and PR.

## Security Rules

- **Never commit secrets**: `.env`, `.env.*`, `cookie.json`, `official-mcp-token.json`, `openapi-oauth.json`, `openapi-client.json`, and any file containing credentials or tokens must never enter the repo. `scripts/check-private-state.sh` enforces this -- run it before every commit.
- `scripts/check-private-state.sh` checks for: tracked private-state paths, cookie/token patterns (including `t=...`, `dp_...`, `Bearer ...`, JWT), OAuth secrets, and local absolute paths. CI runs this on every push and tag.
- `git diff --staged` self-review before committing: verify no secrets, no debug output, no local paths.
- Error messages and response bodies must redact tokens and sensitive patterns. The CLI never prints full cookie or token values.
- `data/private/` is gitignored and reserved for local-only research evidence. Never commit its contents.
- CodeQL and OpenSSF Scorecard workflows are part of repository governance. External workflow actions must be pinned to full commit SHAs with version comments, for example `owner/action@<40-char-sha> # vX`. If workflows change, run `bash scripts/validate-actions-pinned.sh`, `bash scripts/validate-repo-governance.sh`, and `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`.

## Commit Conventions

Use semantic prefixes aligned with the project's resource areas:

| Prefix | Scope |
| --- | --- |
| `feat:` | New command, new endpoint coverage, new channel capability |
| `fix:` | Bug fix, error handling, safety guard |
| `docs:` | README, docs/, skill, or research doc changes |
| `ci:` | Workflow, Makefile, script, or packaging changes |
| `test:` | Test-only changes (no production code) |
| `refactor:` | Internal restructure with no behavior change |

Do not batch unrelated work. Preferred commit shape: one resource area, one auth improvement, one doc batch, or one command family.

Examples: `feat: add task activity reads`, `fix: redact cookie in upgrade error path`, `docs: add official validation matrix`.

## PR / Review Requirements

- All changes go through PRs to `main`. Direct pushes to `main` are reserved for release tags and trivial doc fixes.
- PR checklist: (1) `go test ./...` passes, (2) `go vet ./...` clean, (3) `govulncheck` clean, (4) `check-private-state.sh` clean, (5) repository governance checks pass when workflows or public entry points change, (6) related docs updated (commands.md, api-coverage.md, SKILL.md as applicable).
- At least one approving review before merge. Reviewer checks: correctness, channel isolation, safety flags (`--dry-run` / `--yes`), test coverage, and doc consistency.
- Merged PRs should be squash-merged to keep `main` history linear and semantic-prefixed.

## Release Process

- Tag `main` with a semver tag: `git tag -a vX.Y.Z -m "vX.Y.Z"` and push. The tag must point to a commit reachable from `main`.
- Run `make release-check VERSION=vX.Y.Z` before pushing a release tag. This validates tag metadata, npm version alignment, changelog structure, npm package contents, pinned GitHub Actions, repository governance files, current package-manager template metadata, helper scripts, and workflow syntax without publishing.
- Tag push triggers `.github/workflows/release.yml`: validate → test + vet + vulncheck + private-state check → multi-platform build (6 targets) → npm preflight → GitHub Release with checksums → npm install smoke → npm publish.
- Prefer npm Trusted Publishing/OIDC for npm releases. Keep `NPM_TOKEN` only as a fallback until the npm package trusted publisher is configured and proven by a real release.
- Before tagging: update `CHANGELOG.md` with a `## [vX.Y.Z]` section, bump version in `npm/package.json`, and confirm CI is green on `main`.
- Release notes are auto-generated from `CHANGELOG.md`. Use `workflow_dispatch` with `allow_changelog_fallback=true` only for emergency releases.
- After release: the `dida upgrade` command uses the GitHub Releases API for self-update with SHA-256 verification. Update Homebrew and Scoop templates only after the release assets and `checksums.txt` exist.
- Verify npm publishes against `https://registry.npmjs.org`; local mirrors such as npmmirror can lag behind `latest`.

## Core Rules

禁止提交 `.env`、凭据、私钥。提交前 `git diff --staged` 自查。
