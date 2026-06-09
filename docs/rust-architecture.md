# Rust Architecture

This document defines the target Rust shape for DidaCLI. The rewrite must keep the current CLI contract first: command names, aliases, JSON envelopes, config paths, auth behavior, dry-run previews, privacy guards, release archive names, and npm installation must remain compatible with the Go implementation.

The first Rust release should be a drop-in replacement for the Go binary. Users should be able to keep `~/.dida-cli/`, npm installs, shell scripts, agent prompts, and existing automation unchanged.

## Compatibility Contract

The Rust binary keeps these visible behaviors:

- `dida version` prints the version string only.
- `--json` and `-j` are accepted anywhere in the root argument list and are removed before command dispatch.
- `dida --json` returns a JSON validation error for command `dida`.
- `+today` dispatches to `task today` while preserving flags.
- Help and version output stay plain text even when `--json` is present.
- Unknown commands return exit code `1`; with `--json`, errors are written to stdout and stderr stays empty.
- JSON output uses the envelope `{ "ok": bool, "command": string, "meta"?: any, "data"?: any, "error"?: object }`.
- JSON errors use `error.type`, `error.message`, optional `error.hint`, and optional `error.details`.
- JSON serialization uses pretty indentation, a trailing newline, and unescaped HTML characters.
- Validation happens before auth checks. A malformed flag, missing required argument, bad integer, bad date, or flag-like ID must not be hidden by "missing auth".
- Compact output omits large task fields: `content`, `desc`, `items`, `reminders`, and `raw`.
- Destructive commands require `--yes` unless the command is run with `--dry-run`.
- First-class write commands that support `--dry-run` return a local preview and must not mutate remote state.

Golden tests should pin these shapes before command handlers are ported.

## Crate Layout

Use one workspace with a small binary crate and a library crate:

```text
Cargo.toml
crates/
  dida-cli/
    src/main.rs
  dida-core/
    src/
      lib.rs
      app.rs
      cli/
      config/
      auth/
      webapi/
      official/
      openapi/
      model/
      output/
      privacy/
      upgrade/
```

`crates/dida-cli` should only collect process arguments, inject the build version, pass stdout and stderr handles, and exit with the returned code. The testable entrypoint lives in `dida_core::app::run(args, version, stdout, stderr) -> ExitCode`.

`dida-core` owns all behavior. Keep command code inside the library so tests can call handlers without spawning a child process.

## Runtime Dependencies

Use common Rust crates, but keep the dependency list small:

- `clap` for parser primitives only where it does not change current parsing behavior. High-risk commands may keep manual parsing until parity tests cover them.
- `serde` and `serde_json` for envelopes, models, payload passthrough, and schema output.
- `reqwest` with rustls for HTTP clients.
- `tokio` for async HTTP and OAuth callback listeners.
- `time` or `chrono` for date parsing, RFC3339 output, and millisecond timestamps.
- `directories` is optional. If used, wrap it so `DIDA_CONFIG_DIR` and `~/.dida-cli` stay exact.
- `zip`, `tar`, `flate2`, and `sha2` for upgrade archive handling.

Do not introduce a database, background daemon, global cache service, telemetry, or a new config root during the rewrite.

## Command Dispatch

The root dispatcher should mirror `internal/cli/root.go`:

1. Consume every root-level `--json` and `-j`.
2. Return help for no command unless JSON mode is active.
3. Return help for `-h` and `--help`.
4. Return the raw version string for `version` and `--version`.
5. Dispatch by first token.
6. Return an unknown command error through the same output layer.

Each command family should expose `run(args, ctx) -> Result<CommandResult, CliError>`. `ctx` carries:

- version
- JSON mode
- stdout and stderr writers
- config directory provider
- clock
- HTTP clients
- browser opener
- stdin reader for secret input

The command family list should match the Go root commands: `doctor`, `official`, `openapi`, `agent`, `auth`, `sync`, `settings`, `completed`, `closed`, `trash`, `attachment`, `reminder`, `share`, `calendar`, `stats`, `template`, `search`, `user`, `pomo`, `habit`, `quadrant`, `schema`, `channel`, `raw`, `project`, `folder`, `tag`, `filter`, `column`, `comment`, `task`, `upgrade`, and `+today`.

## Output Layer

Keep a single output module:

