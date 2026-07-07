# Release Strategy: GoReleaser

## Decision

Keep the current hand-written release workflow for `v0.3.x`. Do not migrate to GoReleaser in the next milestone.

## Date

2026-07-07

## Current Release Workflow

The existing workflow already covers the release requirements that are easy to
lose during a migration:

- tag validation, changelog validation, npm version validation, tests, vet,
  Staticcheck, vulnerability scanning, pinned-action checks, and private-state
  scanning
- six release archives for Windows, Linux, and macOS on amd64 and arm64
- `checksums.txt`
- GitHub artifact attestations from `dist/checksums.txt`
- npm install smoke tests on Linux and Windows
- npm publish with provenance, using Trusted Publishing when `NPM_TOKEN` is not
  available
- Homebrew and Scoop templates regenerated from release checksums

## GoReleaser Fit

GoReleaser would help if DidaCLI moves package-manager publishing into the tag
release job. Its docs cover GitHub Actions usage, checksum generation,
Homebrew/Scoop publishing, and cross-repository package-manager updates.

That value is not enough to replace the current workflow yet. The current
workflow has project-specific gates and npm behavior that would need a parity
plan before a migration is safe.

## Risks

| Area | Current workflow | GoReleaser migration risk |
|---|---|---|
| npm provenance | Explicit npm package validation, smoke tests, and `npm publish --provenance` | Needs separate npm wiring or a verified GoReleaser npm path |
| Attestations | GitHub artifact attestation runs on `dist/checksums.txt` | Must prove equivalent artifact identity and verification commands |
| Package-manager repos | Templates are generated locally; external publication is still manual | Cross-repo Homebrew/Scoop publishing needs a token beyond default `GITHUB_TOKEN` |
| Action pinning | All external actions are pinned to full SHAs | GoReleaser action must be pinned and kept current |
| Release notes | Generated from `CHANGELOG.md` with tested fallback rules | Must preserve current changelog rules and manual-dispatch behavior |

## Revisit Conditions

Re-evaluate GoReleaser only after all of these are true:

1. `homebrew-tap` and `scoop-bucket` repositories exist or the project decides
   to keep those submissions manual.
2. A GoReleaser dry run produces the same archive names, binary paths,
   checksums, and release-note shape as the current workflow.
3. npm publishing keeps the same README, package contents, install smoke tests,
   and provenance behavior.
4. GitHub artifact attestation or an approved replacement remains available for
   release archives.
5. Any cross-repository publishing token is scoped, documented, and optional for
   normal GitHub Release plus npm publication.

## Next Action

Keep improving the current release workflow. Use
`scripts/update-packaging-templates.sh` for package-manager templates and build
external Homebrew/Scoop publication as a separate, reviewable step.
