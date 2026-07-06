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
3431b41cafafe1b059a3629181c540d904ef5b547971017bad9a0872c3ae138c  dida_v0.2.5_windows_amd64.zip
9e9c987cb6d18934f2bad264bed536c34f30ba1b1ccd38007bcf0103238908df  dida_v0.2.5_windows_arm64.zip
4879b1249bc1121c2b6f0c4999f7a9d708018951fb6ff361e75e762eb5fc1799  dida_v0.2.5_linux_amd64.tar.gz
dbaa02791f2e3d63ecd02ea9cec8681c190589c7fbedf846e05e7e1f4acfc916  dida_v0.2.5_linux_arm64.tar.gz
a8f373853dc503488249cd03716ab81e4da87702e3f9fe894a4d2c38ea7d9345  dida_v0.2.5_darwin_amd64.tar.gz
92a23ef9f75ccd21a1770e70ab4f2f33f1e6b3aba81595b3230812f26dec122c  dida_v0.2.5_darwin_arm64.tar.gz
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
  sed -i '0,/92a2/s/92a2/0000/' packaging/homebrew/dida.rb
" --version "$current_tag" --checksums-file "$checksums_file"

echo "validate-packaging tests passed"
