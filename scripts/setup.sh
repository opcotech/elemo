#!/usr/bin/env bash

set -euo pipefail

if [ "$CI" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

function setupOAuthClient() {
  log "setting up OAuth2 client"
  ADD_CLIENT_OUT=$(docker compose \
    -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T elemo-server bin/elemo auth add-client \
        --callback-url http://127.0.0.1:3000/api/auth/callback/elemo --public 2>&1 | grep "client-id")

  backupCopyFile "${WEB_DIR}/.env" "${WEB_DIR}/.env.example"
  backupCopyFile "${WEB_DIR}/.env.test.local" "${WEB_DIR}/.env.test.example"

  SECRETS=$(echo "${ADD_CLIENT_OUT}" | jq -r '"ELEMO_CLIENT_ID=" + ."client-id", "ELEMO_CLIENT_SECRET=" + ."client-secret"')
  echo "$SECRETS" >> "${WEB_DIR}/.env"
  echo "$SECRETS" >> "${WEB_DIR}/.env.test.local"
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
  if ! type "pnpm" 2>&1 > /dev/null; then 
    npm install -g pnpm
  fi
  pnpm install --prefix web --unsafe-perm
}

# Run preflight
checkInstalled "docker"
checkInstalled "jq"
checkInstalled "npm"

# Generate dev config if missing
generateConfigIfMissing

# Start services
docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" up --remove-orphans -d
waitAndPrint 5

# Create a new OAuth2 client and configure the front-end
setupOAuthClient
setupDemoData

# Tear down services
docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" down

# Setup the front-end
installFrontEnd

success "the setup finished successfully, now you can run \"make dev\" or \"make start\""
