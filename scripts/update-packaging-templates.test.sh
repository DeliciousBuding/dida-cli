#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
updater="$repo_root/scripts/update-packaging-templates.sh"
validator="$repo_root/scripts/validate-packaging.sh"

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
  cp -R "$repo_root/packaging" "$work/packaging"
  (
    cd "$work"
    bash -c "$mutate"
    set +e
    bash "$updater" "$@" >"$out_file" 2>"$err_file"
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

checksums_file="$(mktemp)"
trap 'rm -f "$checksums_file"' EXIT
cat >"$checksums_file" <<'EOF'
1111111111111111111111111111111111111111111111111111111111111111  dida_v9.8.7_windows_amd64.zip
2222222222222222222222222222222222222222222222222222222222222222  dida_v9.8.7_windows_arm64.zip
3333333333333333333333333333333333333333333333333333333333333333  dida_v9.8.7_linux_amd64.tar.gz
4444444444444444444444444444444444444444444444444444444444444444  dida_v9.8.7_linux_arm64.tar.gz
5555555555555555555555555555555555555555555555555555555555555555  dida_v9.8.7_darwin_amd64.tar.gz
6666666666666666666666666666666666666666666666666666666666666666  dida_v9.8.7_darwin_arm64.tar.gz
EOF

run_case "generates package-manager templates" pass "
  :
" --version v9.8.7 --checksums-file "$checksums_file"

work="$(mktemp -d)"
cp -R "$repo_root/packaging" "$work/packaging"
(
  cd "$work"
  bash "$updater" --version v9.8.7 --checksums-file "$checksums_file" >/dev/null
  grep -q 'version "9.8.7"' packaging/homebrew/dida.rb
  grep -q 'dida_v9.8.7_darwin_arm64.tar.gz' packaging/homebrew/dida.rb
  grep -q '6666666666666666666666666666666666666666666666666666666666666666' packaging/homebrew/dida.rb
  node -e '
    const m = require("./packaging/scoop/dida.json");
    if (m.version !== "9.8.7") throw new Error("wrong version");
    if (m.architecture["64bit"].hash !== "1111111111111111111111111111111111111111111111111111111111111111") throw new Error("wrong amd64 hash");
    if (m.architecture.arm64.extract_dir !== "dida_v9.8.7_windows_arm64") throw new Error("wrong arm64 extract_dir");
    if (m.autoupdate.hash.url !== "$baseurl/checksums.txt") throw new Error("wrong hash autoupdate url");
  '
  grep -q 'Current source release: `v9.8.7`' packaging/README.md
  grep -q 'wingetcreate new https://github.com/DeliciousBuding/dida-cli/releases/download/v9.8.7/dida_v9.8.7_windows_amd64.zip' packaging/winget/README.md
  grep -q 'winget validate --manifest <manifest-directory>' packaging/winget/README.md
  grep -q 'scripts/winget-submission-preflight.sh' packaging/winget/README.md
  bash "$validator" --version v9.8.7 --checksums-file "$checksums_file" >/dev/null
)
rm -rf "$work"

run_case "rejects non semver version" fail ":" --version latest --checksums-file "$checksums_file"

bad_checksums="$(mktemp)"
trap 'rm -f "$checksums_file" "$bad_checksums"' EXIT
grep -v 'darwin_arm64' "$checksums_file" >"$bad_checksums"
run_case "rejects incomplete checksums" fail ":" --version v9.8.7 --checksums-file "$bad_checksums"

echo "update-packaging-templates tests passed"
