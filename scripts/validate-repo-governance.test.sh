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
    tar -cf - README.md CONTRIBUTING.md SECURITY.md RELEASE.md npm/README.md .github
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

run_case "missing npm README token warning fails" fail "grep -v 'Do not paste cookies' npm/README.md > next && mv next npm/README.md"

run_case "missing CodeQL workflow fails" fail "rm .github/workflows/codeql.yml"

run_case "missing Scorecard security-events permission fails" fail "grep -v 'security-events: write' .github/workflows/scorecard.yml > next && mv next .github/workflows/scorecard.yml"

run_case "unpinned Scorecard action fails" fail "sed -i '0,/ossf\\/scorecard-action@[a-f0-9]\\{40\\}/s//ossf\\/scorecard-action@v2.4.3/' .github/workflows/scorecard.yml"

run_case "missing release attestation action fails" fail "grep -v 'actions/attest' .github/workflows/release.yml > next && mv next .github/workflows/release.yml"

echo "validate-repo-governance tests passed"
