#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-release-strategy.sh"

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
  mkdir -p "$work/docs/research"
  cp "$repo_root/docs/research/release-strategy-goreleaser.md" "$work/docs/research/release-strategy-goreleaser.md"
  cp "$repo_root/ROADMAP.md" "$work/ROADMAP.md"
  cp "$repo_root/RELEASE.md" "$work/RELEASE.md"
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

run_case "current release strategy passes" pass ":"
run_case "missing decision file fails" fail "rm docs/research/release-strategy-goreleaser.md"
run_case "missing dry-run condition fails" fail "grep -v 'GoReleaser dry run produces the same archive names' docs/research/release-strategy-goreleaser.md > next && mv next docs/research/release-strategy-goreleaser.md"
run_case "stale undecided roadmap fails" fail "sed -i '/What.s NOT Done Yet/a - goreleaser migration is undecided' ROADMAP.md"
run_case "missing deferred roadmap row fails" fail "grep -v 'goreleaser migration | Deferred' ROADMAP.md > next && mv next ROADMAP.md"
run_case "missing release reference fails" fail "grep -v 'release-strategy-goreleaser.md' RELEASE.md > next && mv next RELEASE.md"

echo "validate-release-strategy tests passed"
