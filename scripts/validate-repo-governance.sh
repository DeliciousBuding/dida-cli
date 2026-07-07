#!/usr/bin/env bash
set -euo pipefail

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    echo "required governance file is missing: $path" >&2
    exit 1
  fi
}

require_text() {
  local path="$1"
  local pattern="$2"
  local label="$3"
  if ! grep -Eq -- "$pattern" "$path"; then
    echo "$path is missing $label" >&2
    exit 1
  fi
}

require_file README.md
require_file Makefile
require_file CONTRIBUTING.md
require_file CODE_OF_CONDUCT.md
require_file SECURITY.md
require_file RELEASE.md
require_file docs/distribution.md
require_file npm/README.md
require_file scripts/package-manager-smoke-preflight.sh
require_file scripts/package-manager-smoke-preflight.test.sh
require_file scripts/winget-submission-preflight.sh
require_file scripts/winget-submission-preflight.test.sh
require_file .github/pull_request_template.md
require_file .github/ISSUE_TEMPLATE/bug_report.yml
require_file .github/ISSUE_TEMPLATE/feature_request.yml
require_file .github/ISSUE_TEMPLATE/config.yml
require_file .github/workflows/codeql.yml
require_file .github/workflows/scorecard.yml
require_file .github/workflows/release.yml

if head -n 1 README.md | grep -qx -- '---'; then
  echo "README.md must not start with internal YAML frontmatter" >&2
  exit 1
fi

require_text npm/README.md 'npm install -g @delicious233/dida-cli' "npm README install command"
require_text npm/README.md 'dida auth cookie set --token-stdin' "npm README auth setup command"
require_text npm/README.md 'Do not paste cookies or tokens into shell history' "npm README token warning"

for heading in '## What' '## Why' '## How' '## Checklist'; do
  require_text .github/pull_request_template.md "^${heading}$" "PR template heading ${heading}"
done

require_text .github/pull_request_template.md 'go test \./\.\.\.' "go test checklist item"
require_text .github/pull_request_template.md 'go vet \./\.\.\.' "go vet checklist item"
require_text .github/pull_request_template.md 'make staticcheck' "staticcheck checklist item"
require_text .github/pull_request_template.md 'govulncheck' "govulncheck checklist item"
require_text .github/pull_request_template.md 'check-private-state\.sh' "private-state checklist item"
require_text .github/pull_request_template.md 'CHANGELOG\.md' "changelog checklist item"
require_text .github/pull_request_template.md 'docs updated|Docs updated' "docs checklist item"

for id in version channel command expected actual doctor os; do
  require_text .github/ISSUE_TEMPLATE/bug_report.yml "id: ${id}$" "bug report field ${id}"
done
require_text .github/ISSUE_TEMPLATE/bug_report.yml 'redact any tokens|Do not paste' "bug report secret warning"

for id in problem solution channel alternatives; do
  require_text .github/ISSUE_TEMPLATE/feature_request.yml "id: ${id}$" "feature request field ${id}"
done

require_text .github/ISSUE_TEMPLATE/config.yml 'blank_issues_enabled: true' "blank issue setting"
require_text .github/ISSUE_TEMPLATE/config.yml 'security/advisories/new' "private security advisory contact link"
require_text .github/ISSUE_TEMPLATE/config.yml 'docs/commands.md' "command reference contact link"
require_text SECURITY.md 'private security advisory' "private advisory reporting path"
require_text CONTRIBUTING.md 'scripts/check-private-state\.sh|check-private-state\.sh' "private-state contribution guidance"
require_text CONTRIBUTING.md 'CODE_OF_CONDUCT\.md' "code of conduct contribution guidance"
require_text CONTRIBUTING.md 'SECURITY\.md' "security reporting contribution guidance"
require_text CODE_OF_CONDUCT.md 'Do not post cookies, tokens' "secret-sharing conduct rule"
require_text CODE_OF_CONDUCT.md 'SECURITY\.md' "security policy reference"

require_text .github/workflows/codeql.yml 'github/codeql-action/init@[a-f0-9]{40}[[:space:]]*# v[0-9]+' "pinned CodeQL init action"
require_text .github/workflows/codeql.yml 'github/codeql-action/analyze@[a-f0-9]{40}[[:space:]]*# v[0-9]+' "pinned CodeQL analyze action"
require_text .github/workflows/codeql.yml 'languages: go' "CodeQL Go language"
require_text .github/workflows/codeql.yml 'security-events: write' "CodeQL security-events permission"
require_text .github/workflows/codeql.yml 'security-extended,security-and-quality' "CodeQL extended query suite"

