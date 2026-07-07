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

## Export External Repositories

Homebrew taps and Scoop buckets are published from separate repositories. After
updating and validating the templates, export the repo-ready layouts:

```bash
bash scripts/export-package-manager-repos.sh
```

The default output is ignored by git:

```text
dist/package-manager-repos/homebrew-tap/
dist/package-manager-repos/scoop-bucket/
```

Use explicit repository names when the external repos are ready:

```bash
bash scripts/export-package-manager-repos.sh \
  --homebrew-repo DeliciousBuding/homebrew-dida \
  --scoop-repo DeliciousBuding/scoop-bucket \
  --scoop-bucket dida
```

The export step does not create GitHub repositories, push commits, or publish a
package-manager channel. It only prepares the directories that should become the
roots of those external repositories after native install smoke tests pass.
