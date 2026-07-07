#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/package-manager-smoke-preflight.sh"

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
  set +e
  bash "$checker" --output "$work/export" "$@" >"$out_file" 2>"$err_file"
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
  rm -rf "$work"
  rm -f "$out_file" "$err_file"
}

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

bash "$checker" --output "$work/export" >"$work/preflight.out"
test -f "$work/export/homebrew-tap/Formula/dida.rb"
test -f "$work/export/scoop-bucket/bucket/dida.json"
grep -q 'No package-manager install smoke was run' "$work/preflight.out"
grep -q 'brew audit --strict --online --formula' "$work/preflight.out"
grep -q 'scoop install' "$work/preflight.out"
grep -q 'This preflight does not create repositories or publish package-manager channels' "$work/preflight.out"

fake_bin="$work/bin"
mkdir -p "$fake_bin"
cat >"$fake_bin/brew" <<'EOF'
#!/usr/bin/env bash
printf 'brew %s\n' "$*" >>"$FAKE_PM_LOG"
EOF
cat >"$fake_bin/scoop" <<'EOF'
#!/usr/bin/env bash
printf 'scoop %s\n' "$*" >>"$FAKE_PM_LOG"
EOF
cat >"$fake_bin/dida" <<'EOF'
#!/usr/bin/env bash
case "$1" in
  version)
    printf '0.2.5\n'
    ;;
  doctor)
    printf '{"ok": true}\n'
    ;;
  *)
    printf 'unexpected dida command: %s\n' "$*" >&2
    exit 1
    ;;
esac
EOF
chmod +x "$fake_bin/brew" "$fake_bin/scoop" "$fake_bin/dida"

export FAKE_PM_LOG="$work/fake.log"
PATH="$fake_bin:$PATH" bash "$checker" \
  --output "$work/smoke-export" \
  --run-homebrew-smoke \
  --run-scoop-smoke \
  --brew-command brew \
  --scoop-command scoop >"$work/smoke.out"

grep -q 'brew audit --strict --online --formula' "$FAKE_PM_LOG"
grep -q 'brew install --formula' "$FAKE_PM_LOG"
grep -q 'brew test dida' "$FAKE_PM_LOG"
grep -q 'brew uninstall dida' "$FAKE_PM_LOG"
grep -q 'scoop install' "$FAKE_PM_LOG"
grep -q 'scoop uninstall dida' "$FAKE_PM_LOG"
grep -q 'Homebrew smoke passed' "$work/smoke.out"
grep -q 'Scoop smoke passed' "$work/smoke.out"

run_case "missing Homebrew command fails when smoke requested" fail \
  --run-homebrew-smoke --brew-command definitely-not-brew

run_case "missing Scoop command fails when smoke requested" fail \
  --run-scoop-smoke --scoop-command definitely-not-scoop

echo "package-manager smoke preflight tests passed"
