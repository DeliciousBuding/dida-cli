#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/check-private-state.sh"

run_case() {
  local name="$1"
  local expected="$2"
  shift 2
  local work
  local out_file
  local err_file
  work="$(mktemp -d)"
  out_file="$(mktemp)"
  err_file="$(mktemp)"
  (
    cd "$work"
    git init -q
    git config user.email test@example.com
    git config user.name test
    printf 'ok\n' > README.md
    git add README.md
    git commit -q -m init
    "$@"
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

run_case "clean repo" pass true

run_case "documented home-relative examples pass" pass bash -c '
  printf "Credentials stay in ~/.dida-cli/.\nInstall skills under ~/.openclaw/skills.\n" > docs.md
  git add docs.md
'

run_case "shell regex snippets pass" pass bash -c '
  printf "%s\n" "sed -n \"s/^dida_\\\\(v[^_]*\\\\)_linux_amd64\\\\.tar.gz$/\\\\1/p\"" > install-note.sh
  git add install-note.sh
'

run_case "tracked env fails" fail bash -c '
  printf "TOKEN=x\n" > .env
  git add -f .env
'

run_case "forced binary path fails" fail bash -c '
  mkdir -p npm/bin
  printf "binary\n" > npm/bin/dida-bin
  git add -f npm/bin/dida-bin
'

run_case "untracked secret fails" fail bash -c '
  printf "access_%s=abcdefghijklmnopqrstuvwxyz123456\n" "token" > leak.txt
'

run_case "nested credential filename fails" fail bash -c '
  mkdir -p fixtures
  printf "{}\n" > fixtures/openapi-oauth.json
  git add -f fixtures/openapi-oauth.json
'

run_case "local path fails" fail bash -c '
  printf "path=D:/%s/private/file.txt\n" "Code" > note.txt
  git add note.txt
'

echo "check-private-state tests passed"
