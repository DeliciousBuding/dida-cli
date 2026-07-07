#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-repo-governance.sh"

run_case() {
  local name="$1"
  local expected="$2"
  local mutate="$3"
  local work
  local out_file
  local err_file
  work="$(mktemp -d)"
  out_file="$(mktemp)"
  err_file="$(mktemp)"
  (
    cd "$repo_root"
    tar -cf - README.md Makefile CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md RELEASE.md docs/distribution.md npm/README.md packaging scripts .github
  ) | (
    cd "$work"
    tar -xf -
  )
  (
    cd "$work"
    bash -c "$mutate"
    set +e
    bash "$checker" >"$out_file" 2>"$err_file"
    code="$?"
    set -e
    if [[ "$expected" == "pass" && "$code" -ne 0 ]]; then
      cat "$out_file"
      cat "$err_file" >&2
      echo "$name: expected pass, got exit $code" >&2
      exit 1
    fi
    if [[ "$expected" == "fail" && "$code" -eq 0 ]]; then
      cat "$out_file"
      cat "$err_file" >&2
      echo "$name: expected fail, got pass" >&2
      exit 1
    fi
  )
  rm -rf "$work"
  rm -f "$out_file" "$err_file"
}

run_case "current governance files pass" pass ":"

run_case "README frontmatter fails" fail "{ printf '%s\n' '---' 'internal: true' '---'; cat README.md; } > next && mv next README.md"

run_case "missing PR private-state gate fails" fail "grep -v 'check-private-state' .github/pull_request_template.md > next && mv next .github/pull_request_template.md"

run_case "missing PR staticcheck gate fails" fail "grep -v 'make staticcheck' .github/pull_request_template.md > next && mv next .github/pull_request_template.md"

run_case "missing npm README token warning fails" fail "grep -v 'Do not paste cookies' npm/README.md > next && mv next npm/README.md"

run_case "missing code of conduct fails" fail "rm CODE_OF_CONDUCT.md"

run_case "missing security contact link fails" fail "grep -v 'security/advisories/new' .github/ISSUE_TEMPLATE/config.yml > next && mv next .github/ISSUE_TEMPLATE/config.yml"

run_case "missing CodeQL workflow fails" fail "rm .github/workflows/codeql.yml"

run_case "missing Scorecard security-events permission fails" fail "grep -v 'security-events: write' .github/workflows/scorecard.yml > next && mv next .github/workflows/scorecard.yml"

run_case "unpinned Scorecard action fails" fail "sed -i '0,/ossf\\/scorecard-action@[a-f0-9]\\{40\\}/s//ossf\\/scorecard-action@v2.4.3/' .github/workflows/scorecard.yml"

run_case "missing release attestation action fails" fail "grep -v 'actions/attest' .github/workflows/release.yml > next && mv next .github/workflows/release.yml"

run_case "missing package-manager export job fails" fail "grep -v 'package-manager-export' .github/workflows/release.yml > next && mv next .github/workflows/release.yml"

run_case "missing package-manager artifact upload fails" fail "grep -v 'dida-package-manager-repos' .github/workflows/release.yml > next && mv next .github/workflows/release.yml"

run_case "missing CI staticcheck command fails" fail "grep -v 'staticcheck' .github/workflows/ci.yml > next && mv next .github/workflows/ci.yml"

run_case "missing Makefile staticcheck target fails" fail "grep -v 'staticcheck' Makefile > next && mv next Makefile"

run_case "missing release-check test gate fails" fail "sed -i '/\$(MAKE) test/d' Makefile"

run_case "missing package-manager smoke preflight test fails" fail "grep -v 'package-manager-smoke-preflight' Makefile > next && mv next Makefile"

run_case "missing package-manager smoke preflight script fails" fail "rm scripts/package-manager-smoke-preflight.sh"

run_case "missing winget submission preflight test fails" fail "grep -v 'winget-submission-preflight' Makefile > next && mv next Makefile"

run_case "missing winget validation command fails" fail "grep -v 'winget validate' packaging/winget/README.md > next && mv next packaging/winget/README.md"

echo "validate-repo-governance tests passed"
