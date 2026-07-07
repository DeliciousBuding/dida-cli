# Package Manager Release Artifact

## Task

Add a release workflow handoff artifact for Homebrew tap and Scoop bucket
repository layouts, generated only after release checksums exist.

## Mode

LOCAL_ONLY. This was a focused CI/CD and distribution-governance slice.

## Analysis

- `scripts/export-package-manager-repos.sh` already prepares the external repo
  roots locally from validated templates.
- Future releases should not depend on a maintainer re-running that export from
  an edited working tree.
- GitHub Actions artifacts are a suitable handoff format for release-generated
  files. The project already pins `actions/upload-artifact` and validates pinned
  workflow actions.

## Plan

1. Add a release job after GitHub Release creation.
2. Regenerate packaging templates from the published release `checksums.txt`.
3. Validate the regenerated templates and export repo roots.
4. Upload the export as `dida-package-manager-repos-vX.Y.Z`.
5. Extend governance checks so the job, commands, artifact name, path, and
   retention stay present.
6. Update release, distribution, roadmap, changelog, and archive docs.

## Progress

- [x] Added `package-manager-export` to `.github/workflows/release.yml`.
- [x] Added governance checks for the export job and artifact.
- [x] Added governance mutation coverage for missing export job/artifact.
- [x] Documented the release artifact handoff path.

## Telemetry

| Field | Value |
|---|---|
| Actual effort | S |
| S.U.P.E.R score | S green, U green, P green, E green, R green |
| Unplanned dependencies | 0 |

## Verification

Run before commit:

```bash
bash scripts/validate-repo-governance.sh
bash scripts/validate-repo-governance.test.sh
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
make release-check VERSION=v0.2.5
go test -count=1 ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
bash scripts/check-private-state.sh
git diff --check
```
