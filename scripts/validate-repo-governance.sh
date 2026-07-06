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
require_file CONTRIBUTING.md
require_file SECURITY.md
require_file RELEASE.md
require_file npm/README.md
require_file .github/pull_request_template.md
require_file .github/ISSUE_TEMPLATE/bug_report.yml
require_file .github/ISSUE_TEMPLATE/feature_request.yml
require_file .github/ISSUE_TEMPLATE/config.yml

if head -n 1 README.md | grep -qx -- '---'; then
  echo "README.md must not start with internal YAML frontmatter" >&2
  exit 1
fi

require_text README.md 'npm (i|install) -g @delicious233/dida-cli' "npm install command"
require_text README.md 'dida doctor --verify --json' "doctor verification command"
require_text README.md 'dida schema list --compact --json' "schema discovery command"
require_text README.md '--dry-run' "dry-run guidance"
require_text npm/README.md 'npm install -g @delicious233/dida-cli' "npm README install command"
require_text npm/README.md 'dida auth cookie set --token-stdin' "npm README auth setup command"
require_text npm/README.md 'Do not paste cookies or tokens into shell history' "npm README token warning"

for heading in '## What' '## Why' '## How' '## Checklist'; do
  require_text .github/pull_request_template.md "^${heading}$" "PR template heading ${heading}"
done

require_text .github/pull_request_template.md 'go test \./\.\.\.' "go test checklist item"
require_text .github/pull_request_template.md 'go vet \./\.\.\.' "go vet checklist item"
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
require_text .github/ISSUE_TEMPLATE/config.yml 'docs/commands.md' "command reference contact link"
require_text SECURITY.md 'private security advisory' "private advisory reporting path"
require_text CONTRIBUTING.md 'scripts/check-private-state\.sh|check-private-state\.sh' "private-state contribution guidance"

echo "repository governance files valid"