```rust
pub struct Envelope<T> {
    pub ok: bool,
    pub command: String,
    pub meta: Option<serde_json::Value>,
    pub data: Option<T>,
    pub error: Option<CliErrorBody>,
}

pub struct CliErrorBody {
    pub r#type: Option<String>,
    pub message: String,
    pub hint: Option<String>,
    pub details: Option<serde_json::Value>,
}
```

Command handlers should not print ad hoc JSON. They return typed data or a `CliError`, and the output layer decides stdout versus stderr:

- JSON success: stdout receives the envelope, exit code `0`.
- JSON failure: stdout receives the envelope, stderr is empty, exit code `1`.
- Plain success: command-specific text goes to stdout.
- Plain failure: `dida: <message>` goes to stderr, followed by `hint: <hint>` when present.

The serializer must preserve key names used today, including mixed-case fields such as `dryRun` in Web API dry-run previews and snake_case fields such as `dry_run` in OpenAPI previews where the current command emits them.

## Config and Secret Files

Config paths remain compatible with the Go implementation:

- Default directory: `~/.dida-cli`
- Override: `DIDA_CONFIG_DIR`
- Web API cookie file: `cookie.json`
- Official MCP token file: `official-mcp-token.json`
- OpenAPI OAuth token file: `openapi-oauth.json`
- OpenAPI client credentials file: `openapi-client.json`

File writes must create the config directory with owner-only permissions where the platform supports them. Token files should use `0600` on Unix-like systems. Windows should use normal user-profile ACLs without trying to emulate Unix mode bits.

Existing JSON files must continue to load. Field names stay unchanged:

- `cookie.json`: `token`, `saved_at`
- `official-mcp-token.json`: `token`, `saved_at`
- `openapi-client.json`: `client_id`, `client_secret`, `saved_at`
- `openapi-oauth.json`: `access_token`, `token_type`, `scope`, `expires_in`, `created_at`, `refresh_token`

The Rust config module should expose exact path helpers and tests for env override, default path, and file-name compatibility.

## Auth Channels

The three auth channels stay separate.

Web API uses only the Dida365 browser cookie named `t`. `auth cookie set` keeps the current safety gates:

- `--token-stdin` is the normal path.
- `--token` is rejected unless `DIDA_ALLOW_TOKEN_ARG=1`.
- A full `Cookie:` header is rejected.
- Multiple cookies separated by `;` are rejected.
- `t=<value>` is normalized to the value.
- Tokens with whitespace, control characters, or other `=` signs are rejected.
- Full token values are never printed. Status output may include length and a redacted preview.
- Stdin token input keeps a size limit and fails before storing oversized input.

Official MCP uses `DIDA365_TOKEN` first, then the saved official token file. The token-set command must use stdin by default and avoid printing the token. `official call` remains schema-backed exploration and has no generic dry-run layer.

OpenAPI uses OAuth client credentials and saved OAuth access data. `DIDA365_OPENAPI_CLIENT_ID` and `DIDA365_OPENAPI_CLIENT_SECRET` keep precedence over saved client config. Browser login uses the same default redirect URI, `http://127.0.0.1:17890/callback`, and returns `authorization_url` details when browser launch fails. Manual login keeps `auth-url`, `exchange-code`, and callback-listener paths.

No command should translate one channel credential into another channel credential.

## HTTP Clients

Create three client modules:

- `webapi`: private Web API cookie client, sync payloads, resources, comments, attachments, productivity reads, and raw GET probes.
- `official`: official MCP token client and first-class wrappers.
- `openapi`: OAuth REST client, OAuth helper endpoints, project/task/focus/habit wrappers.

Each client should accept a base URL override in tests. Production defaults stay fixed to the current upstream hosts.

Raw Web API commands remain GET-only. JSON parse failures must include `error.details.statusCode`, `error.details.path`, and a short `error.details.bodySnippet`.

## Dry-Run Model

Dry-run behavior belongs beside each write command, not in a transport interceptor. The preview must be built before auth loading when the current Go command allows local dry-run without credentials, such as schema-backed official and OpenAPI writes.

Preview data should include:

- `dryRun` or `dry_run`, matching the current command family.
- target endpoint or tool name when available.
- request payload exactly as the command would send it.
- a hint that tells the operator how to execute the write.

Delete and merge commands should return `confirmation_required` when `--yes` is missing and `--dry-run` is absent.

## Privacy Gates

Keep the repository privacy guard as a release and CI gate. The Rust rewrite should port or preserve the existing `scripts/check-private-state.sh` behavior until an equivalent Rust or Node checker exists.

