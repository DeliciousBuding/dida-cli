#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-npm-package.sh"

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
  mkdir -p "$work/npm"
  (
    cd "$repo_root/npm"
    tar --exclude='./bin/dida.exe' --exclude='./bin/dida-bin' -cf - .
  ) | (
    cd "$work/npm"
    tar -xf -
  )
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

current_version="$(node -e 'console.log(require("./npm/package.json").version)')"
current_tag="v${current_version}"

run_case "current npm package passes" pass ":" --version "$current_tag"

run_case "missing bin wrapper fails" fail "rm npm/bin/dida" --version "$current_tag"

run_case "missing README fails" fail "rm npm/README.md" --version "$current_tag"

run_case "wrong package name fails" fail '
  node -e "
    const fs = require(\"fs\");
    const path = \"npm/package.json\";
    const pkg = JSON.parse(fs.readFileSync(path, \"utf8\"));
    pkg.name = \"bad-package\";
    fs.writeFileSync(path, JSON.stringify(pkg, null, 2));
  "
' --version "$current_tag"

run_case "invalid version fails" fail ":" --version not-semver

echo "validate-npm-package tests passed"
