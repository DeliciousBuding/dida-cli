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
ce2a4927d51dca44ebf0f4416d8c938b299f45365fbf5707d1da48b0064342d7  dida_v0.2.6_windows_amd64.zip
f38ca3871d71b8f319da6fb2457eda805858d9fa6452319600028228e1ec921d  dida_v0.2.6_windows_arm64.zip
e60f730dd01343eaa35f79a701a77b944d9d6440236bf8bc7dcfdaae78398dc9  dida_v0.2.6_darwin_amd64.tar.gz
4af818f143b891ed3d55b297de9f52641f22d953e17aa5f7b2cd2804c6a567fc  dida_v0.2.6_darwin_arm64.tar.gz
06b7237d69e7701f997501278645f2199080ba572145a9d831971d367c19a11e  dida_v0.2.6_linux_amd64.tar.gz
aa56c2f60b9b019d89fe684fe5ad3c1bab7eb3df0d4624cefb0462419bed4cdf  dida_v0.2.6_linux_arm64.tar.gz
EOF

current_version="$(node -e 'console.log(require("./packaging/scoop/dida.json").version)')"
current_tag="v${current_version}"

run_case "metadata-only clean templates pass" pass ":" --metadata-only --version "$current_tag"

run_case "metadata-only catches version mismatch" fail '
  node -e "
    const fs = require(\"fs\");
    const path = \"packaging/scoop/dida.json\";
    const data = JSON.parse(fs.readFileSync(path, \"utf8\"));
    data.version = \"9.9.9\";
    fs.writeFileSync(path, JSON.stringify(data, null, 2));
  "
' --metadata-only --version "$current_tag"

run_case "checksum file clean templates pass" pass ":" --version "$current_tag" --checksums-file "$checksums_file"

run_case "checksum file catches hash mismatch" fail "
  sed -i '0,/4af8/s/4af8/0000/' packaging/homebrew/dida.rb
" --version "$current_tag" --checksums-file "$checksums_file"

echo "validate-packaging tests passed"
