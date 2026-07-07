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
ce2a4927d51dca44ebf0f4416d8c938b299f45365fbf5707d1da48b0064342d7  dida_v0.2.6_windows_amd64.zip
f38ca3871d71b8f319da6fb2457eda805858d9fa6452319600028228e1ec921d  dida_v0.2.6_windows_arm64.zip
e60f730dd01343eaa35f79a701a77b944d9d6440236bf8bc7dcfdaae78398dc9  dida_v0.2.6_darwin_amd64.tar.gz
4af818f143b891ed3d55b297de9f52641f22d953e17aa5f7b2cd2804c6a567fc  dida_v0.2.6_darwin_arm64.tar.gz
06b7237d69e7701f997501278645f2199080ba572145a9d831971d367c19a11e  dida_v0.2.6_linux_amd64.tar.gz
aa56c2f60b9b019d89fe684fe5ad3c1bab7eb3df0d4624cefb0462419bed4cdf  dida_v0.2.6_linux_arm64.tar.gz
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
  grep -q 'winget submission preflight passed for DeliciousBuding.DidaCLI v0.2.6' "$work/preflight.out"
  grep -q 'wingetcreate new "https://github.com/DeliciousBuding/dida-cli/releases/download/v0.2.6/dida_v0.2.6_windows_amd64.zip"' "$work/preflight.out"
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
  "grep -v 'dida_v0.2.6_windows_amd64.zip' packaging/winget/README.md > next && mv next packaging/winget/README.md" \
  --checksums-file "$checksums_file"

run_case "invalid package id fails" fail ":" --package-id bad --checksums-file "$checksums_file"

run_case "missing winget command fails when required" fail ":" \
  --require-tools --winget-command definitely-not-winget --checksums-file "$checksums_file"

run_case "missing wingetcreate command fails when required" fail ":" \
  --require-tools --wingetcreate-command definitely-not-wingetcreate --checksums-file "$checksums_file"

run_case "validate requires manifest dir" fail ":" --run-validate --checksums-file "$checksums_file"

echo "winget submission preflight tests passed"
