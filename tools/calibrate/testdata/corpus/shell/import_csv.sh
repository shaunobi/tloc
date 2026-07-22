#!/usr/bin/env bash
set -euo pipefail

input=${1:?'usage: import_csv.sh INPUT.csv TABLE'}
table=${2:?'usage: import_csv.sh INPUT.csv TABLE'}
database_url=${DATABASE_URL:?'DATABASE_URL is required'}

if [[ ! -r "$input" ]]; then
  printf 'cannot read input file: %s\n' "$input" >&2
  exit 2
fi
if [[ ! "$table" =~ ^[a-z_][a-z0-9_]*$ ]]; then
  printf 'unsafe table name: %s\n' "$table" >&2
  exit 2
fi

temporary=$(mktemp "${TMPDIR:-/tmp}/import.XXXXXX.csv")
cleanup() {
  rm -f -- "$temporary"
}
trap cleanup EXIT

# Normalize line endings, discard blank rows, and preserve the header.
sed 's/\r$//' "$input" \
  | awk 'NR == 1 || NF > 0' \
  >"$temporary"

rows=$(( $(wc -l <"$temporary") - 1 ))
if (( rows < 1 )); then
  printf 'no data rows found in %s\n' "$input" >&2
  exit 3
fi

printf 'importing %d rows into %s\n' "$rows" "$table"
psql "$database_url" \
  --set=ON_ERROR_STOP=1 \
  --command="\\copy ${table} FROM '${temporary}' WITH (FORMAT csv, HEADER true)"
printf 'import completed successfully\n'
