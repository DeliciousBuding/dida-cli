#!/usr/bin/env bash
set -euo pipefail

tag="${GITHUB_REF_NAME:-}"
repo="${GITHUB_REPOSITORY:-DeliciousBuding/dida-cli}"
changelog="CHANGELOG.md"
output="release-notes.md"
allow_changelog_fallback="${ALLOW_CHANGELOG_FALLBACK:-false}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      if [[ $# -lt 2 ]]; then
        echo "--tag requires a value" >&2
        exit 1
      fi
      tag="$2"
      shift 2
      ;;
    --repo)
      if [[ $# -lt 2 ]]; then
        echo "--repo requires a value" >&2
        exit 1
      fi
      repo="$2"
      shift 2
      ;;
    --changelog)
      if [[ $# -lt 2 ]]; then
        echo "--changelog requires a path" >&2
        exit 1
      fi
      changelog="$2"
      shift 2
      ;;
    --output)
      if [[ $# -lt 2 ]]; then
        echo "--output requires a path" >&2
        exit 1
      fi
      output="$2"
      shift 2
      ;;
    --allow-changelog-fallback)
      allow_changelog_fallback=true
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$tag" ]]; then
  echo "release tag is required" >&2
  exit 1
fi

base_url="https://github.com/${repo}/releases/download/${tag}"
printf '%s\n\n' "## What's Changed" > "$output"

changelog_section=""
if [[ -f "$changelog" ]]; then
  changelog_section="$(awk -v ver="$tag" '
    /^## \[/ {
      if (found) exit
      if (index($0, "[" ver "]")) { found=1; next }
    }
    found { print }
  ' "$changelog")"
fi

if [[ -n "$changelog_section" ]]; then
  printf '%s\n' "$changelog_section" >> "$output"
elif [[ "$allow_changelog_fallback" == "true" ]]; then
  prev_tag="$(git tag --sort=-v:refname 2>/dev/null | grep -A1 "^${tag}$" | tail -1 || true)"
  if [[ -n "$prev_tag" && "$prev_tag" != "$tag" ]]; then
    printf '%s\n\n' "Changes since ${prev_tag}:" >> "$output"
    git log --oneline --no-merges "${prev_tag}..HEAD" 2>/dev/null | sed 's/^/- /' >> "$output" || true
  else
    printf '%s\n\n' "Changes since the previous release:" >> "$output"
    git log --oneline -20 2>/dev/null | sed 's/^/- /' >> "$output" || true
  fi
else
  echo "CHANGELOG.md must contain a ## [$tag] section before release." >&2
  echo "Manual dispatch can set allow_changelog_fallback=true for an emergency release." >&2
  exit 1
fi

{
  printf '%s\n' '' '---' '' '## Downloads' ''
  printf '%s\n' '| Platform | File | Install |'
  printf '%s\n' '|---|---|---|'
  printf '%s\n' "| **macOS** (Apple Silicon) | [dida_${tag}_darwin_arm64.tar.gz](${base_url}/dida_${tag}_darwin_arm64.tar.gz) | \`curl -fsSL https://raw.githubusercontent.com/${repo}/main/install.sh \| sh\` |"
  printf '%s\n' "| **macOS** (Intel) | [dida_${tag}_darwin_amd64.tar.gz](${base_url}/dida_${tag}_darwin_amd64.tar.gz) | \`curl -fsSL https://raw.githubusercontent.com/${repo}/main/install.sh \| sh\` |"
  printf '%s\n' "| **Linux** (x86_64) | [dida_${tag}_linux_amd64.tar.gz](${base_url}/dida_${tag}_linux_amd64.tar.gz) | \`curl -fsSL https://raw.githubusercontent.com/${repo}/main/install.sh \| sh\` |"
  printf '%s\n' "| **Linux** (ARM64) | [dida_${tag}_linux_arm64.tar.gz](${base_url}/dida_${tag}_linux_arm64.tar.gz) | \`curl -fsSL https://raw.githubusercontent.com/${repo}/main/install.sh \| sh\` |"
  printf '%s\n' "| **Windows** (x86_64) | [dida_${tag}_windows_amd64.zip](${base_url}/dida_${tag}_windows_amd64.zip) | PowerShell / Scoop |"
  printf '%s\n' "| **Windows** (ARM64) | [dida_${tag}_windows_arm64.zip](${base_url}/dida_${tag}_windows_arm64.zip) | PowerShell / Scoop |"
  printf '%s\n' '' '### Quick Install' '' '**npm:**' '```bash' 'npm install -g @delicious233/dida-cli' '```'
  printf '%s\n' '' '**macOS / Linux:**' '```bash' "curl -fsSL https://raw.githubusercontent.com/${repo}/main/install.sh | sh" '```'
  printf '%s\n' '' '**Windows PowerShell:**' '```powershell' "iwr https://raw.githubusercontent.com/${repo}/main/install.ps1 -UseB | iex" '```'
  printf '%s\n' '' '**Go:**' '```bash' "go install github.com/${repo}/cmd/dida@latest" '```'
  printf '%s\n' '' '**Self-update check:**' '```bash' 'dida upgrade --check' '```'
  printf '%s\n' '' '### Verify' '' '```bash' 'dida version' 'dida doctor --json' '```'
  printf '%s\n' '' "SHA-256 checksums are in [\`checksums.txt\`](${base_url}/checksums.txt)."
} >> "$output"

echo "release notes written to $output"
