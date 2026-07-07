#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-roadmap.sh"

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
  cp "$repo_root/ROADMAP.md" "$work/ROADMAP.md"
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

current_version="$(node -e 'console.log("v" + require("./npm/package.json").version)')"

run_case "current roadmap passes" pass ":" --version "$current_version"

run_case "stale latest release fails" fail "grep -v '^Latest release:' ROADMAP.md > next && mv next ROADMAP.md" --version "$current_version"

run_case "stale baseline heading fails" fail "sed -i 's/^## Current Baseline.*/## Current Baseline (as of v0.0.0)/' ROADMAP.md" --version "$current_version"

run_case "stale immediate release block fails" fail "sed -i '/## Current Best Next Tasks/a For v0.2.1 release (immediate):' ROADMAP.md" --version "$current_version"

run_case "missing next milestone fails" fail "sed -i '/For v0.3.0 (next milestone):/d' ROADMAP.md" --version "$current_version"

run_case "stale release workflow status fails" fail 'printf "%s\n" "Status: implemented and smoke-tested through \`v0.2.1\`" >> ROADMAP.md' --version "$current_version"

run_case "stale installer smoke status fails" fail 'printf "%s\n" "installer smoke passed against \`v0.1.16\`" >> ROADMAP.md' --version "$current_version"

run_case "missing npm current package status fails" fail "grep -v '@delicious233/dida-cli@0.2.5' ROADMAP.md > next && mv next ROADMAP.md" --version "$current_version"

run_case "missing package manager artifact handoff fails" fail "grep -v 'dida-package-manager-repos-vX.Y.Z' ROADMAP.md > next && mv next ROADMAP.md" --version "$current_version"

run_case "missing winget submission preflight fails" fail "grep -v 'winget-submission-preflight' ROADMAP.md > next && mv next ROADMAP.md" --version "$current_version"

run_case "non semver version fails" fail ":" --version 0.2.5

echo "validate-roadmap tests passed"
