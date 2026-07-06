#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-release-metadata.sh"

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
  mkdir -p "$work/npm"
  cp "$repo_root/npm/package.json" "$work/npm/package.json"
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

run_case "valid metadata passes" pass ":" --tag v0.2.4 --skip-git-checks

run_case "non-semver tag fails" fail ":" --tag 0.2.4 --skip-git-checks

run_case "npm package version mismatch fails" fail '
  node -e "
    const fs = require(\"fs\");
    const path = \"npm/package.json\";
    const data = JSON.parse(fs.readFileSync(path, \"utf8\"));
    data.version = \"9.9.9\";
    fs.writeFileSync(path, JSON.stringify(data, null, 2));
  "
' --tag v0.2.4 --skip-git-checks

run_case "missing changelog section fails by default" fail ":" --tag v9.9.9 --skip-git-checks

run_case "missing changelog section can be allowed explicitly" pass '
  node -e "
    const fs = require(\"fs\");
    const path = \"npm/package.json\";
    const data = JSON.parse(fs.readFileSync(path, \"utf8\"));
    data.version = \"9.9.9\";
    fs.writeFileSync(path, JSON.stringify(data, null, 2));
  "
' --tag v9.9.9 --skip-git-checks --allow-changelog-fallback

echo "validate-release-metadata tests passed"
