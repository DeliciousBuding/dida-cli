#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
checker="$repo_root/scripts/verify-release-archives.sh"
version="v9.8.7"

make_zip() {
  local archive="$1"
  local folder="$2"
  local binary="$3"
  local extra="${4:-}"
  python3 - "$archive" "$folder" "$binary" "$extra" <<'PY'
import sys, zipfile
archive, folder, binary, extra = sys.argv[1:5]
with zipfile.ZipFile(archive, "w") as zf:
    zf.writestr(f"{folder}/{binary}", "binary\n")
    if extra:
        zf.writestr(f"{folder}/{extra}", "extra\n")
PY
}

make_tar() {
  local archive="$1"
  local folder="$2"
  local binary="$3"
  local executable="$4"
  local extra="${5:-}"
  python3 - "$archive" "$folder" "$binary" "$executable" "$extra" <<'PY'
import io
import sys
import tarfile

archive, folder, binary, executable, extra = sys.argv[1:6]

def add_file(tar, name, body, mode):
    data = body.encode("utf-8")
    info = tarfile.TarInfo(name)
    info.size = len(data)
    info.mode = mode
    tar.addfile(info, io.BytesIO(data))

with tarfile.open(archive, "w:gz") as tar:
    mode = 0o755 if executable == "yes" else 0o644
    add_file(tar, f"{folder}/{binary}", "binary\n", mode)
    if extra:
        add_file(tar, f"{folder}/{extra}", "extra\n", 0o644)
PY
}

make_dist() {
  local dist="$1"
  local unix_executable="${2:-yes}"
  local extra_asset="${3:-}"
  mkdir -p "$dist"
  for arch in amd64 arm64; do
    folder="dida_${version}_windows_${arch}"
    make_zip "$dist/${folder}.zip" "$folder" "dida.exe"
  done
  for os in linux darwin; do
    for arch in amd64 arm64; do
      folder="dida_${version}_${os}_${arch}"
      make_tar "$dist/${folder}.tar.gz" "$folder" "dida" "$unix_executable"
    done
  done
  if [[ -n "$extra_asset" ]]; then
    cp "$dist/dida_${version}_windows_amd64.zip" "$dist/$extra_asset"
  fi
}

run_case() {
  local name="$1"
  local expected="$2"
  local mutate="$3"
  local work
  local out_file
  local err_file
  work="$(mktemp -d)"
  out_file="$(mktemp)"
  err_file="$(mktemp)"
  make_dist "$work/dist"
  (
    cd "$work"
    eval "$mutate"
    set +e
    bash "$checker" "$version" "$work/dist" >"$out_file" 2>"$err_file"
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

run_case "complete archives pass" pass ":"

run_case "missing archive fails" fail "rm dist/dida_${version}_linux_arm64.tar.gz"

run_case "extra archive fails" fail "make_dist dist yes dida_${version}_linux_amd64_extra.tar.gz"

run_case "extra file in archive fails" fail "
  rm dist/dida_${version}_windows_amd64.zip
  make_zip dist/dida_${version}_windows_amd64.zip dida_${version}_windows_amd64 dida.exe README.txt
"

run_case "unix binary without executable bit fails" fail "
  rm dist/dida_${version}_linux_amd64.tar.gz
  make_tar dist/dida_${version}_linux_amd64.tar.gz dida_${version}_linux_amd64 dida no
"

echo "verify-release-archives tests passed"
