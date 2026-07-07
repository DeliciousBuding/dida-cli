# Distribution

DidaCLI should be installable by humans, CI jobs, and LLM/Agent runtimes without
cloning the repository. GitHub Releases are the canonical binary source.

## Current Channels

| Channel | Status | Notes |
| --- | --- | --- |
| GitHub Releases | Primary | Tag pushes build checksum-verified binary archives for Windows, Linux, and macOS, then generates GitHub artifact attestations for the archives. |
| `install.sh` | Primary | Unix-like one-line installer. Supports `DIDA_VERSION`, `DIDA_INSTALL_DIR`, and `DIDA_REPO`; latest resolution uses release checksum assets instead of the GitHub API. |
| `install.ps1` | Primary | Windows PowerShell installer. Supports the same environment variables; latest resolution uses release checksum assets instead of the GitHub API. |
| npm installer package | Published | `@delicious233/dida-cli` downloads the matching GitHub Release binary during postinstall and verifies release checksums. |
| `go install` | Developer fallback | Works when Go is installed, but does not use release archives. |

## Planned Channels

| Channel | Priority | Plan |
| --- | ---: | --- |
| Homebrew tap | 4 | Template is generated from release checksums under `packaging/homebrew/dida.rb`; publish from a dedicated tap after native smoke. |
| Scoop bucket | 4 | Manifest is generated from release checksums under `packaging/scoop/dida.json`; publish from a dedicated bucket after install smoke. |
| winget | 5 | Notes exist in `packaging/winget/README.md`; generate a manifest after release cadence is stable. |

## GitHub Releases

Release workflow:

- Trigger: tag push matching `v*`.
- Platforms:
  - `windows/amd64`, `windows/arm64`
  - `linux/amd64`, `linux/arm64`
  - `darwin/amd64`, `darwin/arm64`
- Archives:
  - Windows: `.zip`
  - Linux/macOS: `.tar.gz`
- Verification: `checksums.txt` with SHA-256 hashes and GitHub artifact attestations.
- Package-manager handoff: after the GitHub Release exists, the workflow
  uploads `dida-package-manager-repos-vX.Y.Z` with repo-ready Homebrew tap and
  Scoop bucket directories.

Verify a downloaded archive after release:

```bash
gh attestation verify dida_vX.Y.Z_linux_amd64.tar.gz --repo DeliciousBuding/dida-cli
```

Create a release:

```bash
make release-check VERSION=vX.Y.Z
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

See [`../RELEASE.md`](../RELEASE.md) for the full maintainer checklist.

## Install Scripts

Unix-like systems:

```bash
curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.ps1 -UseB | iex
```

Pin a version:

```bash
DIDA_VERSION=vX.Y.Z curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Custom repo or install directory:

```bash
DIDA_REPO=DeliciousBuding/dida-cli DIDA_INSTALL_DIR="$HOME/bin" sh install.sh
```

## npm Installer

The `npm/` directory contains the published npm wrapper package:

- package name: `@delicious233/dida-cli`
- package version is tracked in `npm/package.json`
- postinstall downloads the GitHub Release archive matching the package version
- checksum verification uses the release `checksums.txt`
- `bin/dida` is the stable Node wrapper
- Windows stores the downloaded binary as `bin/dida.exe`
- Unix-like systems store the downloaded binary as `bin/dida-bin` so the
  wrapper is not overwritten
- Windows and Linux smoke coverage installs the release binary and runs
  `dida version` plus `dida doctor --json`
- Linux smoke coverage verifies the Unix wrapper/binary split where
  `bin/dida` remains a Node wrapper and the downloaded binary is stored as
  ignored `bin/dida-bin`
- `npm pack --dry-run --json` verifies that the package contains only
  `bin/dida`, `scripts/install.js`, and `package.json`
- release workflow npm preflight checks the npm token, duplicate package
  version, package contents, and provenance-ready publish path before creating
  the GitHub Release
