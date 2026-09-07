#!/usr/bin/env bash

set -euo pipefail

# Start Compose services, (re)initialize demo data, and mint local OAuth credentials.
# This is destructive: it wipes Neo4j graph data and reloads the ACME demo seed.

if [ "${CI:-}" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
# shellcheck source=common.sh
source "${ROOT_DIR}/scripts/common.sh"

readonly ELEMO_SKIP_IMAGE_BUILD="${ELEMO_SKIP_IMAGE_BUILD:-false}"

function usage() {
  cat <<EOF
Initialize or reset the local demo environment.

This starts Compose services, wipes graph data, loads the ACME demo seed,
mints OAuth credentials into web/.env and web/.env.test.local, and leaves
the stack running.

Usage:
  $(basename "$0") --yes

Environment:
  ELEMO_SKIP_IMAGE_BUILD=true  Use an existing local elemo-server image.
EOF
}

function compose() {
  dockerCompose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" "$@"
}

function setupOAuthClient() {
  log "setting up OAuth2 client"

  local api_host
  local webapp_host

  if [ -n "${CODESPACE_NAME:-}" ]; then
    api_host="https://${CODESPACE_NAME}-35478.app.github.dev"
    webapp_host="https://${CODESPACE_NAME}-3000.app.github.dev"
  else
    api_host="http://127.0.0.1:35478"
    webapp_host="http://127.0.0.1:3000"
  fi

  local add_client_out
  add_client_out="$(compose exec -T elemo-server bin/elemo auth add-client \
        --callback-url "${webapp_host}/api/auth/callback/elemo" 2>&1 | grep "client-id")"

  backupCopyFile "${WEB_DIR}/.env" "${WEB_DIR}/.env.example"
  backupCopyFile "${WEB_DIR}/.env.test.local" "${WEB_DIR}/.env.test.example"

  local secrets
  secrets="$(echo "$add_client_out" | jq -r --arg api_host "$api_host" "\"API_BASE_URL=\" + \$api_host + \"\n\" + \"AUTH_CLIENT_ID=\" + .\"client-id\" + \"\n\" + \"AUTH_CLIENT_SECRET=\" + .\"client-secret\"")"
  echo "$secrets" >> "${WEB_DIR}/.env"
  echo "$secrets" >> "${WEB_DIR}/.env.test.local"
}

function setupDemoData() {
  log "loading demo data"
  echo "MATCH (n) DETACH DELETE n" | compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret"
  compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/bootstrap.cypher"
  compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/demo.cypher"
  compose exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo < "${QUERIES_DIR}/bootstrap.sql"
}

function reindexSearchIndex() {
  log "reindexing search"
  compose exec -T elemo-server bin/elemo search reindex --delete-all
}

confirmed=false
while [ "${#}" -gt 0 ]; do
  case "${1}" in
    --yes|-y)
      confirmed=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      error "unknown argument: ${1}"
      ;;
  esac
  shift
done

if [ "${confirmed}" != "true" ]; then
  usage
  error "refusing to initialize demo data without --yes"
fi

validateBoolean() {
  local name="${1}"
  local value="${2}"
  case "${value}" in
    true|false) ;;
    *) error "${name} must be either true or false" ;;
  esac
}

validateBoolean "ELEMO_SKIP_IMAGE_BUILD" "${ELEMO_SKIP_IMAGE_BUILD}"
checkInstalled "docker"
checkInstalled "jq"

generateConfigIfMissing

compose up --remove-orphans -d --wait postgres neo4j
setupDemoData

compose_args=(up --remove-orphans -d)
if [ "${ELEMO_SKIP_IMAGE_BUILD}" == "false" ]; then
  compose_args+=(--build)
fi
compose "${compose_args[@]}"
waitAndPrint 5

setupOAuthClient
reindexSearchIndex

success "demo environment is ready; run \"mise run dev\""
