#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/winget-submission-preflight.sh"

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
  cp -R "$repo_root/scripts" "$work/scripts"
  (
    cd "$work"
    git init -q
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

checksums_file="$(mktemp)"
trap 'rm -f "$checksums_file"' EXIT
cat >"$checksums_file" <<'EOF'
3431b41cafafe1b059a3629181c540d904ef5b547971017bad9a0872c3ae138c  dida_v0.2.5_windows_amd64.zip
9e9c987cb6d18934f2bad264bed536c34f30ba1b1ccd38007bcf0103238908df  dida_v0.2.5_windows_arm64.zip
4879b1249bc1121c2b6f0c4999f7a9d708018951fb6ff361e75e762eb5fc1799  dida_v0.2.5_linux_amd64.tar.gz
dbaa02791f2e3d63ecd02ea9cec8681c190589c7fbedf846e05e7e1f4acfc916  dida_v0.2.5_linux_arm64.tar.gz
a8f373853dc503488249cd03716ab81e4da87702e3f9fe894a4d2c38ea7d9345  dida_v0.2.5_darwin_amd64.tar.gz
92a23ef9f75ccd21a1770e70ab4f2f33f1e6b3aba81595b3230812f26dec122c  dida_v0.2.5_darwin_arm64.tar.gz
EOF

run_case "current winget preflight passes" pass ":" --checksums-file "$checksums_file"

work="$(mktemp -d)"
trap 'rm -rf "$work"; rm -f "$checksums_file"' EXIT
cp -R "$repo_root/packaging" "$work/packaging"
cp -R "$repo_root/scripts" "$work/scripts"
(
  cd "$work"
  git init -q
  bash "$checker" --checksums-file "$checksums_file" >"$work/preflight.out"
  grep -q 'winget submission preflight passed for DeliciousBuding.DidaCLI v0.2.5' "$work/preflight.out"
  grep -q 'wingetcreate new "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.5/dida_v0.2.5_windows_amd64.zip"' "$work/preflight.out"
  grep -q 'winget validate --manifest <manifest-directory>' "$work/preflight.out"
  grep -q 'does not generate manifests or submit packages' "$work/preflight.out"
)

fake_bin="$work/bin"
manifest_dir="$work/manifests"
mkdir -p "$fake_bin" "$manifest_dir"
cat >"$fake_bin/winget" <<'EOF'
#!/usr/bin/env bash
printf 'winget %s\n' "$*" >>"$FAKE_WINGET_LOG"
EOF
cat >"$fake_bin/wingetcreate" <<'EOF'
#!/usr/bin/env bash
printf 'wingetcreate %s\n' "$*" >>"$FAKE_WINGET_LOG"
EOF
chmod +x "$fake_bin/winget" "$fake_bin/wingetcreate"
export FAKE_WINGET_LOG="$work/winget.log"
(
  cd "$work"
  PATH="$fake_bin:$PATH" bash "$checker" \
    --checksums-file "$checksums_file" \
    --manifest-dir "$manifest_dir" \
    --run-validate \
    --winget-command winget >"$work/validate.out"
)
grep -q 'winget validate --manifest' "$FAKE_WINGET_LOG"
grep -q 'winget manifest validation passed' "$work/validate.out"

run_case "missing winget URL fails" fail \
  "grep -v 'dida_v0.2.5_windows_amd64.zip' packaging/winget/README.md > next && mv next packaging/winget/README.md" \
  --checksums-file "$checksums_file"

run_case "invalid package id fails" fail ":" --package-id bad --checksums-file "$checksums_file"

run_case "missing winget command fails when required" fail ":" \
  --require-tools --winget-command definitely-not-winget --checksums-file "$checksums_file"

run_case "missing wingetcreate command fails when required" fail ":" \
  --require-tools --wingetcreate-command definitely-not-wingetcreate --checksums-file "$checksums_file"

run_case "validate requires manifest dir" fail ":" --run-validate --checksums-file "$checksums_file"

echo "winget submission preflight tests passed"
