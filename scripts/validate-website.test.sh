#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-website.sh"

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
  mkdir -p "$work/docs" "$work/npm"
  cp "$repo_root/docs/index.html" "$work/docs/index.html"
  cp "$repo_root/npm/package.json" "$work/npm/package.json"
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

run_case "current website passes" pass ":"
run_case "missing current version fails" fail "perl -0pi -e 's/Latest release v[0-9]+\\.[0-9]+\\.[0-9]+/Latest release v0.0.0/' docs/index.html"
run_case "missing npm install fails" fail "grep -v 'npm install -g @delicious233/dida-cli' docs/index.html > next && mv next docs/index.html"
run_case "missing token stdin command fails" fail "grep -v 'dida auth cookie set --token-stdin --json' docs/index.html > next && mv next docs/index.html"
run_case "missing doctor verify command fails" fail "grep -v 'dida doctor --verify --json' docs/index.html > next && mv next docs/index.html"
run_case "missing compact schema command fails" fail "grep -v 'dida schema list --compact --json' docs/index.html > next && mv next docs/index.html"
run_case "missing latest task command fails" fail "grep -v 'dida task latest --limit 10 --project inbox --compact --json' docs/index.html > next && mv next docs/index.html"
run_case "stale old task alias fails" fail "printf '%s\n' 'dida +today --json' >> docs/index.html"
run_case "parent asset path fails" fail "printf '%s\n' '<img src=\"../assets/logo.svg\" alt=\"DidaCLI\">' >> docs/index.html"

echo "validate-website tests passed"
