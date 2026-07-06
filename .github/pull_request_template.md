## What

<!-- Brief description of the change. -->

## Why

<!-- Motivation or issue reference. Closes #123 -->

## How

<!-- Implementation notes for reviewers. -->

## Checklist

- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] `make staticcheck` passes
- [ ] `go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...` passes
- [ ] `bash scripts/check-private-state.sh` passes
- [ ] Schema updated (if adding/changing commands)
- [ ] Docs updated (if adding/changing commands)
- [ ] CHANGELOG.md updated (if user-facing change)
- [ ] Release/package gates updated (if touching `.github/workflows/`, `scripts/`, `npm/`, or `packaging/`)
