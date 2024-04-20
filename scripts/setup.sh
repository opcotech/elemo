#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
CMD_DIR="${ROOT_DIR}/cmd/elemo"
DEV_CONFIG_DIR="${ROOT_DIR}/configs/development"
DOCKER_DEPLOY_DIR="${ROOT_DIR}/deploy/docker"
QUERIES_DIR="${ROOT_DIR}/assets/queries"
SCRIPTS_DIR="${ROOT_DIR}/scripts"
WEB_DIR="${ROOT_DIR}/web"

export ELEMO_CONFIG="${DEV_CONFIG_DIR}/config.local.gen.yml"

function checkInstalled() {
  local program="${1}"

  if ! type "${program}"; then
      echo "Couldn't find ${program} in your PATH. Make sure it is installed."
      exit 1
  fi
}

function waitAndPrint() {
  echo "waiting ${1} seconds to let the services boot"
  sleep "${1}"
}

function backupCopyFile() {
  local backupFile="${1}"
  local copyFile="${2}"

  [[ -f "${backupFile}" ]] && mv "${backupFile}" "${backupFile}.bkp"
  cp "${copyFile}" "${backupFile}"
}

function setupOAuthClient() {
  # Add Oauth2 client to DB
  go run "${CMD_DIR}/main.go" auth add-client --callback-url http://127.0.0.1:3000/api/auth/callback/elemo --public 2>&1
  ADD_CLIENT_OUT=$(go run "${CMD_DIR}/main.go" auth add-client --callback-url http://127.0.0.1:3000/api/auth/callback/elemo --public 2>&1 | grep "client-id")

  # Add generated client secrets to frontend config
  backupCopyFile "${WEB_DIR}/.env" "${WEB_DIR}/.env.example"
  backupCopyFile "${WEB_DIR}/.env.test.local" "${WEB_DIR}/.env.test.example"

  SECRETS=$(echo "${ADD_CLIENT_OUT}" | jq -r '"ELEMO_CLIENT_ID=" + ."client-id", "ELEMO_CLIENT_SECRET=" + ."client-secret"')
  echo "$SECRETS" >> "${WEB_DIR}/.env"
  echo "$SECRETS" >> "${WEB_DIR}/.env.test.local"
}

function setupDemoData() {
  # Spin up services
  docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" up neo4j --remove-orphans -d
  docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" up postgres --remove-orphans -d
  waitAndPrint 5

  # Load data into Neo4J
  echo "MATCH (n) DETACH DELETE n" | docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret"
  docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/bootstrap.cypher"
  docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/demo.cypher"

  # Load data into Postgres
  docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo < "${QUERIES_DIR}/bootstrap.sql"
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

# Spin up services
docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" up neo4j postgres --remove-orphans -d
waitAndPrint 5

# Create a new OAuth2 client and configure the front-end
setupOAuthClient
setupDemoData

# Tear down services
docker-compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" down

# Setup the front-end
installFrontEnd

success "the setup finished successfully, now you can run \"make dev\" or \"make start\""
