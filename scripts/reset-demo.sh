#!/usr/bin/env bash

set -euo pipefail

if [ "${CI:-}" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

readonly COMPOSE_FILE="${DOCKER_DEPLOY_DIR}/docker-compose.yml"

function usage() {
  cat <<EOF
Wipe demo data and reload the ACME seed without restarting data stores.
The API server is restarted so the plugin registry is empty.

Usage:
  $(basename "$0") --yes

EOF
}

function compose() {
  docker compose -f "${COMPOSE_FILE}" "$@"
}

function requireYes() {
  local confirmed=false
  local arg

  for arg in "$@"; do
    case "${arg}" in
      --yes|-y)
        confirmed=true
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        usage
        error "unknown argument: ${arg}"
        ;;
    esac
  done

  if [ "${confirmed}" != "true" ]; then
    usage
    error "refusing to reset demo data without --yes"
  fi
}

function wipeNeo4j() {
  log "wiping neo4j"
  echo "MATCH (n) DETACH DELETE n" | compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret"
}

function bootstrapDatabase() {
  log "bootstrapping databases"
  compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/bootstrap.cypher"
  compose exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo < "${QUERIES_DIR}/bootstrap.sql"
  compose exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo -v ON_ERROR_STOP=1 <<'SQL'
TRUNCATE TABLE user_tokens, notifications RESTART IDENTITY CASCADE;
TRUNCATE TABLE oauth2_tokens RESTART IDENTITY CASCADE;
TRUNCATE TABLE plugin_storage, plugin_activations, plugin_installations;
SQL
}

function wipePluginInstallations() {
  log "wiping plugin installations"
  shopt -s nullglob
  local plugin_root
  for plugin_root in "${PLUGINS_DIR}"/*/; do
    # Source checkouts keep plugin.yaml at the plugin root; installs are
    # {pluginId}/{version}/plugin.yaml.
    if [ -f "${plugin_root}plugin.yaml" ]; then
      continue
    fi
    rm -rf "${plugin_root}"
  done
  shopt -u nullglob

  # Dockerized runs keep installs in a named volume instead of PLUGINS_DIR.
  compose rm --stop --force elemo-server elemo-worker >/dev/null
  docker volume rm --force elemo_plugin_data >/dev/null
  compose up --detach elemo-server elemo-worker >/dev/null
  waitAndPrint 5
}

function reloadPluginRuntime() {
  log "reloading plugin runtime"
  compose restart elemo-server
  waitAndPrint 5
}

function loadDemoData() {
  log "loading demo data"
  go run "${TOOLS_DIR}/workload-prefill" \
    -config "${CONFIG_DIR}/config.local.gen.yml" \
    -profile smoke \
    -yes
}

function flushRedis() {
  log "flushing redis"
  compose exec -T redis redis-cli FLUSHDB
}

function emptyS3() {
  log "emptying s3 bucket"
  compose exec -T localstack awslocal s3 rm s3://elemo --recursive >/dev/null
}

function reindexSearch() {
  log "reindexing search"
  compose exec -T elemo-server bin/elemo search reindex --delete-all
}

requireYes "$@"
checkInstalled "docker"
checkInstalled "go"

wipeNeo4j
emptyS3
bootstrapDatabase
wipePluginInstallations
loadDemoData
flushRedis
reindexSearch
reloadPluginRuntime

success "demo data reset"
