# Rust Migration Plan

This plan migrates DidaCLI from Go to Rust without changing the public CLI contract. It is staged so each phase can ship behind parity tests before the Go implementation is removed.

## Non-Negotiable Compatibility

The Rust rewrite preserves:

- command names, aliases, flags, help behavior, and exit codes;
- JSON envelopes and error shapes;
- stdout and stderr routing in JSON and plain modes;
- `~/.dida-cli` and `DIDA_CONFIG_DIR`;
- Web API cookie auth, official MCP token auth, and OpenAPI OAuth separation;
- token stdin defaults, token redaction, and `DIDA_ALLOW_TOKEN_ARG` behavior;
- dry-run previews and `--yes` gates;
- compact task output;
- privacy guard coverage;
- GitHub Release archive names and checksums;
- npm wrapper package behavior.

Any intentional difference needs a migration note, a test, and a README or command-doc update before release.

## Migration Architecture

The migration follows the crate boundaries in `docs/rust-architecture.md`:

- `dida-cli` lands first and stays process-only.
- `dida-core` ports command behavior in small slices: root parser, output, config, auth, local commands, Web API reads, Web API writes, official MCP, OpenAPI, upgrade.
- `dida-http` ports reusable network mechanics separately from command handlers: transport, bounded response reads, retry policy, timeout policy, channel clients, downloads, checksums.

Porting order within each command family:

1. Add the command ID to the manifest or schema surface.
2. Add parser tests for flags, aliases, positionals, validation order, and JSON mode.
3. Add output fixtures for success, validation failure, auth failure, dry-run, confirmation, and transport failure where relevant.
4. Implement local parsing and dry-run behavior in `dida-core`.
5. Add auth provider and config/state tests.
6. Wire the HTTP call through `dida-http` with fake-transport tests.
7. Run representative Go/Rust parity fixtures.
8. Mark the command status in the migration tracker.

Do not port commands by copying Go file order. Port by user-facing contract and failure risk.

## Command Status Tracker

Track every command with these fields:

| Field | Meaning |
| --- | --- |
| `id` | Stable schema or command ID, such as `task.create` |
| `family` | Root family, such as `task`, `openapi`, or `official` |
| `channel` | `local`, `webapi`, `official-mcp`, `official-openapi`, or `upgrade` |
| `stage` | `not-started`, `skeleton`, `local`, `network`, `parity`, or `released` |
| `auth` | Required auth channel or `none` |
| `dry_run` | Whether dry-run parity is required |
| `confirm` | Whether `--yes` parity is required |
| `fixtures` | Golden fixture files or test names |
| `notes` | Known compatibility gaps |

The tracker can start as a checked-in Markdown table or JSON file. Release gates should read it once the command list is stable.

Stage rules:

- `skeleton` commands parse enough to return a stable `not_implemented` or `not_ported` error.
- `local` commands pass parse, validation, output, config, dry-run, and confirmation tests.
- `network` commands pass fake-transport tests and never use live network in normal CI.
- `parity` commands match Go fixtures for representative outcomes.
- `released` commands are included in archive and npm smoke tests.

## Phase 0: Contract Inventory

Freeze the current Go behavior before writing Rust command handlers.

Deliverables:

- A fixture set for root dispatch, output envelopes, auth errors, validation errors, dry-run previews, confirmation errors, compact task output, and upgrade metadata.
- A command manifest generated from `didaCommandSchemas()` or an equivalent static export.
- A list of config files and JSON field names used by Web API, official MCP, and OpenAPI auth.
- A release contract note covering archive names, nested paths, binary names, `checksums.txt`, npm postinstall paths, and Windows staged upgrade status.

Required checks:

```bash
go test -count=1 ./...
bash scripts/check-private-state.sh
bash scripts/validate-packaging.sh --metadata-only
bash scripts/verify-release-archives.test.sh
npm test --prefix npm
```

## Phase 1: Rust Skeleton

Add the Cargo workspace and a Rust binary that can run root-level commands with no network behavior.

Deliverables:

- `crates/dida-cli` process entrypoint.
- `crates/dida-core` library entrypoint with injectable stdout, stderr, stdin, clock, browser opener, and HTTP clients.
- Output module with the current envelope and error body.
- Config module with `DIDA_CONFIG_DIR`, default `~/.dida-cli`, and path helpers.
- Root dispatcher for `version`, `--version`, help, missing command, unknown command, `--json`, and `+today`.

Parity tests:

- `dida version` prints only the version.
- `dida --json` returns a validation envelope on stdout.
- JSON errors leave stderr empty.
- Help stays plain text.
- Unknown command behavior matches Go.
- `+today --limit 2 --json` dispatches as `task today` or reaches the same auth envelope once task reads are ported.

