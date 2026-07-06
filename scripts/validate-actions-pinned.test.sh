#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-actions-pinned.sh"

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
  mkdir -p "$work/.github"
  cp -R "$repo_root/.github/workflows" "$work/.github/workflows"
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

run_case "current workflows pass" pass ":"

run_case "tag reference fails" fail "sed -i '0,/actions\\/checkout@[a-f0-9]\\{40\\}/s//actions\\/checkout@v6/' .github/workflows/ci.yml"

run_case "missing version comment fails" fail "sed -i '0,/ # v6/s///' .github/workflows/ci.yml"

echo "validate-actions-pinned tests passed"
