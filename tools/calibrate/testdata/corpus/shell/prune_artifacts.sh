#!/usr/bin/env bash
set -euo pipefail

root=${1:?'usage: prune_artifacts.sh DIRECTORY [DAYS]'}
days=${2:-30}

if [[ ! -d "$root" ]]; then
  printf 'artifact directory not found: %s\n' "$root" >&2
  exit 2
fi
if [[ ! "$days" =~ ^[0-9]+$ ]] || (( days < 1 )); then
  printf 'retention days must be a positive integer\n' >&2
  exit 2
fi

root=$(cd "$root" && pwd -P)
candidate_list=$(mktemp "${TMPDIR:-/tmp}/artifacts.XXXXXX")
trap 'rm -f -- "$candidate_list"' EXIT

find "$root" -type f \
  \( -name '*.zip' -o -name '*.tar.gz' -o -name '*.checksums' \) \
  -mtime "+${days}" -print0 >"$candidate_list"

count=0
bytes=0
while IFS= read -r -d '' artifact; do
  size=$(wc -c <"$artifact")
  printf 'removing %s (%s bytes)\n' "${artifact#"$root"/}" "$size"
  rm -f -- "$artifact"
  count=$((count + 1))
  bytes=$((bytes + size))
done <"$candidate_list"

find "$root" -depth -type d -empty -mindepth 1 -delete
printf 'removed %d artifacts totaling %d bytes\n' "$count" "$bytes"
