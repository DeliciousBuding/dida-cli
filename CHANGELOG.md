# Changelog

All notable changes to DidaCLI are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Version numbers follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `dida upgrade` self-update command: checks GitHub Releases, downloads platform binary, verifies SHA-256, replaces executable
- `dida upgrade --check` mode for version query without install
- Download progress output on stderr during upgrade (percentage-based)
- npm auto-publish job in release workflow (publishes on tag push)
- `upgrade` registered in schema registry
- Agent Safety section in README (EN + ZH) with 5 clear safety rules
- Supported Platforms table showing all 6 build targets in README

### Changed
- Quick Start flow: `dida auth cookie set --token-stdin` is now primary; `--browser` is secondary
- Auth error hints now recommend `--token-stdin` instead of browser-only flow
- Interactive terminal hint when using `--token-stdin` (prompts for Ctrl+D/Ctrl+Z)

### Fixed
- OpenAPI OAuth `ExchangeCode` now uses 30s timeout instead of `http.DefaultClient`
- OpenAPI OAuth error messages no longer leak raw response body (secret leakage prevention)
- Official MCP client now uses 60s timeout instead of `http.DefaultClient`
- `dida upgrade` now fails when `checksums.txt` is missing from release assets (previously skipped verification silently)
- Release notes template uses `__VER__` placeholder instead of generic `VERSION` to avoid unintended sed replacements
- Semver comparison in upgrade uses proper numeric comparison (fixes downgrade-as-upgrade bug)
- Web API client: regex patterns hoisted to package-level (avoids recompilation per call)

### Tests
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
- Comprehensive schema coverage for all commands

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
