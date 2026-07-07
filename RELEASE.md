# Release Process

DidaCLI releases are tag-driven. GitHub Releases are the canonical binary source, and the npm package mirrors the same semantic version.

## Version Rules

- Use tags in the form `vX.Y.Z`.
- `npm/package.json` must contain the same version without the leading `v`.
- `CHANGELOG.md` must contain a `## [vX.Y.Z] - YYYY-MM-DD` section before a normal release.
- The release tag must point to a commit reachable from `origin/main`.
- Prefer annotated tags:

```bash
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

## Preflight

Run these checks before pushing a release tag:

```bash
make release-check VERSION=vX.Y.Z
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
```

`make release-check` validates release metadata, changelog structure, npm package contents, pinned GitHub Actions, repository governance files, current package-manager template metadata, shell release helpers, and workflow syntax. It does not publish anything.

## Automation

Pushing a `vX.Y.Z` tag runs `.github/workflows/release.yml`:

1. validate tag, changelog, npm version, formatting, tests, vet, vulnerability scan, and packaging metadata
2. build Windows, Linux, and macOS archives for amd64 and arm64
3. verify archive shape and binary version
4. generate SHA-256 checksums and GitHub artifact attestations for the release archives
5. create or update the GitHub Release with `checksums.txt`
6. smoke-test npm install on Linux and Windows
7. export Homebrew tap and Scoop bucket repository layouts as a workflow artifact
8. publish `@delicious233/dida-cli` to npm with provenance, unless that version already exists

Preferred npm authentication: configure npm Trusted Publishing for the `@delicious233/dida-cli` package with this GitHub repository and the `release.yml` workflow. This uses GitHub Actions OIDC and does not require a long-lived npm token.

Fallback npm authentication: define `NPM_TOKEN` as a repository secret. The release workflow validates token auth during npm preflight and uses it only when Trusted Publishing is not available.

Emergency manual dispatch may set `allow_changelog_fallback=true`, but normal releases must use an explicit changelog section.

## Release Strategy

The current release workflow stays hand-written through `v0.3.x`. The
GoReleaser decision record is in
[`docs/research/release-strategy-goreleaser.md`](docs/research/release-strategy-goreleaser.md).
Re-evaluate only after archive, checksum, npm provenance, attestation, and
package-manager publishing parity are proven.

## Workflow Dependency Pinning

External GitHub Actions in `.github/workflows/` are pinned to full commit SHAs. Keep the trailing version comment, such as `# v6`, so reviewers can see the intended upstream version when updating the SHA.

Run the pinned-actions validator after workflow changes:

```bash
bash scripts/validate-actions-pinned.sh
```

## Release Archive Provenance

The release workflow generates GitHub artifact attestations from `dist/checksums.txt`. The attestation step uses GitHub OIDC and does not require a signing key or repository secret.

After a release, verify an archive with GitHub CLI:

```bash
gh attestation verify dida_vX.Y.Z_linux_amd64.tar.gz --repo DeliciousBuding/dida-cli
```

## Package Manager Templates

Homebrew and Scoop templates contain SHA-256 checksums from the latest published GitHub Release. Keep those templates on the latest checksum-verified release until the next release assets exist, then update them from `checksums.txt`:

```bash
bash scripts/update-packaging-templates.sh --version vX.Y.Z
bash scripts/validate-packaging.sh --version vX.Y.Z
```

Use `--checksums-file <path>` for both commands when preparing from a downloaded or staged checksum file. Publish to an external Homebrew tap or Scoop bucket only after native package-manager smoke tests.

The release workflow also exports repo-ready layouts after the GitHub Release
exists. Download the `dida-package-manager-repos-vX.Y.Z` workflow artifact, then
use its `homebrew-tap/` and `scoop-bucket/` directories as the roots for the
external repositories after native smoke tests pass.

Before creating or updating external package-manager repositories, run the
preflight. Without `--run-*` flags it exports and checks the layouts, then
prints the native smoke commands. Use the smoke flags only on hosts that have
the package manager installed.

```bash
bash scripts/package-manager-smoke-preflight.sh
bash scripts/package-manager-smoke-preflight.sh --run-homebrew-smoke
bash scripts/package-manager-smoke-preflight.sh --run-scoop-smoke
```

## Post-Release Verification

Verify npm against the official registry. Local mirrors can lag after publish.

```bash
npm view @delicious233/dida-cli@X.Y.Z version readme --registry=https://registry.npmjs.org
npm install @delicious233/dida-cli@X.Y.Z --registry=https://registry.npmjs.org
gh attestation verify dida_vX.Y.Z_linux_amd64.tar.gz --repo DeliciousBuding/dida-cli
```
