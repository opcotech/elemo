#!/usr/bin/env bash

set -euo pipefail

if [ "${CI:-}" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
# shellcheck source=scripts/common.sh
source "${ROOT_DIR}/scripts/common.sh"

readonly ELEMO_KEEP_BACKEND="${ELEMO_KEEP_BACKEND:-false}"
readonly ELEMO_SKIP_IMAGE_BUILD="${ELEMO_SKIP_IMAGE_BUILD:-false}"
readonly PLAYWRIGHT_BROWSERS="${PLAYWRIGHT_BROWSERS:-}"

function validateBoolean() {
  local name="${1}"
  local value="${2}"

  case "${value}" in
    true|false) ;;
    *) error "${name} must be either true or false" ;;
  esac
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
  add_client_out="$(docker compose \
    -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T elemo-server bin/elemo auth add-client \
        --callback-url "${webapp_host}/api/auth/callback/elemo" --public 2>&1 | grep "client-id")"

  backupCopyFile "${WEB_DIR}/.env" "${WEB_DIR}/.env.example"
  backupCopyFile "${WEB_DIR}/.env.test.local" "${WEB_DIR}/.env.test.example"

  local secrets
  secrets="$(echo "${add_client_out}" | jq -r --arg api_host "$api_host" "\"VITE_API_BASE_URL=\" + \$api_host + \"\n\" + \"VITE_AUTH_CLIENT_ID=\" + .\"client-id\" + \"\n\" + \"VITE_AUTH_CLIENT_SECRET=\" + .\"client-secret\"")"
  echo "$secrets" >> "${WEB_DIR}/.env"
  echo "$secrets" >> "${WEB_DIR}/.env.test.local"
}

function setupDemoData() {
  log "loading demo data"
  echo "MATCH (n) DETACH DELETE n" | docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret"
  docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/bootstrap.cypher"
  docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/demo.cypher"
  docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo < "${QUERIES_DIR}/bootstrap.sql"
}

function installFrontEnd() {
  log "installing front-end requirements"
  if ! type "pnpm" > /dev/null 2>&1; then
    npm install -g pnpm
  fi
  pnpm --dir web install --unsafe-perm
  pnpm --dir web generate

  if [ -n "${PLAYWRIGHT_BROWSERS}" ]; then
    local -a browsers
    read -r -a browsers <<< "${PLAYWRIGHT_BROWSERS}"
    pnpm --dir web exec playwright install --with-deps "${browsers[@]}"
  else
    pnpm --dir web exec playwright install --with-deps
  fi

  pnpm --dir web build
}

# Run preflight
validateBoolean "ELEMO_KEEP_BACKEND" "${ELEMO_KEEP_BACKEND}"
validateBoolean "ELEMO_SKIP_IMAGE_BUILD" "${ELEMO_SKIP_IMAGE_BUILD}"
checkInstalled "docker"
checkInstalled "jq"
checkInstalled "npm"

# Generate dev config if missing
generateConfigIfMissing

# Start services
compose_args=(-f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" up --remove-orphans -d)
if [ "${ELEMO_SKIP_IMAGE_BUILD}" == "false" ]; then
  compose_args+=(--build)
fi
docker compose "${compose_args[@]}"
waitAndPrint 5

# Create a new OAuth2 client and configure the front-end
setupOAuthClient
setupDemoData

# Tear down services
if [ "${ELEMO_KEEP_BACKEND}" == "false" ]; then
  docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" down
fi

# Setup the front-end
installFrontEnd

success "the setup finished successfully, now you can run \"make dev\" or \"make start\""