CI additions:

```bash
cargo fmt --check
cargo clippy --all-targets --all-features -- -D warnings
cargo test --workspace
```

Exit criteria:

- `dida-cli` has no command logic beyond process wiring.
- `dida-core` exposes one testable app entrypoint.
- Root JSON flag stripping matches Go for repeated `--json` and `-j`.
- No network client is initialized for version, help, missing command, or unknown command.
- Unknown command JSON failures write stdout only and return exit code `1`.

## Phase 2: Config and Auth

Port local credential handling before network commands.

Deliverables:

- Cookie normalization, save, load, clear, status, and redaction.
- Browser login profile cleanup behavior matching `auth logout`.
- Official MCP token save, load, clear, status, and env precedence.
- OpenAPI client credential save/load/status with env precedence.
- OpenAPI OAuth token save/load/logout/status using existing file names and field names.
- Stdin secret readers with size limits.

Parity tests:

- `DIDA_CONFIG_DIR` overrides the config root.
- Missing home fallback matches the Go behavior where practical.
- Existing Go-written JSON files load in Rust.
- `auth cookie set --token secret --json` is rejected unless `DIDA_ALLOW_TOKEN_ARG=1`.
- `auth cookie set --token-stdin --json` stores the normalized value and does not print it.
- `Cookie: t=...`, multi-cookie input, whitespace, control characters, and extra `=` are rejected.
- `auth status --verify --json` returns an auth error before Web API reads are available.
- Official and OpenAPI status commands never print full secrets.

Exit criteria:

- All existing Go-written credential files load without migration.
- Auth file writes are atomic at the file level.
- Secret status output is produced by auth modules as redacted summaries.
- Env precedence is tested per channel and never crosses channels.
- Failed token normalization leaves existing files unchanged.

## Phase 3: Local Discovery Commands

Port commands that do not need remote calls.

Deliverables:

- `doctor` without `--verify`.
- `schema list`, `schema show`, and `channel list`.
- `openapi doctor`, `openapi auth-url`, and local OpenAPI client status.
- `agent context` argument parsing with a temporary "remote reads not ported" error until sync lands.

Parity tests:

- Schema compact output exposes command IDs, command strings, status, auth requirements, dry-run support, confirmation requirements, and compact-output support.
- Every dry-run schema command contains `--dry-run`.
- `docs/commands.md` mentions each schema command prefix.
- `channel list --json` lists exactly the three auth channels.
- `openapi doctor --json` includes `default_redirect_uri`, `default_scope`, and `next`.

Exit criteria:

- Local commands work without any credential files.
- Schema and channel outputs are stable fixtures.
- `agent context` returns a stable not-ported remote-read error until Web API reads are available.
- Docs coverage checks fail when a schema command is missing from `docs/commands.md`.

## Phase 4: Web API Reads

Port the private Web API client and read-only command families.

Deliverables:

- Cookie-auth HTTP client.
- Sync all and checkpoint.
- Project, folder, tag, filter, column, task, completed, closed, trash, settings, attachment quota, reminder, share, calendar, stats, template, search, user, pomo, habit, quadrant, and comment reads.
- Raw GET probes with v1/v2 API version selection.
- Model normalization for tasks, dates, tags, columns, and compact output.

Parity tests:

- Validation-before-auth for flag-like IDs and bad flags.
- Missing cookie errors match current `auth` type and hints.
- Compact task output omits large fields.
- Raw JSON failures include status code, path, and body snippet.
- Date range parsing matches current day-boundary behavior.

Live checks stay opt-in and require an operator-controlled account.

Exit criteria:

- Web API client tests use fake transport or local servers only.
- Missing cookie errors keep current type, message, hint, and output stream behavior.
- Sync checkpoint writes occur after successful response normalization.
- Raw GET probes reject non-GET methods by construction.
- Compact output fixtures cover representative task shapes with large fields present upstream.

## Phase 5: Web API Writes

Port first-class Web API writes after read parity is stable.

Deliverables:

- Task create, update, complete, delete, move, and parent.
- Project create, update, delete.
- Folder create, update, delete.
- Tag create, update, rename, merge, delete.
- Column create.
- Comment create, update, delete, including attachment upload.

Parity tests:

- Every write parses and validates flags before loading auth.
- `--dry-run` returns the same preview shape and does not require network access when the Go command does not.
- Delete and merge commands return `confirmation_required` without `--yes`.
- Comment attachment dry-run reports quota and upload intent without reading or uploading more than needed.
- Multipart upload uses field name `file`.

