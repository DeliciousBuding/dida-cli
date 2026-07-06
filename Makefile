.PHONY: test build install-local actionlint release-check

test:
	go test ./...

build:
	go build -o bin/dida ./cmd/dida

install-local: build
	mkdir -p "$(HOME)/.local/bin"
	cp bin/dida "$(HOME)/.local/bin/dida"

actionlint:
	go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12

release-check:
ifndef VERSION
	$(error usage: make release-check VERSION=vX.Y.Z)
endif
	bash scripts/validate-release-metadata.sh --tag "$(VERSION)" --skip-git-checks
	bash scripts/validate-release-metadata.test.sh
	bash scripts/validate-changelog.sh --tag "$(VERSION)"
	bash scripts/validate-changelog.test.sh
	bash scripts/generate-release-notes.test.sh
	bash scripts/validate-npm-package.sh --version "$(VERSION)"
	bash scripts/validate-npm-package.test.sh
	bash scripts/validate-repo-governance.sh
	bash scripts/validate-repo-governance.test.sh
	bash scripts/validate-packaging.sh --metadata-only --version "$(VERSION)"
	bash scripts/validate-packaging.test.sh
	bash scripts/verify-release-archives.test.sh
	$(MAKE) actionlint
