# Security Policy

## Supported Versions

Security fixes target the current `main` branch and the latest `v0.2.x`
release.

## Reporting a Vulnerability

Please open a private security advisory on GitHub or contact the maintainer privately.

Do not include live Dida365 cookies, bearer tokens, private task exports, or full response dumps in public issues.

## Security Baseline

Last reviewed: 2026-10-23.

- Go dependency surface: standard library only; `go.mod` is the version source.
- npm package: no runtime dependencies.
- GitHub Actions: pinned to full commit SHAs with release comments.
- Required pull-request checks: tests, vet, Staticcheck, `govulncheck`,
  private-state scan, CodeQL, and agent-plugin manifest validation. OpenSSF
  Scorecard runs on `main` and on a schedule.
- `main` rejects force pushes and deletion, requires linear history, and
  requires the pull-request checks above. While the project has one active
  maintainer, pull requests use agent-assisted review rather than a second
  human approval.
- Release publication: tag-triggered only after `make release-check VERSION=...`.

The update cadence, CVE response targets, release gates, and trust boundaries
are defined in
[`docs/security/dependency-and-release-policy.md`](docs/security/dependency-and-release-policy.md).

## Credential Handling

DidaCLI stores credentials under `~/.dida-cli/` and redacts token previews in CLI output. The repository ignores local env files, private data, and generated binaries.

Agents and scripts should import cookies through stdin or browser capture. They should never ask users to paste secrets into chat.

## Agent Plugins

The bundled Codex/Claude Code plugin declares the official remote Dida365 MCP
server but does not bundle credentials or a local privileged executable.
Plugin writes still go through the host's approval UI and DidaCLI's own
`--dry-run`/`--yes` boundaries.
