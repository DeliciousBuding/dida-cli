#!/usr/bin/env sh
set -eu

repo="${DIDA_REPO:-DeliciousBuding/dida-cli}"
version="${DIDA_VERSION:-}"
install_dir="${DIDA_INSTALL_DIR:-$HOME/.local/bin}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: required command not found: $1" >&2
    exit 1
  }
}

download() {
  url="$1"
  out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$out"
  else
    echo "error: curl or wget is required" >&2
    exit 1
  fi
}

detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) echo "linux" ;;
    darwin) echo "darwin" ;;
    msys*|mingw*|cygwin*) echo "windows" ;;
    *) echo "error: unsupported OS: $os" >&2; exit 1 ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "error: unsupported arch: $arch" >&2; exit 1 ;;
  esac
}

os="$(detect_os)"
arch="$(detect_arch)"

archive_ext="tar.gz"
binary_name="dida"
if [ "$os" = "windows" ]; then
  archive_ext="zip"
  binary_name="dida.exe"
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM
if [ -z "$version" ]; then
  base_url="https://github.com/$repo/releases/latest/download"
  download "$base_url/checksums.txt" "$tmp_dir/checksums.txt"
  asset="$(grep "  dida_v.*_${os}_${arch}\\.${archive_ext}\$" "$tmp_dir/checksums.txt" | awk '{print $2}' | head -n 1)"
  version="$(printf '%s\n' "$asset" | sed -n "s/^dida_\\(v[^_]*\\)_${os}_${arch}\\.${archive_ext}\$/\\1/p")"
else
  base_url="https://github.com/$repo/releases/download/$version"
  download "$base_url/checksums.txt" "$tmp_dir/checksums.txt"
  asset="dida_${version}_${os}_${arch}.${archive_ext}"
fi
if [ -z "$version" ] || [ -z "$asset" ]; then
  echo "error: could not resolve latest release asset for $os/$arch" >&2
  exit 1
fi

echo "Installing DidaCLI $version for $os/$arch"
download "$base_url/$asset" "$tmp_dir/$asset"

expected="$(grep "  $asset\$" "$tmp_dir/checksums.txt" | awk '{print $1}')"
if [ -z "$expected" ]; then
  echo "error: checksum not found for $asset" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$tmp_dir/$asset" | awk '{print $1}')"
else
  need_cmd shasum
  actual="$(shasum -a 256 "$tmp_dir/$asset" | awk '{print $1}')"
fi
if [ "$actual" != "$expected" ]; then
  echo "error: checksum mismatch for $asset" >&2
  exit 1
fi

mkdir -p "$install_dir"
if [ "$archive_ext" = "zip" ]; then
  need_cmd unzip
  unzip -q "$tmp_dir/$asset" -d "$tmp_dir/out"
else
  tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"
  mv "$tmp_dir/dida_${version}_${os}_${arch}" "$tmp_dir/out"
fi

cp "$tmp_dir/out/dida_${version}_${os}_${arch}/$binary_name" "$install_dir/$binary_name" 2>/dev/null || cp "$tmp_dir/out/$binary_name" "$install_dir/$binary_name"
chmod +x "$install_dir/$binary_name" 2>/dev/null || true

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "PATH note: add $install_dir to PATH to run dida from any shell." ;;
esac

"$install_dir/$binary_name" version
"$install_dir/$binary_name" doctor --json
