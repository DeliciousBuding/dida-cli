# Rust Rewrite Compatibility Test Plan

This plan defines the first compatibility gate for comparing the legacy Go
`dida` binary with the Rust rewrite. The gate is intentionally local-only:
cases must not require account credentials, saved config, network access, or
live Dida365/TickTick API calls.

## Goals

- Detect user-visible CLI contract drift before the Rust binary is promoted.
- Compare exit code, stdout, and stderr for safe command surfaces.
- Keep fixtures and scripts independent of Go internals so the same harness can
  run against any two binaries.
- Make it easy to expand the matrix as Rust command coverage grows.

## Non-Goals

- No live account/API verification.
- No mutation of real local DidaCLI config.
- No comparison of implementation-only behavior or private package APIs.
- No CLI implementation edits from this harness.

## Safe Command Scope

Initial cases cover:

- Version and root help: `version`, `--version`, `--help`, and no arguments.
- Global JSON/help parsing, including `--json` with no command.
- Focused subcommand help for high-traffic auth/task surfaces.
- Local schema discovery: `schema list`, filtered `schema list`, and
  `schema show`.
- Local channel guide: `channel list`.
- Local doctor status without endpoint verification.
- Dry-run write previews that should stop before auth or network calls.
- Parser and validation failures that should be reported before auth checks.
- Missing-auth read flows that should stop before network calls.

Commands are executed with an isolated `DIDA_CONFIG_DIR` and Dida-related
credential environment variables cleared.

## Comparison Rules

- Exit codes must match the case manifest.
- Non-JSON output is compared after line-ending normalization.
- JSON stdout is parsed and re-emitted in a canonical compact form before
  comparison.
- Generated local config/cookie paths in JSON output are normalized because old
  and new binaries run with separate isolated config directories.
- Stderr is compared after line-ending normalization.
- A case failure must print the command, side-by-side exit codes, and artifact
  paths for stdout/stderr inspection.

## Running

Build or provide both binaries, then run:

```powershell
pwsh -NoProfile -File scripts/compat-test.ps1 `
  -OldBinary .\bin\dida-go.exe `
  -NewBinary .\target\release\dida.exe
```

For a harness smoke test before the Rust binary exists, point both parameters at
the same built Go binary.

## Expansion Path

1. Add a new local-only command to `tests/compat/cases.json`.
2. Prefer a dry-run or validation shape over an authenticated read/write.
3. If a case might touch config or credentials, document why it remains safe.
4. Once Rust supports a wider command family, add focused cases for aliases and
   important error messages before widening to live integration tests.
