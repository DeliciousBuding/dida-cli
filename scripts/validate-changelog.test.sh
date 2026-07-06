#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-changelog.sh"

run_case() {
  local name="$1"
  local expected="$2"
  local mutate="$3"
  shift 3
  local work
  local out_file
  local err_file
  work="$(mktemp -d)"
  out_file="$(mktemp)"
  err_file="$(mktemp)"
  cp "$repo_root/CHANGELOG.md" "$work/CHANGELOG.md"
  (
    cd "$work"
    bash -c "$mutate"
    set +e
    bash "$checker" "$@" >"$out_file" 2>"$err_file"
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

run_case "current changelog passes" pass ":" --tag v0.2.4

run_case "missing unreleased link fails" fail "grep -v '^\\[Unreleased\\]:' CHANGELOG.md > next && mv next CHANGELOG.md"

run_case "missing required tag fails" fail "sed -i '/^## \\[v0.2.4\\]/d' CHANGELOG.md" --tag v0.2.4

run_case "missing compare link fails" fail "grep -v '^\\[v0.2.4\\]:' CHANGELOG.md > next && mv next CHANGELOG.md"

echo "validate-changelog tests passed"
