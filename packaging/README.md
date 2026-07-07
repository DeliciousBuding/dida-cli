# Packaging Templates

This directory contains maintainer-facing packaging templates for distribution
channels that usually live in separate repositories or registries.

Current source release: `v0.2.5`

## Channels

| Channel | File | Status |
| --- | --- | --- |
| Homebrew | `homebrew/dida.rb` | Template with macOS and Linux checksums |
| Scoop | `scoop/dida.json` | Template with Windows amd64 and arm64 checksums |
| winget | `winget/README.md` | Submission notes; manifest intentionally not generated yet |

## Update Rules

1. Publish a GitHub Release tag and confirm all archives plus `checksums.txt`
   are attached.
2. Run:

   ```bash
   bash scripts/update-packaging-templates.sh --version v0.2.5
   ```

   Use `--checksums-file <path>` when preparing from a downloaded or staged
   checksum file.
3. Run:

   ```bash
   bash scripts/validate-packaging.sh --version v0.2.5 --checksums-file <path>
   ```

   If you rely on the published GitHub Release checksum asset, omit
   `--checksums-file`.
4. Test the template in the target package manager before publishing it to an
   external tap, bucket, or manifest repository.
5. Do not add credentials, local paths, private test accounts, or release
   automation secrets here.
