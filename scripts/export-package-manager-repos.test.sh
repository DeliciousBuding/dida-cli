#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
exporter="$repo_root/scripts/export-package-manager-repos.sh"

assert_same_file() {
  local want="$1"
  local got="$2"
  if ! git -c core.autocrlf=false diff --no-index --quiet "$want" "$got"; then
    echo "file differs: $got" >&2
    exit 1
  fi
}

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
  bash "$exporter" --output "$work/export" "$@" >"$out_file" 2>"$err_file"
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

bash "$exporter" --output "$work/export" >/dev/null

homebrew_dir="$work/export/homebrew-tap"
scoop_dir="$work/export/scoop-bucket"

test -f "$homebrew_dir/Formula/dida.rb"
test -f "$homebrew_dir/README.md"
test -f "$homebrew_dir/LICENSE"
test -f "$scoop_dir/bucket/dida.json"
test -f "$scoop_dir/README.md"
test -f "$scoop_dir/LICENSE"

assert_same_file "$repo_root/packaging/homebrew/dida.rb" "$homebrew_dir/Formula/dida.rb"
assert_same_file "$repo_root/packaging/scoop/dida.json" "$scoop_dir/bucket/dida.json"
assert_same_file "$repo_root/LICENSE" "$homebrew_dir/LICENSE"
assert_same_file "$repo_root/LICENSE" "$scoop_dir/LICENSE"

grep -q 'Generated from `DeliciousBuding/dida-cli` release `v0.2.5`' "$homebrew_dir/README.md"
grep -q 'brew tap DeliciousBuding/dida' "$homebrew_dir/README.md"
grep -q 'brew install dida' "$homebrew_dir/README.md"
grep -q 'Create the GitHub repository `DeliciousBuding/homebrew-dida`' "$homebrew_dir/README.md"
grep -q 'source template is' "$homebrew_dir/README.md"

grep -q 'Generated from `DeliciousBuding/dida-cli` release `v0.2.5`' "$scoop_dir/README.md"
grep -q 'scoop bucket add dida https://github.com/DeliciousBuding/scoop-bucket' "$scoop_dir/README.md"
grep -q 'scoop install dida' "$scoop_dir/README.md"
grep -q 'Create the GitHub repository `DeliciousBuding/scoop-bucket`' "$scoop_dir/README.md"
grep -q 'source template is' "$scoop_dir/README.md"

case "$(grep -R -E 'C:\\|/Users/|/home/|TOKEN|SECRET|cookie' "$work/export" || true)" in
  "")
    ;;
  *)
    echo "export contains local paths or secret-like words" >&2
    exit 1
    ;;
esac

custom="$(mktemp -d)"
trap 'rm -rf "$work" "$custom"' EXIT
bash "$exporter" \
  --output "$custom/export" \
  --homebrew-repo Example/homebrew-tools \
  --scoop-repo Example/scoop-extras \
  --scoop-bucket extras >/dev/null
grep -q 'brew tap Example/tools' "$custom/export/homebrew-tap/README.md"
grep -q 'scoop bucket add extras https://github.com/Example/scoop-extras' "$custom/export/scoop-bucket/README.md"

run_case "rejects invalid homebrew repo" fail --homebrew-repo bad-name
run_case "rejects invalid scoop repo" fail --scoop-repo bad-name
run_case "rejects invalid scoop bucket name" fail --scoop-bucket 'bad/name'

echo "export-package-manager-repos tests passed"
