#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-research-audit.sh"

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
  mkdir -p "$work/docs/research" "$work/npm"
  cp "$repo_root/docs/research/prompt-to-artifact-checklist.md" "$work/docs/research/prompt-to-artifact-checklist.md"
  cp "$repo_root/docs/research/roadmap-completion-audit.md" "$work/docs/research/roadmap-completion-audit.md"
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

run_case "current research audits pass" pass ":" --version "$current_version"

run_case "stale prompt release baseline fails" fail "sed -i 's/v0.2.5/v0.2.1/' docs/research/prompt-to-artifact-checklist.md" --version "$current_version"

run_case "stale roadmap release baseline fails" fail "sed -i 's/v0.2.5/v0.1.16/' docs/research/roadmap-completion-audit.md" --version "$current_version"

run_case "missing prompt artifact handoff fails" fail "grep -v 'dida-package-manager-repos-vX.Y.Z' docs/research/prompt-to-artifact-checklist.md > next && mv next docs/research/prompt-to-artifact-checklist.md" --version "$current_version"

run_case "missing roadmap artifact evidence fails" fail "grep -v 'Release workflow exports package-manager repo layouts' docs/research/roadmap-completion-audit.md > next && mv next docs/research/roadmap-completion-audit.md" --version "$current_version"

run_case "non semver version fails" fail ":" --version latest

echo "validate-research-audit tests passed"
