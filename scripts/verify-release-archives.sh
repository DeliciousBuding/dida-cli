#!/usr/bin/env bash
set -euo pipefail

version="${1:-${GITHUB_REF_NAME:-}}"
dist_dir="${2:-dist}"

if [[ -z "$version" ]]; then
  echo "usage: scripts/verify-release-archives.sh <version> [dist-dir]" >&2
  exit 1
fi

version="${version#v}"
tag="v${version}"
python_bin="$(command -v python3 || command -v python || true)"

list_zip_entries() {
  local archive="$1"
  if [[ -n "$python_bin" ]]; then
    "$python_bin" -c 'import sys, zipfile; print("\n".join(zipfile.ZipFile(sys.argv[1]).namelist()))' "$archive"
  else
    unzip -Z1 "$archive"
  fi
}

extract_zip() {
  local archive="$1"
  local target="$2"
  if [[ -n "$python_bin" ]]; then
    "$python_bin" -c 'import sys, zipfile; zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])' "$archive" "$target"
  else
    unzip -q "$archive" -d "$target"
  fi
}

expected=(
  "dida_${tag}_windows_amd64.zip"
  "dida_${tag}_windows_arm64.zip"
  "dida_${tag}_linux_amd64.tar.gz"
  "dida_${tag}_linux_arm64.tar.gz"
  "dida_${tag}_darwin_amd64.tar.gz"
  "dida_${tag}_darwin_arm64.tar.gz"
)

shopt -s nullglob
archives=("${dist_dir}"/*.zip "${dist_dir}"/*.tar.gz)
if [[ "${#archives[@]}" -ne "${#expected[@]}" ]]; then
  printf 'expected %s release archives, found %s\n' "${#expected[@]}" "${#archives[@]}" >&2
  printf '%s\n' "${archives[@]}" >&2
  exit 1
fi

tmpdirs=()
cleanup() {
  if [[ "${#tmpdirs[@]}" -gt 0 ]]; then
    rm -rf "${tmpdirs[@]}"
  fi
}
trap cleanup EXIT

for asset in "${expected[@]}"; do
  archive="${dist_dir}/${asset}"
  if [[ ! -f "$archive" ]]; then
    echo "missing expected archive: $archive" >&2
    exit 1
  fi

  folder="${asset%.zip}"
  folder="${folder%.tar.gz}"
  if [[ "$asset" == *windows* ]]; then
    binary="dida.exe"
  else
    binary="dida"
  fi

  if [[ "$asset" == *.zip ]]; then
    mapfile -t entries < <(list_zip_entries "$archive" | tr -d '\r')
  else
    mapfile -t entries < <(tar -tzf "$archive" | tr -d '\r')
  fi

  files=()
  for entry in "${entries[@]}"; do
    [[ "$entry" == */ ]] && continue
    files+=("$entry")
  done

  if [[ "${#files[@]}" -ne 1 || "${files[0]}" != "${folder}/${binary}" ]]; then
    printf 'archive %s should contain only %s/%s, got:\n' "$asset" "$folder" "$binary" >&2
    printf '%s\n' "${entries[@]}" >&2
    exit 1
  fi

  work="$(mktemp -d)"
  tmpdirs+=("$work")
  if [[ "$asset" == *.zip ]]; then
    extract_zip "$archive" "$work"
  else
    tar -xzf "$archive" -C "$work"
  fi

  if [[ "$asset" == *windows* ]]; then
    test -f "$work/$folder/$binary"
  else
    test -f "$work/$folder/$binary"
    if [[ ! -x "$work/$folder/$binary" ]]; then
      mode_line="$(tar -tvzf "$archive" "$folder/$binary" | head -n 1)"
      if [[ "$mode_line" != -rwx* ]]; then
        echo "archive $asset should preserve executable mode for $folder/$binary" >&2
        exit 1
      fi
    fi
  fi
  rm -rf "$work"
done

tmpdirs=()
trap - EXIT

echo "release archives match ${tag}"