- release metadata and release notes are validated by reusable scripts under
  `scripts/` and by CI before tag publishing

Before each publish:

1. Confirm the GitHub tag, npm package version, and binary version match.
2. Run `npm pack --dry-run --json` from `npm/`.
3. Run a native macOS npm installer smoke when macOS packaging changes.

## `go install`

For developers with Go installed:

```bash
go install github.com/DeliciousBuding/dida-cli/cmd/dida@latest
```

Use Release assets for normal users and agents. `go install` depends on a Go
toolchain and may not match packaged release behavior.

## Package Manager Templates

`packaging/` contains maintainer templates for package managers that generally
live outside this repository:

- `packaging/homebrew/dida.rb` pins macOS and Linux release archives
  with checksums. The formula installs the binary from the release archive's
  top-level platform directory.
- `packaging/scoop/dida.json` pins Windows amd64 and arm64 release
  archives with checksums.
- `packaging/winget/README.md` records the maintainer handoff for winget
  manifest generation.

Static validation can compare Homebrew and Scoop template URLs plus SHA-256
hashes against a release `checksums.txt` after the release archives exist.
Release archive listing checks confirm Homebrew must install from the nested
`dida_v.../dida` path and Scoop's `extract_dir` matches the Windows zip's
top-level `dida_v..._windows_<arch>/` directory. Native package-manager smoke
tests still need hosts with `brew`, Scoop, and `wingetcreate` installed.

Regenerate package-manager templates after a release:

```bash
bash scripts/update-packaging-templates.sh --version vX.Y.Z
bash scripts/validate-packaging.sh --version vX.Y.Z
```

Use `--checksums-file <path>` with both commands when using a local
`checksums.txt` copy.

Export repository-ready Homebrew tap and Scoop bucket layouts:

```bash
bash scripts/export-package-manager-repos.sh
```

The default export is ignored by git and writes:

- `dist/package-manager-repos/homebrew-tap/Formula/dida.rb`
- `dist/package-manager-repos/homebrew-tap/README.md`
- `dist/package-manager-repos/scoop-bucket/bucket/dida.json`
- `dist/package-manager-repos/scoop-bucket/README.md`

The recommended external repositories are:

- `DeliciousBuding/homebrew-dida`, installed as `brew tap DeliciousBuding/dida`
- `DeliciousBuding/scoop-bucket`, installed as
  `scoop bucket add dida https://github.com/DeliciousBuding/scoop-bucket`

Exporting does not create those repositories, push commits, or publish the
package-manager channels. Treat it as a local staging step before native
Homebrew and Scoop install smoke tests.

Run the package-manager preflight before using an export:

```bash
bash scripts/package-manager-smoke-preflight.sh
```

The default preflight exports the Homebrew tap and Scoop bucket layouts, checks
that the expected files are present, rejects local paths or secret-like text,
and prints the native smoke commands. On a host with the package manager
installed, add the matching flag:

```bash
bash scripts/package-manager-smoke-preflight.sh --run-homebrew-smoke
bash scripts/package-manager-smoke-preflight.sh --run-scoop-smoke
```

After a tag release, the same export is available as the
`dida-package-manager-repos-vX.Y.Z` workflow artifact. Use that artifact when
publishing the external tap or bucket, so the submitted files are generated from
the release checksum asset rather than from a hand-edited local copy.

Homebrew, Scoop, and winget files are maintainer templates for external
package-manager submissions.

## winget Submission

winget is still a deferred channel. The repository keeps submission notes under
`packaging/winget/README.md`, not generated manifests.

Run the winget preflight before using `wingetcreate`:

```bash
bash scripts/winget-submission-preflight.sh
```

The preflight checks the package id, current Windows release URLs, packaging
metadata, and submission boundary, then prints the `wingetcreate new` and
`winget validate --manifest` commands. If a manifest directory already exists
on a Windows packaging host, validate it without committing generated files:

```bash
bash scripts/winget-submission-preflight.sh --manifest-dir <manifest-directory> --run-validate
```
