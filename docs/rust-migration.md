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

## Remaining Decisions

- Choose whether to keep some high-traffic command parsers manual or move them to `clap` after golden tests cover exact error ordering.
- Decide whether the Rust source-install path is `cargo install --git` or a published crate. Do this before removing `go install` from docs.
- Decide when to replace shell privacy and packaging scripts with Rust tools. Keep the shell scripts until replacements prove byte-for-byte equivalent on fixture repos.
- Decide whether live Web API probes belong in CI as manual workflow jobs or stay as local maintainer checks.