Exit criteria:

- Every write has a local dry-run fixture or a documented reason it cannot support dry-run.
- Every destructive command has a missing-`--yes` fixture.
- Validation-before-auth tests cover malformed IDs, dates, integers, and enum values.
- Network mutation tests assert method, path, headers, and payload through fake transport.
- Failed mutation responses never update local state.

## Phase 6: Official MCP

Port official token commands, schema discovery, generic calls, and first-class wrappers.

Deliverables:

- Token status, set, and clear.
- Official doctor, tools, show, and call.
- Official project, task, habit, and focus wrappers.
- Local dry-run previews for first-class official writes.

Parity tests:

- `DIDA365_TOKEN` takes precedence over saved token config.
- Usage errors occur before missing-token errors.
- `official call` remains explicit and has no generic dry-run layer.
- First-class task, habit, and focus writes support local `--dry-run`.
- Floating-point habit values reject invalid values such as `NaN` and infinities.

Exit criteria:

- Official MCP token loading uses `DIDA365_TOKEN` before saved token files.
- Generic `official call` stays schema-backed and does not gain implicit dry-run.
- MCP RPC errors map into stable CLI error bodies.
- First-class wrappers can be tested without network through fake transport.

## Phase 7: Official OpenAPI

Port OAuth and official REST wrappers.

Deliverables:

- Browser OAuth login with callback listener.
- Manual OAuth URL, callback listener, and code exchange.
- Project, task, focus, and habit commands.
- OpenAPI dry-run previews and `--yes` enforcement.

Parity tests:

- Browser launch failure returns `error.type: "browser"` and `details.authorization_url`.
- `--no-open` prints manual URL guidance in plain mode.
- OAuth client env vars override saved client config.
- Usage errors happen before missing-token errors.
- OpenAPI dry-run writes do not require a saved token when current Go behavior allows it.
- Focus `--type` values and habit date formats match current parsing.

Exit criteria:

- OAuth callback listener tests cover success, error, timeout, and port conflict.
- Browser launch failures keep manual authorization URL details in JSON.
- OAuth refresh failures do not remove saved client credentials.
- OpenAPI request wrappers share timeout, retry, and response-size policy with the other clients.

## Phase 8: Upgrade and Packaging

Port self-upgrade and switch release builds to Rust while keeping package contracts.

Deliverables:

- GitHub latest release lookup.
- Platform asset selection.
- Concurrent archive and checksum downloads where practical.
- SHA-256 verification.
- Zip and tar.gz extraction from the nested archive root.
- Unix in-place replacement.
- Windows staged replacement with `status: "scheduled"`.
- CI build matrix for Windows, Linux, and macOS on amd64 and arm64.
- npm smoke tests against Rust release assets.

Parity tests:

- Asset names match `dida_vX.Y.Z_<os>_<arch>`.
- Missing asset and missing checksum errors match current JSON error categories.
- Checksum mismatch fails before extraction.
- Windows replacement script handles paths with `%`.
- npm install downloads to `bin/dida.exe` on Windows and `bin/dida-bin` on Unix-like systems.
- `npm pack --dry-run --json` contains only the wrapper, installer script, and package metadata.

Exit criteria:

- Archive verification passes for every release target.
- Checksum mismatch fails before extraction.
- Extraction validates the expected archive root directory.
- Windows staged replacement reports `status: "scheduled"` and leaves the running binary intact.
- npm smoke tests install from Rust release assets on Windows and Linux.

## Phase 9: Cutover

Make Rust the primary implementation after parity and release smoke pass.

Steps:

1. Build Rust release archives with the existing names.
2. Run release archive verifier against Rust archives.
3. Run Linux and Windows npm install smoke tests.
4. Run `dida version`, `dida doctor --json`, `dida schema list --compact --json`, and `dida upgrade --check --json` from installed release artifacts.
5. Update README language from Go binary to native binary.
6. Replace `go install` guidance with a Rust source-install path only after it is tested.
7. Remove Go CI gates after the Go command tree is deleted.
8. Keep privacy, packaging, npm, and archive-verifier gates.

Exit criteria:

- The stable release contains Rust binaries under the existing asset names.
- npm stable install resolves those assets and runs `dida version`.
- Go fallback assets remain reachable for one release cycle.
- README, command docs, and install docs no longer describe Go as the primary implementation.
- The command status tracker has no `network` commands left below `parity` for released surfaces.

## Regression Matrix

Run this matrix before the first Rust release candidate:

