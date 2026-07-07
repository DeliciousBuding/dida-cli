STATICCHECK_VERSION ?= v0.7.0
GOVULNCHECK_VERSION ?= v1.3.0
CLI_COVER_PROFILE ?= coverage-cli

.PHONY: test vet vuln private-state build install-local coverage-cli actionlint staticcheck release-check

test:
	go test ./...

vet:
	go vet ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

private-state:
	bash scripts/check-private-state.sh

build:
	go build -o bin/dida ./cmd/dida

install-local: build
	mkdir -p "$(HOME)/.local/bin"
	cp bin/dida "$(HOME)/.local/bin/dida"

coverage-cli:
	go test ./internal/cli -coverprofile=$(CLI_COVER_PROFILE)
	go tool cover -func=$(CLI_COVER_PROFILE)

actionlint:
	go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION) ./...

release-check:
ifndef VERSION
	$(error usage: make release-check VERSION=vX.Y.Z)
endif
	bash scripts/validate-release-metadata.sh --tag "$(VERSION)" --skip-git-checks
	bash scripts/validate-release-metadata.test.sh
	bash scripts/validate-changelog.sh --tag "$(VERSION)"
	bash scripts/validate-changelog.test.sh
	$(MAKE) test
	$(MAKE) vet
	$(MAKE) vuln
	$(MAKE) private-state
	bash scripts/validate-website.sh
	bash scripts/validate-website.test.sh
	bash scripts/generate-release-notes.test.sh
	bash scripts/validate-npm-package.sh --version "$(VERSION)"
	bash scripts/validate-npm-package.test.sh
	bash scripts/validate-repo-governance.sh
	bash scripts/validate-repo-governance.test.sh
	bash scripts/validate-actions-pinned.sh
	bash scripts/validate-actions-pinned.test.sh
	bash scripts/update-packaging-templates.test.sh
	bash scripts/export-package-manager-repos.test.sh
	bash scripts/package-manager-smoke-preflight.test.sh
	bash scripts/winget-submission-preflight.test.sh
	bash scripts/validate-packaging.sh --metadata-only
	bash scripts/validate-packaging.test.sh
	bash scripts/verify-release-archives.test.sh
	$(MAKE) staticcheck
	$(MAKE) actionlint
