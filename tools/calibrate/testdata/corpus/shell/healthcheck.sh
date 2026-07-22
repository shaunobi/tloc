#!/usr/bin/env bash
set -euo pipefail

base_url=${1:-http://127.0.0.1:8080}
attempts=${ATTEMPTS:-8}
delay=${DELAY_SECONDS:-2}

request_health() {
  curl --silent --show-error \
    --connect-timeout 2 \
    --max-time 5 \
    --write-out $'\n%{http_code}' \
    "${base_url%/}/health"
}

for ((attempt = 1; attempt <= attempts; attempt++)); do
  if response=$(request_health); then
    status=${response##*$'\n'}
    body=${response%$'\n'*}
    if [[ "$status" == 200 ]] && jq -e '.status == "ok"' <<<"$body" >/dev/null; then
      version=$(jq -r '.version // "unknown"' <<<"$body")
      printf 'service is healthy (version %s)\n' "$version"
      exit 0
    fi
    printf 'attempt %d/%d returned HTTP %s\n' "$attempt" "$attempts" "$status" >&2
  else
    printf 'attempt %d/%d could not connect\n' "$attempt" "$attempts" >&2
  fi

  if (( attempt < attempts )); then
    sleep "$delay"
    delay=$((delay < 16 ? delay * 2 : 16))
  fi
done

printf 'service did not become healthy: %s\n' "$base_url" >&2
exit 1
