#!/usr/bin/env bash
set -euo pipefail

repo="DeliciousBuding/dida-cli"
version=""
checksums_file=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo)
      if [[ $# -lt 2 ]]; then
        echo "--repo requires owner/name" >&2
        exit 1
      fi
      repo="$2"
      shift 2
      ;;
    --version)
      if [[ $# -lt 2 ]]; then
        echo "--version requires vX.Y.Z or X.Y.Z" >&2
        exit 1
      fi
      version="${2#v}"
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
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$version" ]]; then
  echo "--version is required" >&2
  exit 1
fi

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "version must be X.Y.Z or vX.Y.Z; got $version" >&2
  exit 1
fi

if [[ ! "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "repo must use owner/name format; got $repo" >&2
  exit 1
fi

tag="v$version"
checksums_url="https://github.com/${repo}/releases/download/${tag}/checksums.txt"
if [[ -n "$checksums_file" ]]; then
  if [[ ! -f "$checksums_file" ]]; then
    echo "checksums file is missing: $checksums_file" >&2
    exit 1
  fi
  checksums="$(cat "$checksums_file")"
else
  checksums="$(curl -fsSL "$checksums_url")"
fi

hash_for() {
  local asset="$1"
  local hash
  hash="$(awk -v asset="$asset" '$2 == asset { print $1 }' <<<"$checksums")"
  if [[ ! "$hash" =~ ^[a-f0-9]{64}$ ]]; then
    echo "missing sha256 for $asset in checksums.txt" >&2
    exit 1
  fi
  printf '%s' "$hash"
}

asset_windows_amd64="dida_${tag}_windows_amd64.zip"
asset_windows_arm64="dida_${tag}_windows_arm64.zip"
asset_linux_amd64="dida_${tag}_linux_amd64.tar.gz"
asset_linux_arm64="dida_${tag}_linux_arm64.tar.gz"
asset_darwin_amd64="dida_${tag}_darwin_amd64.tar.gz"
asset_darwin_arm64="dida_${tag}_darwin_arm64.tar.gz"

hash_windows_amd64="$(hash_for "$asset_windows_amd64")"
hash_windows_arm64="$(hash_for "$asset_windows_arm64")"
hash_linux_amd64="$(hash_for "$asset_linux_amd64")"
hash_linux_arm64="$(hash_for "$asset_linux_arm64")"
hash_darwin_amd64="$(hash_for "$asset_darwin_amd64")"
hash_darwin_arm64="$(hash_for "$asset_darwin_arm64")"

base_url="https://github.com/${repo}/releases/download/${tag}"
mkdir -p packaging/homebrew packaging/scoop packaging/winget

cat > packaging/homebrew/dida.rb <<EOF
class Dida < Formula
  desc "JSON-first CLI for Dida365 and TickTick"
  homepage "https://github.com/${repo}"
  version "${version}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "${base_url}/${asset_darwin_arm64}"
      sha256 "${hash_darwin_arm64}"
    else
      url "${base_url}/${asset_darwin_amd64}"
      sha256 "${hash_darwin_amd64}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "${base_url}/${asset_linux_arm64}"
      sha256 "${hash_linux_arm64}"
    else
      url "${base_url}/${asset_linux_amd64}"
      sha256 "${hash_linux_amd64}"
    end
  end

  def install
    binary = Dir["**/dida"].find { |path| File.file?(path) }
    odie "dida binary not found in release archive" unless binary

    bin.install binary => "dida"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dida version")
    assert_match "\\"ok\\": true", shell_output("#{bin}/dida doctor --json")
  end
end
EOF

node - "$repo" "$version" "$hash_windows_amd64" "$hash_windows_arm64" <<'NODE'
const fs = require("fs");

const [repo, version, hash64, hashArm64] = process.argv.slice(2);
const tag = `v${version}`;
const base = `https://github.com/${repo}/releases/download/${tag}`;
const manifest = {
  version,
  description: "JSON-first CLI for Dida365 and TickTick.",
  homepage: `https://github.com/${repo}`,
  license: "MIT",
  architecture: {
    "64bit": {
      url: `${base}/dida_${tag}_windows_amd64.zip`,
      hash: hash64,
      extract_dir: `dida_${tag}_windows_amd64`,
    },
    arm64: {
      url: `${base}/dida_${tag}_windows_arm64.zip`,
      hash: hashArm64,
      extract_dir: `dida_${tag}_windows_arm64`,
    },
  },
  bin: "dida.exe",
  checkver: {
    github: `https://github.com/${repo}`,
  },
  autoupdate: {
    architecture: {
      "64bit": {
        url: `https://github.com/${repo}/releases/download/v$version/dida_v$version_windows_amd64.zip`,
        extract_dir: "dida_v$version_windows_amd64",
      },
      arm64: {
        url: `https://github.com/${repo}/releases/download/v$version/dida_v$version_windows_arm64.zip`,
        extract_dir: "dida_v$version_windows_arm64",
      },
    },
    hash: {
      url: "$baseurl/checksums.txt",
    },
  },
};

fs.writeFileSync("packaging/scoop/dida.json", `${JSON.stringify(manifest, null, 2)}\n`);
NODE

cat > packaging/README.md <<EOF
# Packaging Templates

This directory contains maintainer-facing packaging templates for distribution
channels that usually live in separate repositories or registries.

Current source release: \`${tag}\`

## Channels

| Channel | File | Status |
| --- | --- | --- |
| Homebrew | \`homebrew/dida.rb\` | Template with macOS and Linux checksums |
| Scoop | \`scoop/dida.json\` | Template with Windows amd64 and arm64 checksums |
| winget | \`winget/README.md\` | Submission notes; manifest intentionally not generated yet |

## Update Rules

1. Publish a GitHub Release tag and confirm all archives plus \`checksums.txt\`
   are attached.
2. Run:

   \`\`\`bash
   bash scripts/update-packaging-templates.sh --version ${tag}
   \`\`\`

   Use \`--checksums-file <path>\` when preparing from a downloaded or staged
   checksum file.
3. Run:

   \`\`\`bash
   bash scripts/validate-packaging.sh --version ${tag} --checksums-file <path>
   \`\`\`

   If you rely on the published GitHub Release checksum asset, omit
   \`--checksums-file\`.
4. Test the template in the target package manager before publishing it to an
   external tap, bucket, or manifest repository.
5. Do not add credentials, local paths, private test accounts, or release
   automation secrets here.
EOF

cat > packaging/winget/README.md <<EOF
# winget Packaging Notes

Do not submit a winget manifest until DidaCLI has a stable release cadence and
the package identifier is final.

Recommended identifier once ready:

\`\`\`text
DeliciousBuding.DidaCLI
\`\`\`

Recommended installer source:

\`\`\`text
https://github.com/${repo}/releases
\`\`\`

winget usually works best with explicit installer manifests generated by
\`wingetcreate\` from the Windows release archives:

\`\`\`powershell
wingetcreate new ${base_url}/${asset_windows_amd64}
winget validate --manifest <manifest-directory>
\`\`\`

Before submitting:

1. Verify the Windows archive installs a single \`dida.exe\` binary.
2. Confirm the package identifier, publisher, package name, and license URL.
3. Test the generated manifest locally.
4. Keep local test paths, account data, tokens, and any private notes out of the
   manifest repository.

Current validation boundary:

- Generate and validate the manifest with \`wingetcreate\` on a Windows
  packaging host before submission.
- Run \`scripts/winget-submission-preflight.sh\` before using \`wingetcreate\`
  so the release URL, package id, and submission boundary are checked.
- Keep Homebrew and Scoop native install smoke results in their package-manager
  review notes.
EOF

echo "packaging templates updated for ${tag}"