The guard must continue to reject:

- `.env` files and local config directories.
- `cookie.json`, `official-mcp-token.json`, `openapi-oauth.json`, and `openapi-client.json`.
- logs, dumps, captures, databases, archives, release outputs, and downloaded npm binaries.
- Dida365 tokens, OAuth access or refresh tokens, OAuth client secrets, cookie headers, bearer tokens, JWTs, and local user paths.

Rust tests should also cover redaction in `auth status`, OpenAPI status, official token status, browser errors, and upgrade errors.

## Schema and Agent Surface

`dida schema list`, `schema show`, `channel list`, and `agent context` are part of the compatibility surface. The Rust implementation should load schema data from typed constants or generated static data, then test that:

- every documented schema command appears in `docs/commands.md`;
- every dry-run-capable schema command includes `--dry-run` in the command string;
- channel output lists `webapi`, `official-mcp`, and `official-openapi`;
- agent outline mode keeps project/filter metadata, task references, and a deduplicated compact `taskIndex`.

Agent docs and the companion skill depend on these commands. Treat schema changes as public API changes.

## Upgrade and Distribution

The Rust binary must keep the current release asset contract:

- Archive names: `dida_vX.Y.Z_<os>_<arch>.zip` for Windows and `dida_vX.Y.Z_<os>_<arch>.tar.gz` for Linux and macOS.
- Archive root directory: `dida_vX.Y.Z_<os>_<arch>/`.
- Binary name: `dida.exe` on Windows and `dida` elsewhere.
- Checksum asset: `checksums.txt` with SHA-256 lines.
- `dida upgrade --check --json` and `dida upgrade --json` keep existing envelope shapes.
- Windows self-upgrade stages a replacement helper and returns `status: "scheduled"`.

The npm package remains a Node wrapper package named `@delicious233/dida-cli`. The Rust rewrite changes the downloaded binary, not the npm package contract:

- `npm/bin/dida` stays the stable Node wrapper.
- Windows downloads to `npm/bin/dida.exe`.
- Unix-like installs download to `npm/bin/dida-bin`.
- `npm pack --dry-run --json` must still include only `bin/dida`, `scripts/install.js`, and `package.json`.
- npm postinstall continues to resolve GitHub Release assets and verify `checksums.txt`.

`go install` disappears after the Rust cutover. Do not remove it from user docs until a Rust source-install replacement exists, such as `cargo install --git` or a published crate, and only after release assets pass install smoke tests.

## CI Shape

During migration, CI should run both stacks where both exist:

- Go: `gofmt`, `go test -count=1 ./...`, `go vet ./...`, and `govulncheck`.
- Rust: `cargo fmt --check`, `cargo clippy --all-targets --all-features -- -D warnings`, `cargo test --workspace`, and `cargo audit` or `cargo deny`.
- Node: npm installer tests and package dry-run validation.
- Shell: privacy guard, packaging template validator, release archive verifier, and their tests.
- Build matrix: Windows amd64 and arm64, Linux amd64 and arm64, macOS amd64 and arm64.

The final Rust-only CI can remove Go gates after parity tests, release smoke, and npm smoke pass on the Rust binary.

## Test Strategy

Port tests by contract, not file order. Start with golden fixtures for:

- root dispatch and `--json` handling;
- JSON success and failure envelopes;
- validation-before-auth cases;
- config path compatibility;
- cookie normalization and token redaction;
- dry-run previews for Web API, official MCP, and OpenAPI writes;
- `--yes` enforcement;
- compact task output;
- schema and command-reference coverage;
- release asset selection and checksum verification;
- npm install script expectations.

Live Web API and OpenAPI tests should stay opt-in. Unit and integration tests must use local HTTP test servers and temporary config directories.

## Cutover Rule

The Rust binary can replace the Go binary when these checks are true:

1. Command reference examples either pass or fail with the same documented auth requirement.
2. Golden JSON fixtures match for representative success, validation, auth, dry-run, confirmation, and transport errors.
3. Existing config files load without migration.
4. npm installation downloads the Rust release binary and `dida version` returns the tag.
5. `dida doctor --json`, `dida schema list --compact --json`, and `dida agent context --outline --json` keep agent-facing shapes.
6. Privacy guard and release archive verifier pass without weakening patterns.
7. The release workflow produces the same asset names and checksum file.
