# Changelog

All notable changes to DidaCLI are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Version numbers follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `dida task latest` reads the latest synced tasks by creation time, with inbox alias support and compact JSON output.
- Release governance helpers validate semver tags, npm package version alignment, changelog sections, and generated GitHub Release notes before tag publishing.
- Dependabot monitors GitHub Actions and the npm wrapper package for dependency updates.

### Changed
- CI stores coverage profiles under `coverage/profile.txt` and runs coverage commands through Bash on all test runners for cross-platform consistency.
- Release notes generation is now a tested script instead of inline workflow logic.
- Changelog and npm package contents are now validated by reusable scripts in CI and release preflight.
- npm publishing now prefers Trusted Publishing/OIDC when `NPM_TOKEN` is absent, while retaining token auth as a fallback.

### Fixed
- `dida task list --filter <invalid>` now returns a validation error before auth or sync loading.

## [v0.2.4] - 2026-06-24

### Added
- `dida attachment download` — download existing task attachments via the Web API v1 download endpoint. Supports `--project`, `--task`, `--attachment`, `--output`, and `--force`. Uses atomic temp-file writes and refuses to overwrite local files by default.

### Fixed
- Release workflow shell lint issues are fixed for formatting, checksum generation, release notes, and npm smoke packaging.
- npm release authentication now fails with a clear registry error before GitHub release creation.

## [v0.2.3] - 2026-06-09

### Fixed
- Go formatting baseline now matches Linux CI `gofmt`.

## [v0.2.2] - 2026-06-09

### Added
- CI coverage for private-state checks, packaging metadata validation, release archive verification, workflow linting, cross-platform builds, Go vet, and vulnerability scanning.
- Release script tests for packaging metadata, private-state hygiene, archive shape, archive contents, and executable bits.
- npm installer offline tests for redirects, checksum matching, byte limits, retry handling, unsupported platforms, and asset parsing.

### Changed
- `dida upgrade` downloads the archive and checksum file concurrently, retries transient artifact downloads, enforces a 200 MiB artifact cap, and reports Windows deferred replacement as `status: "scheduled"`.
- The npm installer now retries transient downloads, enforces a 200 MiB response cap, and reports final download attempts in errors.
- Release workflow now builds six archives in parallel, verifies archive contents before publishing, smoke-tests the Linux binary, and skips npm publish when the package version already exists.

### Fixed
- Web API task, project, comment, column, folder, and tag commands now reject flag-like IDs before auth, API calls, or dry-run success output.
- OpenAPI and Official command wrappers now reject flag-like task and project IDs in local validation.
- Official habit check-in value parsing now rejects `NaN` and infinite values.
- Windows self-upgrade stages the replacement through a helper script instead of attempting to replace the running executable.
- Packaging and release scripts now clean temporary files on failure paths covered by the test suite.

## [v0.2.1] - 2026-05-12

### Added
- `dida upgrade` self-update command: checks GitHub Releases, downloads platform binary, verifies SHA-256, replaces executable
- `dida upgrade --check` mode for version query without install
- Download progress output on stderr during upgrade (percentage-based)
- npm auto-publish job in release workflow (publishes on tag push)
- `upgrade` registered in schema registry
- README safety note for dry-run previews, destructive confirmations, and token redaction
- Release workflow archive verification for all six build targets

### Changed
- Quick Start flow: `dida auth cookie set --token-stdin` is now primary; `--browser` is secondary
- Auth error hints now recommend `--token-stdin` instead of browser-only flow
- Interactive terminal hint when using `--token-stdin` (prompts for Ctrl+D/Ctrl+Z)
- `attachment quota` now fetches independent quota endpoints in parallel
- README install snippets now use copy-pasteable URLs and `upgrade --check` for post-install update checks
- npm installer now defaults to the GitHub Release tag that matches the npm package version

### Fixed
- Global `--json` is accepted before root `version` and `--help`
- CLI integer flags now reject trailing junk such as `--limit 10x`
- Auth commands now reject unknown flags instead of silently ignoring them
- `comment update --file` now fails validation because update does not attach files
- OpenAPI project and Official task commands now return local usage errors before auth errors
- Release workflow now validates tag checkout, checks archive contents, verifies binary version, and runs npm preflight before creating the GitHub release
- Private-state checks now cover untracked files, local DidaCLI credential filenames, common dumps, packet captures, and local absolute paths
- OpenAPI OAuth `ExchangeCode` now uses 30s timeout instead of `http.DefaultClient`
- OpenAPI OAuth error messages no longer leak raw response body (secret leakage prevention)
- Official MCP client now uses 60s timeout instead of `http.DefaultClient`
- `dida upgrade` now fails when `checksums.txt` is missing from release assets (previously skipped verification silently)
- Release notes template uses `__VER__` placeholder instead of generic `VERSION` to avoid unintended sed replacements
- Semver comparison in upgrade uses proper numeric comparison (fixes downgrade-as-upgrade bug)
- Web API client: regex patterns hoisted to package-level (avoids recompilation per call)

