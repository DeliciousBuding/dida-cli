#!/usr/bin/env bash
set -euo pipefail

workflow_dir=".github/workflows"
version_comment_re='#[[:space:]]*v[0-9]'

if [[ ! -d "$workflow_dir" ]]; then
  echo "workflow directory not found: $workflow_dir" >&2
  exit 1
fi

failed=false
while IFS= read -r file; do
  while IFS= read -r line; do
    lineno="${line%%:*}"
    text="${line#*:}"
    uses="$(sed -nE 's/^[[:space:]]*-?[[:space:]]*uses:[[:space:]]*([^[:space:]#]+).*/\1/p' <<<"$text")"
    [[ -n "$uses" ]] || continue
    [[ "$uses" == ./* ]] && continue
    [[ "$uses" == docker://* ]] && continue
    ref="${uses##*@}"
    if [[ ! "$ref" =~ ^[a-f0-9]{40}$ ]]; then
      echo "$file:$lineno uses is not pinned to a full commit SHA: $uses" >&2
      failed=true
      continue
    fi
    if [[ ! "$text" =~ $version_comment_re ]]; then
      echo "$file:$lineno pinned action must keep a version comment for review/update context" >&2
      failed=true
    fi
  done < <(grep -nE '^[[:space:]]*-?[[:space:]]*uses:[[:space:]]*[^[:space:]#]+@' "$file" || true)
done < <(find "$workflow_dir" -type f \( -name '*.yml' -o -name '*.yaml' \) | sort)

if [[ "$failed" == true ]]; then
  exit 1
fi

echo "GitHub Actions dependencies are pinned to commit SHAs"
