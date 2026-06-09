#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/validate-packaging.sh"

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
072c213cb22d11a236b21e4ab0a8b86cf840e53661e4b9178bb930128c0faf33  dida_v0.2.2_windows_amd64.zip
59f917bd790b71eea3f314d2650e6d4067d6ce15f80a5b89bc74979735d7a221  dida_v0.2.2_windows_arm64.zip
7329965fab555d840c1d8e96c468441301a4df8629f19b1f0b6f8f8b8467605c  dida_v0.2.2_linux_amd64.tar.gz
3104fdb4efb9916277828a6b853304c684c14df307b7b1859af0896a465ba55a  dida_v0.2.2_linux_arm64.tar.gz
dd5a653c4c6ddeee44f01e08ffa940acba17fb9edcfe4365bb9a6b52277033ad  dida_v0.2.2_darwin_amd64.tar.gz
47fee46e764877034fa769bd6a74ff9d9479f43e07605adc73e23eebfc76b39e  dida_v0.2.2_darwin_arm64.tar.gz
EOF

run_case "metadata-only clean templates pass" pass ":" --metadata-only --version v0.2.2

run_case "metadata-only catches version mismatch" fail '
  node -e "
    const fs = require(\"fs\");
    const path = \"packaging/scoop/dida.json\";
    const data = JSON.parse(fs.readFileSync(path, \"utf8\"));
    data.version = \"9.9.9\";
    fs.writeFileSync(path, JSON.stringify(data, null, 2));
  "
' --metadata-only --version v0.2.2

run_case "checksum file clean templates pass" pass ":" --version v0.2.2 --checksums-file "$checksums_file"

run_case "checksum file catches hash mismatch" fail "
  sed -i '0,/47fe/s/47fe/0000/' packaging/homebrew/dida.rb
" --version v0.2.2 --checksums-file "$checksums_file"

echo "validate-packaging tests passed"
