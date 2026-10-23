# Dependency and Release Security Policy

Last verified: 2026-10-23.

This document is the implementation policy for dependency drift, vulnerability
response, and release publication. Runtime facts remain sourced from `go.mod`,
the Makefile, workflows, and release metadata; this page explains how those
sources are governed.

## Dependency Inventory

| Surface | Current design | Source of truth |
| --- | --- | --- |
| Go application | Standard library only | `go.mod` |
| npm wrapper | No runtime dependencies; downloads the release binary | `npm/package.json` |
| Verification tools | Staticcheck, govulncheck, actionlint | `Makefile` and workflow `env` |
| GitHub Actions | Pinned to full commit SHAs | `.github/workflows/*.yml` |
| Agent plugin | Skill plus remote official MCP server; no bundled token or local server | `plugins/dida-cli/` |

The project deliberately avoids a dependency lockfile for the Go binary because
it has no external Go modules.

## Update Cadence

- Dependabot version updates run monthly and group minor/patch Actions changes.
- Dependabot alerts and security updates remain enabled independently of the
  monthly version-update schedule.
- High and critical vulnerabilities should be patched within 7 days.
- Other actionable vulnerabilities should be patched within 30 days.
- Tool upgrades are accepted only when the project's supported Go 1.26 line
  remains valid or an explicit compatibility migration is approved.

## Verification Gates

Every change runs:

```bash
GOTOOLCHAIN=go1.26.10 go test -race -count=1 ./...
GOTOOLCHAIN=go1.26.10 go vet ./...
GOTOOLCHAIN=go1.26.10 go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
GOTOOLCHAIN=go1.26.10 go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
bash scripts/check-private-state.sh
python3 scripts/validate-agent-plugins.py
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
```

CodeQL and OpenSSF Scorecard run independently on `main` and pull requests. A
release additionally runs `make release-check VERSION=vX.Y.Z`.

## Release Boundary

- Releases are created only from a commit reachable from `main`.
- The version must agree across the changelog, npm package metadata, release
  metadata, packaging templates, and website metadata.
- Build archives include checksums and artifact attestations.
- npm publish runs only from the tag-triggered release workflow.
- npm publication uses GitHub Actions OIDC through a package-scoped Trusted
  Publisher; no long-lived npm token is configured in the repository.
- A release is not considered published to npm until the official registry
  exposes the exact version and promotes it to the `latest` dist-tag.
- Pull requests never receive publishing credentials.

## Response Procedure

1. Reproduce with the current supported Go minor line.
2. Check whether the finding is in application code, the Go toolchain, an
   action, or the npm wrapper.
3. Prefer the smallest supported patch upgrade.
4. Re-run the full gate and preserve the scanner output in the PR or advisory.
5. If no safe upgrade exists, document the affected path, compensating control,
   owner, and expiry date instead of silently ignoring the finding.
