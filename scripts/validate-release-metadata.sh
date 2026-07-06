#!/usr/bin/env bash
set -euo pipefail

tag="${GITHUB_REF_NAME:-}"
allow_changelog_fallback="${ALLOW_CHANGELOG_FALLBACK:-false}"
skip_git_checks=false

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
    --allow-changelog-fallback)
      allow_changelog_fallback=true
      shift
      ;;
    --skip-git-checks)
      skip_git_checks=true
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

if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "release tag must use semver format vX.Y.Z; got $tag" >&2
  exit 1
fi

version="${tag#v}"
package_version="$(node -e 'console.log(require("./npm/package.json").version)')"
if [[ "$package_version" != "$version" ]]; then
  echo "npm/package.json version $package_version does not match $tag" >&2
  exit 1
fi

if [[ "$allow_changelog_fallback" != "true" ]] && ! grep -q "^## \\[$tag\\]" CHANGELOG.md; then
  echo "CHANGELOG.md must contain ## [$tag] before release" >&2
  exit 1
fi

if [[ "$skip_git_checks" != "true" ]]; then
  if [[ "${GITHUB_REF_TYPE:-tag}" != "tag" ]]; then
    echo "release must run from a tag ref, got ${GITHUB_REF_TYPE:-unknown}:$tag" >&2
    exit 1
  fi

  git rev-parse --verify --quiet "refs/tags/$tag" >/dev/null
  tag_commit="$(git rev-list -n 1 "$tag")"
  head_commit="$(git rev-parse HEAD)"
  if [[ "$head_commit" != "$tag_commit" ]]; then
    echo "checked-out HEAD $head_commit does not match tag $tag at $tag_commit" >&2
    exit 1
  fi

  if git rev-parse --verify --quiet origin/main >/dev/null; then
    if ! git merge-base --is-ancestor "$tag_commit" origin/main; then
      echo "release tag $tag must point to a commit reachable from origin/main" >&2
      exit 1
    fi
  fi
fi

echo "release metadata valid for $tag"