require_text .github/workflows/scorecard.yml 'ossf/scorecard-action@[a-f0-9]{40}[[:space:]]*# v[0-9]' "pinned OpenSSF Scorecard action"
require_text .github/workflows/scorecard.yml 'publish_results: true' "Scorecard public results publishing"
require_text .github/workflows/scorecard.yml 'github/codeql-action/upload-sarif@[a-f0-9]{40}[[:space:]]*# v[0-9]+' "pinned Scorecard SARIF upload"
require_text .github/workflows/scorecard.yml 'security-events: write' "Scorecard security-events permission"
require_text .github/workflows/scorecard.yml 'id-token: write' "Scorecard OIDC permission"
require_text .github/workflows/scorecard.yml 'actions/upload-artifact@[a-f0-9]{40}[[:space:]]*# v[5-9]' "pinned Scorecard SARIF artifact upload"
require_text .github/workflows/scorecard.yml 'actions/download-artifact@[a-f0-9]{40}[[:space:]]*# v[5-9]' "pinned Scorecard SARIF artifact download"
require_text .github/workflows/release.yml 'actions/upload-artifact@[a-f0-9]{40}[[:space:]]*# v[5-9]' "pinned upload-artifact action"
require_text .github/workflows/release.yml 'actions/download-artifact@[a-f0-9]{40}[[:space:]]*# v[5-9]' "pinned download-artifact action"
require_text .github/workflows/release.yml 'attestations: write' "release attestation permission"
require_text .github/workflows/release.yml 'id-token: write' "release OIDC permission"
require_text .github/workflows/release.yml 'actions/attest@[a-f0-9]{40}[[:space:]]*# v[0-9]+' "pinned release attestation action"
require_text .github/workflows/release.yml 'subject-checksums: dist/checksums\.txt' "release archive checksum attestation input"
require_text .github/workflows/release.yml 'package-manager-export:' "package-manager export job"
require_text .github/workflows/release.yml 'scripts/update-packaging-templates\.sh --version "\$\{GITHUB_REF_NAME\}"' "release package-manager template update"
require_text .github/workflows/release.yml 'scripts/export-package-manager-repos\.sh' "release package-manager repo export"
require_text .github/workflows/release.yml 'name: dida-package-manager-repos-\$\{\{ github\.ref_name \}\}' "package-manager artifact name"
require_text .github/workflows/release.yml 'path: dist/package-manager-repos' "package-manager artifact path"
require_text .github/workflows/release.yml 'retention-days: 30' "package-manager artifact retention"
require_text Makefile 'package-manager-smoke-preflight\.test\.sh' "package-manager smoke preflight tests"
require_text Makefile 'winget-submission-preflight\.test\.sh' "winget submission preflight tests"
require_text packaging/winget/README.md 'winget validate --manifest' "winget validation command"
require_text Makefile '^staticcheck:' "Makefile staticcheck target"
require_text Makefile 'honnef\.co/go/tools/cmd/staticcheck@\$\(STATICCHECK_VERSION\)' "pinned Staticcheck make target"
require_text Makefile '\$\(MAKE\) test' "release-check test gate"
require_text Makefile '\$\(MAKE\) vet' "release-check vet gate"
require_text Makefile '\$\(MAKE\) vuln' "release-check vulnerability gate"
require_text Makefile '\$\(MAKE\) private-state' "release-check private-state gate"
require_text .github/workflows/ci.yml 'STATICCHECK_VERSION: v[0-9]+\.[0-9]+\.[0-9]+' "CI Staticcheck version"
require_text .github/workflows/ci.yml 'honnef\.co/go/tools/cmd/staticcheck@\$\{STATICCHECK_VERSION\}' "CI Staticcheck command"
require_text .github/workflows/release.yml 'STATICCHECK_VERSION: v[0-9]+\.[0-9]+\.[0-9]+' "release Staticcheck version"
require_text .github/workflows/release.yml 'honnef\.co/go/tools/cmd/staticcheck@\$\{STATICCHECK_VERSION\}' "release Staticcheck command"

echo "repository governance files valid"