| Area | Commands |
| --- | --- |
| Root | `dida version`, `dida --json`, `dida --help`, `dida nope --json` |
| Auth | `dida auth login --json`, `dida auth cookie set --token-stdin --json`, `dida auth status --verify --json`, `dida auth logout --json` |
| Schema | `dida schema list --compact --json`, `dida schema show task.create --json`, `dida channel list --json` |
| Agent | `dida agent context --outline --json`, `dida agent context --full --json` |
| Web API reads | `dida sync all --json`, `dida task today --compact --json`, `dida project list --json`, `dida raw get /batch/check/0 --json` |
| Web API writes | `dida task create --project <id> --title T --dry-run --json`, `dida task delete <id> --project <id> --json`, `dida tag merge old new --dry-run --json` |
| Official MCP | `dida official token status --json`, `dida official show list_projects --json`, `dida official task batch-add --args-json '{"tasks":[{"title":"T"}]}' --dry-run --json` |
| OpenAPI | `dida openapi doctor --json`, `dida openapi auth-url --json`, `dida openapi project create --args-json '{"name":"P","viewMode":"list","kind":"TASK"}' --dry-run --json` |
| Upgrade | `dida upgrade --check --json`, `dida upgrade --json` from a writable test install |
| Distribution | install scripts, npm postinstall, release archive verifier, packaging template validator |

## Rollout Controls

Rollout moves through named channels:

| Channel | Published Asset | npm Behavior | Intended Use |
| --- | --- | --- | --- |
| `dev-only` | none | unchanged | CI and local implementation |
| `shadow` | CI artifacts only | unchanged | Go/Rust fixture comparison |
| `preview` | prerelease Rust archives | stable npm still uses Go release | manual operator testing |
| `staging` | staging Rust release | npm smoke points at staging release | install and upgrade validation |
| `stable` | normal Rust release archives | npm downloads Rust assets | user cutover |

Promotion requirements:

- `dev-only` to `shadow`: root, output, config, and auth tests pass.
- `shadow` to `preview`: local discovery commands and representative Web API reads reach parity.
- `preview` to `staging`: Web API writes, official MCP, OpenAPI, and upgrade flows pass fake-transport tests.
- `staging` to `stable`: regression matrix passes on Windows, Linux, and macOS release artifacts.

Rollback requirements:

- Keep the last Go-backed stable release documented until the first Rust stable release completes one release cycle.
- Do not change npm package names, wrapper paths, or postinstall environment variables during Rust rollout.
- If a stable Rust release is pulled, publish a patch release that points npm back to the last known-good asset set.

## Performance Targets

Performance is a release gate for local command paths:

| Command | Target |
| --- | --- |
| `dida version` | under 20 ms process time |
| `dida --help` | under 50 ms process time |
| `dida schema list --compact --json` | under 100 ms process time |
| `dida auth status --json` | under 100 ms process time without verification |
| `dida upgrade --check --json` | no slower than the Go implementation by more than 25% on the same network |

Implementation rules:

- Local commands should not create HTTP clients.
- Pure local commands should not start a Tokio runtime.
- Tests should use fake clocks and fake transports instead of sleeps.
- Upgrade downloads may use parallel archive and checksum fetches, but final output order stays deterministic.
- Large response handling uses bounded reads and streaming downloads.

Record performance in release-candidate smoke logs. A local-command regression above 2x the Go baseline blocks stable promotion.

## Failure Isolation Checks

Add tests or release smoke cases for these failure boundaries:

- Parser and validation failures do not read credential files.
- Missing Web API cookie does not affect official MCP or OpenAPI commands.
- Missing official MCP token does not affect Web API commands.
- Missing OpenAPI OAuth token does not affect saved OpenAPI client credentials.
- Browser OAuth failure returns manual login details and leaves existing OAuth files intact.
- Web API raw GET failures do not update sync checkpoints.
- Write-command upstream failures do not update local state.
- Upgrade checksum failure leaves the current binary and staged files untouched.
- Archive-root mismatch fails before replacement.
- Redaction runs on all error paths that include headers, URLs, request bodies, or filesystem paths.

The first error visible to the operator should be the one they can act on. Cleanup failures can be recorded in details, but they must not hide the command failure.

## Remaining Decisions

- Choose whether to keep some high-traffic command parsers manual or move them to `clap` after golden tests cover exact error ordering.
- Decide whether the Rust source-install path is `cargo install --git` or a published crate. Do this before removing `go install` from docs.
- Decide when to replace shell privacy and packaging scripts with Rust tools. Keep the shell scripts until replacements prove byte-for-byte equivalent on fixture repos.
- Decide whether live Web API probes belong in CI as manual workflow jobs or stay as local maintainer checks.
