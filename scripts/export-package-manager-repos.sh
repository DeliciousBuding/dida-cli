#!/usr/bin/env bash
set -euo pipefail

output_dir="dist/package-manager-repos"
homebrew_repo="DeliciousBuding/homebrew-dida"
scoop_repo="DeliciousBuding/scoop-bucket"
scoop_bucket_name="dida"

usage() {
  cat <<'EOF'
usage: bash scripts/export-package-manager-repos.sh [options]

Options:
  --output <dir>          Export directory. Default: dist/package-manager-repos
  --homebrew-repo <repo>  GitHub repo for the tap, owner/name. Default: DeliciousBuding/homebrew-dida
  --scoop-repo <repo>     GitHub repo for the bucket, owner/name. Default: DeliciousBuding/scoop-bucket
  --scoop-bucket <name>   Scoop bucket alias. Default: dida
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      if [[ $# -lt 2 ]]; then
        echo "--output requires a directory" >&2
        exit 1
      fi
      output_dir="$2"
      shift 2
      ;;
    --homebrew-repo)
      if [[ $# -lt 2 ]]; then
        echo "--homebrew-repo requires owner/name" >&2
        exit 1
      fi
      homebrew_repo="$2"
      shift 2
      ;;
    --scoop-repo)
      if [[ $# -lt 2 ]]; then
        echo "--scoop-repo requires owner/name" >&2
        exit 1
      fi
      scoop_repo="$2"
      shift 2
      ;;
    --scoop-bucket)
      if [[ $# -lt 2 ]]; then
        echo "--scoop-bucket requires a name" >&2
        exit 1
      fi
      scoop_bucket_name="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

repo_name_re='^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$'
if [[ ! "$homebrew_repo" =~ $repo_name_re ]]; then
  echo "homebrew repo must use owner/name format; got $homebrew_repo" >&2
  exit 1
fi
if [[ ! "$scoop_repo" =~ $repo_name_re ]]; then
  echo "scoop repo must use owner/name format; got $scoop_repo" >&2
  exit 1
fi
if [[ ! "$scoop_bucket_name" =~ ^[A-Za-z0-9_.-]+$ ]]; then
  echo "scoop bucket name must be a simple bucket alias; got $scoop_bucket_name" >&2
  exit 1
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

homebrew_version="$(sed -n 's/.*version "\([^"]*\)".*/\1/p' packaging/homebrew/dida.rb | head -n 1)"
scoop_version="$(node -e 'console.log(require("./packaging/scoop/dida.json").version)')"
if [[ -z "$homebrew_version" || "$homebrew_version" != "$scoop_version" ]]; then
  echo "packaging versions differ: homebrew=${homebrew_version:-missing} scoop=${scoop_version:-missing}" >&2
  exit 1
fi
tag="v$homebrew_version"

case "$output_dir" in
  ""|"/"|"."|"./")
    echo "--output must point to a non-root export directory" >&2
    exit 1
    ;;
esac

homebrew_tap_owner="${homebrew_repo%%/*}"
homebrew_tap_name="${homebrew_repo#*/}"
homebrew_tap_alias="${homebrew_tap_name#homebrew-}"

homebrew_dir="$output_dir/homebrew-tap"
scoop_dir="$output_dir/scoop-bucket"

rm -rf "$homebrew_dir" "$scoop_dir"
mkdir -p "$homebrew_dir/Formula" "$scoop_dir/bucket"

cp packaging/homebrew/dida.rb "$homebrew_dir/Formula/dida.rb"
cp packaging/scoop/dida.json "$scoop_dir/bucket/dida.json"
cp LICENSE "$homebrew_dir/LICENSE"
cp LICENSE "$scoop_dir/LICENSE"

cat >"$homebrew_dir/README.md" <<EOF
# DidaCLI Homebrew Tap

This repository is the publishable Homebrew tap for DidaCLI.

Generated from \`DeliciousBuding/dida-cli\` release \`${tag}\`.

## Publish Boundary

This export is not the source of truth. The source template is
\`packaging/homebrew/dida.rb\` in the main DidaCLI repository.

Before publishing:

1. Create the GitHub repository \`${homebrew_repo}\`.
2. Push this directory as the repository root.
3. Run a native Homebrew install smoke on macOS or Linux.

## Install

After the tap repository is published:

\`\`\`bash
brew tap ${homebrew_tap_owner}/${homebrew_tap_alias}
brew install dida
dida version
dida doctor --json
\`\`\`

## Update

Regenerate this repository from the main DidaCLI repository after each release:

\`\`\`bash
bash scripts/update-packaging-templates.sh --version ${tag}
bash scripts/export-package-manager-repos.sh --homebrew-repo ${homebrew_repo}
\`\`\`
EOF

cat >"$scoop_dir/README.md" <<EOF
# DidaCLI Scoop Bucket

This repository is the publishable Scoop bucket for DidaCLI.

Generated from \`DeliciousBuding/dida-cli\` release \`${tag}\`.

## Publish Boundary

This export is not the source of truth. The source template is
\`packaging/scoop/dida.json\` in the main DidaCLI repository.

Before publishing:

1. Create the GitHub repository \`${scoop_repo}\`.
2. Push this directory as the repository root.
3. Run a native Scoop install smoke on Windows.

## Install

After the bucket repository is published:

\`\`\`powershell
scoop bucket add ${scoop_bucket_name} https://github.com/${scoop_repo}
scoop install dida
dida version
dida doctor --json
\`\`\`

## Update

Regenerate this repository from the main DidaCLI repository after each release:

\`\`\`bash
bash scripts/update-packaging-templates.sh --version ${tag}
bash scripts/export-package-manager-repos.sh --scoop-repo ${scoop_repo}
\`\`\`
EOF

echo "exported package-manager repos for ${tag} to ${output_dir}"
