#!/usr/bin/env bash
set -euo pipefail

fail=0

check_path() {
  local path="$1"
  case "$path" in
    .env|.env.*|*/.env|*/.env.*|\
    .dida-cli/*|data/private/*|tmp/*|.cache/*|.tmp/*|\
    cookie.json|*/cookie.json|official-mcp-token.json|*/official-mcp-token.json|\
    openapi-oauth.json|*/openapi-oauth.json|openapi-client.json|*/openapi-client.json|\
    identity.json|*/identity.json|\
    *.out|*.exe|*.dll|*.so|*.dylib|*.log|*.tmp|*.test|\
    *.prof|*.coverprofile|*.tgz|*.zip|*.tar|*.tar.gz|\
    *.har|*.pcap|*.pcapng|*.jsonl|*.ndjson|*.dump|*.sqlite|*.db|*.bak|\
    release-notes.md|checksums.txt|node_modules/*|npm/node_modules/*|\
    npm/bin/dida-bin|npm/bin/dida.exe)
      printf 'private-state path is tracked: %s\n' "$path" >&2
      fail=1
      ;;
  esac
}

while IFS= read -r -d '' path; do
  check_path "$path"
done < <(git ls-files -z --cached --others --exclude-standard)

secret_patterns=(
  'Dida365 API token|dp_[A-Za-z0-9_-]{24,}'
  'OAuth refresh token|refresh_token["'\'' ]*[:=]["'\'' ]*[A-Za-z0-9._~+/=-]{24,}'
  'OAuth access token|access_token["'\'' ]*[:=]["'\'' ]*[A-Za-z0-9._~+/=-]{24,}'
  'OAuth client secret|client_secret["'\'' ]*[:=]["'\'' ]*[A-Za-z0-9._~+/=-]{24,}'
  'Cookie header token|Cookie:[[:space:]]*t=[^[:space:];]{24,}'
  'Set-Cookie token|Set-Cookie:[^[:space:]]*t=[^[:space:];]{24,}'
  'Raw t cookie|(^|[^A-Za-z0-9_-])t=[A-Za-z0-9._~+/=-]{24,}'
  'Authorization bearer token|Authorization:[[:space:]]*Bearer[[:space:]]+[A-Za-z0-9._~+/=-]{24,}'
  'JWT|eyJ[A-Za-z0-9_-]{16,}\.[A-Za-z0-9_-]{16,}\.[A-Za-z0-9_-]{16,}'
  'Local user path|([A-Za-z]:[\\/](Users|Code|Projects|Temp)[\\/][A-Za-z0-9._ -][^[:space:]]*|\\\\[A-Za-z0-9._$-]+\\[A-Za-z0-9._$-]+|%USERPROFILE%[\\/][^[:space:]]+|/(Users|home)/[A-Za-z0-9._-]+)'
)

grep_paths=(
  .
  ':(exclude)scripts/check-private-state.sh'
)

for entry in "${secret_patterns[@]}"; do
  name="${entry%%|*}"
  pattern="${entry#*|}"
  if git grep -I -n -E "$pattern" -- . \
    ':(exclude)scripts/check-private-state.sh'; then
    printf 'possible secret matched pattern (%s): %s\n' "$name" "$pattern" >&2
    fail=1
  fi
  if git grep --untracked -I -n -E "$pattern" -- "${grep_paths[@]}"; then
    printf 'possible secret matched pattern (%s): %s\n' "$name" "$pattern" >&2
    fail=1
  fi
done

exit "$fail"
