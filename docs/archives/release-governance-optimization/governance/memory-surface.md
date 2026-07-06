# Memory Surface

Archived on 2026-07-07.

The workflow used native Codex memory for prior DidaCLI release context. No repo-local fallback memory file was created.

Durable facts recorded or relied on during this run:

- The real repository for this run was the local checkout of `DeliciousBuding/dida-cli`.
- `main` is the release branch for `DeliciousBuding/dida-cli`.
- npm publishes must be verified against `https://registry.npmjs.org`; mirrors and cached package pages can lag behind `latest`.
- Release archive attestations are wired in `.github/workflows/release.yml`, but the first archive attestations will only be produced by the next real `vX.Y.Z` tag release after this archive.
- npm Trusted Publishing still requires package-owner setup before removing the `NPM_TOKEN` fallback.
