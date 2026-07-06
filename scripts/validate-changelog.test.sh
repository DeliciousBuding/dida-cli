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

current_tag="$(sed -n 's/^## \[\(v[0-9][^]]*\)\].*/\1/p' "$repo_root/CHANGELOG.md" | head -n 1)"

run_case "current changelog passes" pass ":" --tag "$current_tag"

run_case "missing unreleased link fails" fail "grep -v '^\\[Unreleased\\]:' CHANGELOG.md > next && mv next CHANGELOG.md"

run_case "missing required tag fails" fail "sed -i '/^## \\[${current_tag}\\]/d' CHANGELOG.md" --tag "$current_tag"

run_case "missing compare link fails" fail "grep -v '^\\[${current_tag}\\]:' CHANGELOG.md > next && mv next CHANGELOG.md"

echo "validate-changelog tests passed"
