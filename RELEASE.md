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

`make release-check` validates release metadata, packaging templates, shell release helpers, and workflow syntax. It does not publish anything.

## Automation

Pushing a `vX.Y.Z` tag runs `.github/workflows/release.yml`:

1. validate tag, changelog, npm version, formatting, tests, vet, vulnerability scan, and packaging metadata
2. build Windows, Linux, and macOS archives for amd64 and arm64
3. verify archive shape and binary version
4. create or update the GitHub Release with `checksums.txt`
5. smoke-test npm install on Linux and Windows
6. publish `@delicious233/dida-cli` to npm with provenance, unless that version already exists

Preferred npm authentication: configure npm Trusted Publishing for the `@delicious233/dida-cli` package with this GitHub repository and the `release.yml` workflow. This uses GitHub Actions OIDC and does not require a long-lived npm token.

Fallback npm authentication: define `NPM_TOKEN` as a repository secret. The release workflow validates token auth during npm preflight and uses it only when Trusted Publishing is not available.

Emergency manual dispatch may set `allow_changelog_fallback=true`, but normal releases must use an explicit changelog section.
