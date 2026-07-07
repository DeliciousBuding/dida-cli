#!/usr/bin/env bash
set -euo pipefail

site="docs/index.html"

if [[ ! -f "$site" ]]; then
  echo "website homepage is missing: $site" >&2
  exit 1
fi

version="v$(node -e 'console.log(require("./npm/package.json").version)')"

require_text() {
  local pattern="$1"
  local label="$2"
  if ! grep -Eq -- "$pattern" "$site"; then
    echo "$site is missing $label" >&2
    exit 1
  fi
}

reject_text() {
  local pattern="$1"
  local label="$2"
  if grep -Eq -- "$pattern" "$site"; then
    echo "$site contains stale or unsafe website copy: $label" >&2
    exit 1
  fi
}

escaped_version="${version//./\\.}"

require_text "Latest release ${escaped_version}" "current npm/GitHub release label"
require_text 'npm install -g @delicious233/dida-cli' "npm install command"
require_text 'dida auth cookie set --token-stdin --json' "stdin cookie import command"
require_text 'dida doctor --verify --json' "doctor verification command"
require_text 'dida schema list --compact --json' "compact schema command"
require_text 'dida task latest --limit 10 --project inbox --compact --json' "latest task compact read command"
require_text 'dida completion bash' "shell completion command"
require_text 'dida doctor --check-upgrade --json' "upgrade diagnostic command"
require_text 'quickstart\.md' "quickstart link"
require_text 'commands\.md' "command reference link"
require_text 'agent-usage\.md' "agent usage link"
require_text 'github\.com/DeliciousBuding/dida-cli/blob/main/README\.md' "README link"
require_text 'github\.com/DeliciousBuding/dida-cli/blob/main/SECURITY\.md' "security policy link"
require_text 'npmjs\.com/package/@delicious233/dida-cli' "npm package link"

reject_text 'dida \+today --json' "old hero task alias"
reject_text '\.\./assets/' "docs-site parent asset path"

echo "website homepage valid for $version"
