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
Wipe demo data and reload the ACME seed without restarting services.

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

function loadDemoGraph() {
  log "loading demo graph"
  compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/bootstrap.cypher"
  compose exec -T neo4j cypher-shell -u "neo4j" -p "neo4jsecret" < "${QUERIES_DIR}/demo.cypher"
}

function resetPostgres() {
  log "resetting postgres data (keeping oauth2_clients)"
  compose exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo < "${QUERIES_DIR}/bootstrap.sql"
  compose exec -T postgres psql postgres://elemo:pgsecret@postgres/elemo -v ON_ERROR_STOP=1 <<'SQL'
TRUNCATE TABLE user_tokens, notifications RESTART IDENTITY CASCADE;
TRUNCATE TABLE oauth2_tokens RESTART IDENTITY CASCADE;
SQL
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

wipeNeo4j
loadDemoGraph
resetPostgres
flushRedis
emptyS3
reindexSearch

success "demo data reset; OAuth2 clients were left in place"
