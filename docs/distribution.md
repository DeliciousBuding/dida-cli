# Distribution

DidaCLI should be installable by humans, CI jobs, and LLM/Agent runtimes without
cloning the repository. GitHub Releases are the canonical binary source.

## Current Channels

| Channel | Status | Notes |
| --- | --- | --- |
| GitHub Releases | Primary | Tag pushes build checksum-verified binary archives for Windows, Linux, and macOS. |
| `install.sh` | Primary | Unix-like one-line installer. Supports `DIDA_VERSION`, `DIDA_INSTALL_DIR`, and `DIDA_REPO`; latest resolution uses release checksum assets instead of the GitHub API. |
| `install.ps1` | Primary | Windows PowerShell installer. Supports the same environment variables; latest resolution uses release checksum assets instead of the GitHub API. |
| npm installer package | Smoke-tested skeleton | `npm/` contains a placeholder package that downloads GitHub Release binaries. Not published yet. |
| `go install` | Developer fallback | Works when Go is installed, but does not use release archives. |

## Planned Channels

| Channel | Priority | Plan |
| --- | ---: | --- |
| Homebrew tap | 4 | Template exists in `packaging/homebrew/dida.rb`; publish from a dedicated tap after native smoke. |
| Scoop bucket | 4 | Template exists in `packaging/scoop/dida.json`; publish from a dedicated bucket after install smoke. |
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
- Verification: `checksums.txt` with SHA-256 hashes.

Create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

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
DIDA_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/DeliciousBuding/dida-cli/main/install.sh | sh
```

Custom repo or install directory:

```bash
DIDA_REPO=DeliciousBuding/dida-cli DIDA_INSTALL_DIR="$HOME/bin" sh install.sh
```

## npm Installer Skeleton

The `npm/` directory is a package skeleton for a future npm distribution:

- package name placeholder: `@vectorcontrol/dida-cli`
- postinstall downloads the matching GitHub Release archive
- `bin/dida` is the stable Node wrapper
- Windows stores the downloaded binary as `bin/dida.exe`
- Unix-like systems store the downloaded binary as `bin/dida-bin` so the
  wrapper is not overwritten
- local Windows and WSL Linux smoke tests installed `v0.1.4` from GitHub
  Releases in temporary copies and verified `node bin/dida version`
- Windows `install.ps1` latest smoke installed `v0.1.8` from GitHub Releases
  and verified `dida version` plus the installer's `dida doctor --json` check
- WSL Linux `install.sh` latest smoke installed `v0.1.8` from GitHub Releases
  and verified `dida version` plus the installer's `dida doctor --json` check

Do not publish it until:

1. Package ownership and final npm scope are confirmed.
2. macOS npm installer smoke is repeated on a native macOS host.
3. Publishing automation and provenance policy are defined.

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

- `packaging/homebrew/dida.rb` pins `v0.1.8` macOS and Linux release archives
  with checksums.
- `packaging/scoop/dida.json` pins `v0.1.8` Windows amd64 and arm64 release
  archives with checksums.
- `packaging/winget/README.md` records the future winget submission boundary
  without committing generated manifests prematurely.

These files are not published channels yet. Treat them as release engineering
inputs for a future tap, bucket, or winget submission.