### Tests
- Added CLI regression tests for root global flags, auth unknown flags, comment update file rejection, strict integer parsing, and auth-before-validation ordering
- Added Web API coverage proving `AttachmentQuota` overlaps independent reads
- Upgrade integration tests: full flow mock, missing checksums failure, checksum mismatch failure
- Progress reader unit test
- config package: 0% to 83.3%
- officialmcp package: 39.9% to 85.0%
- openapi package: 49.6% to 82.7%
- model package: 62.9% to 91.0%
- auth package: 48.8% to 66.7%
- cli package: 32.3% to 37.8%
- webapi package: 80.5% to 84.2%

## [v0.2.0] - 2026-05-10

Major milestone: complete three-channel support with OpenAPI OAuth.

### Added
- OpenAPI OAuth browser login flow with `openapi login --browser`
- OpenAPI callback listener with loopback host validation
- OpenAPI token persistence and status via `openapi status`
- OpenAPI project/task/focus/habit command wrappers with local dry-run
- Compact schema index for faster agent command discovery
- Channel selection guide for choosing between Web API / MCP / OpenAPI
- OpenAPI setup documentation (English and Chinese)

### Fixed
- OAuth callback error handling and browser login hardening
- Schema discovery and OAuth browser error messages

### Docs
- OpenAPI live verification evidence
- Agent channel alignment guidance

## [v0.1.16] - 2026-05-09

Packaging fixes and distribution hardening.

### Fixed
- Homebrew binary path in release archive
- npm installer checksum resolution
- Auth error messages now point to browser login
- OpenAPI login recommends browser flow

### Docs
- npm installer environment variable documentation
- Scoop archive layout verification
- Official MCP dry-run warning for `official call`

## [v0.1.15] - 2026-05-08

Schema and help quality improvements.

### Added
- Comment attachment live smoke evidence
- Schema coverage for all commands

### Fixed
- Dry-run flag advertised in schema metadata
- Help flags no longer trigger write side-effects
- Task and comment subcommand help display

## [v0.1.14] - 2026-05-07

### Fixed
- Agent context help output

## [v0.1.13] - 2026-05-06

### Added
- Outline mode for `agent context` -- returns task ID references with deduplicated taskIndex for lower token usage

## [v0.1.12] - 2026-05-05

OpenAPI foundation and documentation index.

### Added
- OpenAPI auth schema entries for `openapi login/auth-url/exchange-code`
- OpenAPI login timeout details in error output
- Documentation index at `docs/README.md`

### Fixed
- Official MCP task update wrapping

### Docs
- Winget packaging notes
- Attachment quota blocker documentation

## [v0.1.10] - 2026-05-03

Comment attachment upload support.

### Added
- `comment create --file <path>` for multipart attachment upload
- Attachment quota check before upload
- Verified comment attachment flow with live smoke tests

## [v0.1.8] - 2026-04-30

Official MCP token management.

### Added
- `official token set/status/clear` for local official token config
- Token persistence under `~/.dida-cli/`

## [v0.1.5] - 2026-04-26

Doctor verification and packaging.

### Added
- `doctor --verify` endpoint verification
- Raw probe error detail exposure
- Package manager templates (Homebrew, Scoop)

### Fixed
- npm Unix wrapper stability
- Schema auth metadata accuracy

## [v0.1.0] - 2026-04-20

Initial release.

### Features
- Web API channel with browser cookie auth
- Task CRUD: today, list, search, upcoming, create, update, complete, delete, move
- Project / folder / tag / column / comment management
- Completed history, closed history, trash browsing
- Pomodoro and habit reads
- Sharing, calendar, statistics, templates, search metadata
- `raw get` read-only probe
- `schema list/show` for machine-readable command discovery
- `agent context` for compact agent context packs
- `sync all/checkpoint` for full and incremental sync
- `--json` output with consistent envelope `{ ok, command, meta, data, error }`
- `--dry-run` for all write commands
- `--yes` required for destructive operations
- Token redaction in output
- GitHub Actions CI (test, vet, govulncheck)
- Tag-triggered release workflow with six-platform binaries
- Install scripts for macOS, Linux, and Windows
- MIT License

[Unreleased]: https://github.com/DeliciousBuding/dida-cli/compare/v0.2.4...HEAD
[v0.2.4]: https://github.com/DeliciousBuding/dida-cli/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/DeliciousBuding/dida-cli/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/DeliciousBuding/dida-cli/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/DeliciousBuding/dida-cli/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.16...v0.2.0
[v0.1.16]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.15...v0.1.16
[v0.1.15]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.14...v0.1.15
[v0.1.14]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.13...v0.1.14
[v0.1.13]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.12...v0.1.13
[v0.1.12]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.10...v0.1.12
[v0.1.10]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.8...v0.1.10
[v0.1.8]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.5...v0.1.8
[v0.1.5]: https://github.com/DeliciousBuding/dida-cli/compare/v0.1.0...v0.1.5
[v0.1.0]: https://github.com/DeliciousBuding/dida-cli/releases/tag/v0.1.0
