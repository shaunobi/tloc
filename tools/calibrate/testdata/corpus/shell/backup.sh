#!/usr/bin/env bash
set -euo pipefail

source_dir=${1:?"usage: backup.sh SOURCE DESTINATION"}
destination=${2:?"usage: backup.sh SOURCE DESTINATION"}
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
archive="${destination%/}/backup-${timestamp}.tar.gz"

if [[ ! -d "$source_dir" ]]; then
  printf 'source directory does not exist: %s\n' "$source_dir" >&2
  exit 2
fi

mkdir -p "$destination"
tar \
  --exclude='.git' \
  --exclude='node_modules' \
  -czf "$archive" \
  -C "$source_dir" .

printf 'created %s (%s bytes)\n' "$archive" "$(wc -c <"$archive")"
