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

current_version="$(node -e 'console.log(require("./packaging/scoop/dida.json").version)')"
current_tag="v${current_version}"

node - <<'NODE' >"$checksums_file"
const fs = require("fs");

const scoop = JSON.parse(fs.readFileSync("packaging/scoop/dida.json", "utf8"));
const version = scoop.version;
for (const [arch, key] of [["amd64", "64bit"], ["arm64", "arm64"]]) {
  console.log(
    `${scoop.architecture[key].hash}  dida_v${version}_windows_${arch}.zip`
  );
}

const homebrew = fs.readFileSync("packaging/homebrew/dida.rb", "utf8");
const pattern = /url "([^"]+\/(dida_v[^"]+))"\s*\n\s*sha256 "([a-f0-9]{64})"/g;
for (const match of homebrew.matchAll(pattern)) {
  console.log(`${match[3]}  ${match[2]}`);
}
NODE

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

run_case "checksum file catches hash mismatch" fail '
  node -e "
    const fs = require(\"fs\");
    const path = \"packaging/homebrew/dida.rb\";
    const text = fs.readFileSync(path, \"utf8\");
    const mismatched = text.replace(
      /(sha256 \")([0-9a-f])/,
      (_, prefix, first) => prefix + (first === \"0\" ? \"f\" : \"0\")
    );
    fs.writeFileSync(path, mismatched);
  "
' --version "$current_tag" --checksums-file "$checksums_file"

echo "validate-packaging tests passed"
