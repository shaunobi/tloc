#!/usr/bin/env bash
set -euo pipefail

environment=${1:?'usage: deploy.sh ENVIRONMENT IMAGE'}
image=${2:?'usage: deploy.sh ENVIRONMENT IMAGE'}
namespace="app-${environment}"

log() {
  printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*"
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    printf 'required command not found: %s\n' "$1" >&2
    return 1
  }
}

require_command kubectl
require_command docker

case "$environment" in
  staging|production) ;;
  *) printf 'unsupported environment: %s\n' "$environment" >&2; exit 2 ;;
esac

log "checking image ${image}"
docker manifest inspect "$image" >/dev/null

previous=$(kubectl -n "$namespace" get deployment api \
  -o jsonpath='{.spec.template.spec.containers[0].image}')

rollback() {
  log "deployment failed; restoring ${previous}"
  kubectl -n "$namespace" set image deployment/api "api=${previous}"
}
trap rollback ERR

kubectl -n "$namespace" set image deployment/api "api=${image}"
kubectl -n "$namespace" rollout status deployment/api --timeout=180s
trap - ERR
log "deployed ${image} to ${namespace}"
