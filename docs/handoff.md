# DidaCLI Handoff

Last updated: 2026-06-09 21:40 HKT

## Current State

Main repo:

- Path: `D:\Code\Projects\tools\DidaCLI`
- Branch: `main`
- Current pushed commit: `4f061b6 fix release notes shell lint`
- Latest main CI: success, GitHub Actions run `27209994229`
- Existing tag: `v0.2.3` at `93a6db5`
- `v0.2.4` tag has not been created or pushed.
- npm registry currently has `@delicious233/dida-cli` versions `0.2.0` and `0.2.1` only.

Rust rewrite worktree:

- Path: `C:\Users\Ding\.config\superpowers\worktrees\DidaCLI\codex-rust-rewrite`
- Branch: `codex/rust-rewrite`
- Current pushed commit before this handoff document: `e05adea improve rust cli offline parity`
- This handoff update adds verified schema parity work in `crates/dida-cli/src/lib.rs` and `crates/dida-cli/src/schema.rs`.
- After pushing this handoff commit, check `git log --oneline -3` for the exact latest branch commit.

## What Is Done

Stable Go baseline:

- Release workflow shell lint is fixed on `main`.
- Packaging metadata is aligned to `v0.2.4`.
- `npm/package.json` is `0.2.4`.
- `CHANGELOG.md` has `v0.2.4`.
- Packaging templates point at `v0.2.4`.
- Main CI is green after `4f061b6`.

Rust rewrite:

- Workspace scaffold exists with `dida-cli`, `dida-core`, and `dida-http`.
- Architecture and migration docs exist:
  - `docs/rust-architecture.md`
  - `docs/rust-migration.md`
- Compatibility harness exists:
  - `scripts/compat-test.ps1`
  - `tests/compat/cases.json`
  - `tests/compat/PLAN.md`
- `dida-http` has retry, timeout, bounded download, checksum, and mock transport tests.
- Rust CLI now matches Go for the offline/local compatibility matrix.

## Verified Commands

Main baseline:

```powershell
cd D:\Code\Projects\tools\DidaCLI
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
go test -count=1 ./...
cd npm
npm test
```

Release helper checks that passed during stabilization:

```powershell
cd D:\Code\Projects\tools\DidaCLI
bash scripts/check-private-state.sh
bash scripts/check-private-state.test.sh
bash scripts/validate-packaging.test.sh
bash scripts/verify-release-archives.test.sh
bash scripts/validate-packaging.sh --metadata-only --version v0.2.4
```

Rust rewrite:

```powershell
cd C:\Users\Ding\.config\superpowers\worktrees\DidaCLI\codex-rust-rewrite
cargo fmt
cargo test
cargo clippy --all-targets -- -D warnings
go build -o .\bin\dida-go.exe .\cmd\dida
cargo build -p dida-cli
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts\compat-test.ps1 -OldBinary .\bin\dida-go.exe -NewBinary .\target\debug\dida.exe -KeepArtifacts
```

Latest verified result:

- `cargo test`: passed, 18 tests.
- `cargo clippy --all-targets -- -D warnings`: passed.
- Compatibility harness: passed, `35/35`.
- Main CI: passed, run `27209994229`.

## Release Status

Do not push `v0.2.4` until npm authentication is fixed.

Reason:

- The release workflow now fails fast when `NPM_TOKEN` is missing or invalid.
- Earlier release logs showed `npm whoami` returning `E401 Unauthorized`.
- Generated release notes advertise npm install. A GitHub release without npm publish would create a broken install path.

Release path after token fix:

```powershell
cd D:\Code\Projects\tools\DidaCLI
git fetch origin
git checkout main
git pull --ff-only
git tag v0.2.4
git push origin v0.2.4
```

Then watch:

```powershell
gh run list --workflow release.yml --limit 10
gh run view <run-id> --log-failed
npm view @delicious233/dida-cli versions --json --registry=https://registry.npmjs.org
```

## Dirty Tree Notes

Main worktree shows many modified Go files under `internal/cli` and `internal/webapi`.

Observed behavior:

- The file blob hashes matched the index earlier.
- `git diff` showed no content changes for representative Go files.
- The status appears to be Windows line-ending noise.

Do not stage or commit those Go files unless a new content diff proves a real change. Use explicit path staging.

Rust schema parity was implemented in:

- `crates/dida-cli/src/lib.rs`
- `crates/dida-cli/src/schema.rs`

If these files are dirty in a later session, run `git diff` before editing. They are the source of truth for the current `35/35` offline compatibility gate.

## Privacy Rules

Keep these out of the repo:

- `.env`
- Dida cookie files
- OAuth tokens
- local config directories
- browser dumps
- packet captures
- local database files
- temp compatibility artifacts
- generated `target/` and `bin/` outputs

Run the private-state guard before release or merge work:

```powershell
bash scripts/check-private-state.sh
```

## Next Work

Immediate:

1. Confirm `codex/rust-rewrite` remains `35/35` on the compatibility harness after pull.
2. Rotate or fix the GitHub Actions `NPM_TOKEN`.
3. Push `v0.2.4` only after npm auth is valid.
4. Continue Rust porting behind the compatibility gate.

Rust rewrite:

1. Move the current compatibility-only CLI code toward the target architecture in `docs/rust-architecture.md`.
2. Split CLI behavior out of `dida-cli` into `dida-core` without breaking the `35/35` compatibility gate.
3. Port config and auth file handling next.
4. Port dry-run write payload builders into typed Rust structs.
5. Port Web API reads through `dida-http` using fake transports first.
6. Add Rust CI jobs after the branch stays green locally.
7. Keep Go release builds active until Rust reaches full command parity.

Merge/release readiness requires stronger evidence than the current branch has:

- all Go tests still pass;
- all Rust tests and clippy pass;
- Go-vs-Rust compatibility covers command families beyond the current offline set;
- release archives install and smoke-test the Rust binary on supported platforms;
- npm postinstall accepts the final binary names;
- private-state guard passes;
- GitHub Actions are green;
- npm publish succeeds.
