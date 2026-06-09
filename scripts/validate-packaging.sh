#!/usr/bin/env bash
set -euo pipefail

repo="DeliciousBuding/dida-cli"
version_override=""
checksums_file=""
metadata_only=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      if [[ $# -lt 2 ]]; then
        echo "--version requires a value" >&2
        exit 1
      fi
      version_override="${2#v}"
      shift 2
      ;;
    --checksums-file)
      if [[ $# -lt 2 ]]; then
        echo "--checksums-file requires a path" >&2
        exit 1
      fi
      checksums_file="$2"
      shift 2
      ;;
    --metadata-only)
      metadata_only=true
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

homebrew_version="$(sed -n 's/.*version "\([^"]*\)".*/\1/p' packaging/homebrew/dida.rb | head -n 1)"
scoop_version="$(node -e 'console.log(require("./packaging/scoop/dida.json").version)')"
readme_version="$(sed -n 's/Current source release: `v\([^`]*\)`.*/\1/p' packaging/README.md | head -n 1)"

version="${version_override:-$homebrew_version}"
if [[ -z "$homebrew_version" || "$homebrew_version" != "$scoop_version" || "$homebrew_version" != "$readme_version" || "$homebrew_version" != "$version" ]]; then
  echo "packaging versions differ: expected=${version:-missing} homebrew=${homebrew_version:-missing} scoop=${scoop_version:-missing} readme=${readme_version:-missing}" >&2
  exit 1
fi

tag="v$version"
checksums_url="https://github.com/${repo}/releases/download/${tag}/checksums.txt"
checksums=""
checksums_source="$checksums_url"
if [[ "$metadata_only" != true ]]; then
  if [[ -n "$checksums_file" ]]; then
    checksums="$(cat "$checksums_file")"
    checksums_source="$checksums_file"
  else
    checksums="$(curl -fsSL "$checksums_url")"
  fi
fi

expected_hash() {
  local asset="$1"
  awk -v asset="$asset" '$2 == asset { print $1 }' <<<"$checksums"
}

check_asset() {
  local asset="$1"
  local got="$2"
  local want
  want="$(expected_hash "$asset")"
  if [[ -z "$want" ]]; then
    echo "missing $asset in $checksums_source" >&2
    exit 1
  fi
  if [[ "$got" != "$want" ]]; then
    echo "hash mismatch for $asset: got $got want $want" >&2
    exit 1
  fi
}

mapfile -t hb_urls < <(sed -n 's/.*url "\([^"]*\)".*/\1/p' packaging/homebrew/dida.rb)
mapfile -t hb_hashes < <(sed -n 's/.*sha256 "\([^"]*\)".*/\1/p' packaging/homebrew/dida.rb)
if [[ "${#hb_urls[@]}" -ne 4 || "${#hb_hashes[@]}" -ne 4 ]]; then
  echo "expected 4 Homebrew URLs and hashes" >&2
  exit 1
fi
for i in "${!hb_urls[@]}"; do
  url="${hb_urls[$i]}"
  asset="${url##*/}"
  [[ "$url" == "https://github.com/${repo}/releases/download/${tag}/"* ]] || {
    echo "Homebrew URL does not use $tag: $url" >&2
    exit 1
  }
  if [[ "$metadata_only" == true ]]; then
    [[ "${hb_hashes[$i]}" =~ ^[a-f0-9]{64}$ ]] || {
      echo "Homebrew hash for $asset is not a sha256 hex string" >&2
      exit 1
    }
  else
    check_asset "$asset" "${hb_hashes[$i]}"
  fi
done

scoop_assets="$(mktemp)"
trap 'rm -f "$scoop_assets"' EXIT
node - <<'NODE' > "$scoop_assets"
const pkg = require("./packaging/scoop/dida.json");
for (const [arch, info] of Object.entries(pkg.architecture)) {
  console.log([arch, info.url, info.hash, info.extract_dir].join("\t"));
}
NODE

while IFS=$'\t' read -r arch url hash extract_dir; do
  asset="${url##*/}"
  [[ "$url" == "https://github.com/${repo}/releases/download/${tag}/"* ]] || {
    echo "Scoop URL does not use $tag: $url" >&2
    exit 1
  }
  [[ "$extract_dir" == "${asset%.zip}" ]] || {
    echo "Scoop extract_dir mismatch for $arch: got $extract_dir want ${asset%.zip}" >&2
    exit 1
  }
  if [[ "$metadata_only" == true ]]; then
    [[ "$hash" =~ ^[a-f0-9]{64}$ ]] || {
      echo "Scoop hash for $asset is not a sha256 hex string" >&2
      exit 1
    }
  else
    check_asset "$asset" "$hash"
  fi
done < "$scoop_assets"
rm -f "$scoop_assets"
trap - EXIT

grep -q "releases/download/${tag}/dida_${tag}_windows_amd64.zip" packaging/winget/README.md || {
  echo "winget README does not reference $tag windows amd64 asset" >&2
  exit 1
}

if [[ "$metadata_only" == true ]]; then
  echo "packaging metadata matches $tag"
else
  echo "packaging templates match $tag"
fi
