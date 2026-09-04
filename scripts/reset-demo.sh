#!/usr/bin/env bash

set -euo pipefail

if [ "${CI:-}" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

readonly COMPOSE_FILE="${DOCKER_DEPLOY_DIR}/docker-compose.yml"

WITH_OAUTH=false

function usage() {
  cat <<EOF
Wipe demo data and reload the smoke-profile seed without restarting data stores.
The API server is restarted so the plugin registry is empty.

Usage:
  $(basename "$0") --yes [--with-oauth]

Environment (optional; local defaults match docker-compose.yml):
  ELEMO_COMPOSE_OVERRIDE   Extra Compose file (production override)
  ELEMO_COMPOSE_ENV_FILE   Compose --env-file (vaulted passwords)
  ELEMO_PREFILL_CONFIG     Config for workload-prefill (default: \$ELEMO_CONFIG)
  ELEMO_CONFIGS_DIR        Host configs dir mounted at /src/configs in Docker
  ELEMO_PREFILL_DOCKER     Set to 1 to run workload-prefill in a Go container
  ELEMO_COMPOSE_NETWORK    Network for dockerized prefill (default: elemo-network)
  ELEMO_GOLANG_IMAGE       Image for dockerized prefill (default: golang:1.26-trixie)
  NEO4J_USER / NEO4J_PASSWORD or NEO4J_AUTH
  POSTGRES_USER / POSTGRES_PASSWORD / POSTGRES_DB

  --with-oauth also needs:
  ELEMO_OAUTH_CALLBACK_URL ELEMO_OAUTH_SECRETS_FILE ELEMO_WEB_ENV ELEMO_API_BASE_URL
  ELEMO_WEB_SERVICE        Optional systemd unit to restart after remint

EOF
}

function compose() {
  local args=(-f "${COMPOSE_FILE}")
  if [ -n "${ELEMO_COMPOSE_OVERRIDE:-}" ] && [ -f "${ELEMO_COMPOSE_OVERRIDE}" ]; then
    args+=(-f "${ELEMO_COMPOSE_OVERRIDE}")
  fi
  if [ -n "${ELEMO_COMPOSE_ENV_FILE:-}" ] && [ -f "${ELEMO_COMPOSE_ENV_FILE}" ]; then
    args+=(--env-file "${ELEMO_COMPOSE_ENV_FILE}")
  fi
  docker compose "${args[@]}" "$@"
}

function requireYes() {
  local confirmed=false
  local arg

  for arg in "$@"; do
    case "${arg}" in
      --yes|-y)
        confirmed=true
        ;;
      --with-oauth)
        WITH_OAUTH=true
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

function resolveCredentials() {
  if [ -z "${NEO4J_PASSWORD:-}" ] && [ -n "${NEO4J_AUTH:-}" ]; then
    NEO4J_USER="${NEO4J_AUTH%%/*}"
    NEO4J_PASSWORD="${NEO4J_AUTH#*/}"
  fi
  NEO4J_USER="${NEO4J_USER:-neo4j}"
  NEO4J_PASSWORD="${NEO4J_PASSWORD:-neo4jsecret}"
  POSTGRES_USER="${POSTGRES_USER:-elemo}"
  POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-pgsecret}"
  POSTGRES_DB="${POSTGRES_DB:-elemo}"
}

function cypher() {
  compose exec -T neo4j cypher-shell -u "${NEO4J_USER}" -p "${NEO4J_PASSWORD}" "$@"
}

function psql() {
  compose exec -T \
    -e "PGUSER=${POSTGRES_USER}" \
    -e "PGPASSWORD=${POSTGRES_PASSWORD}" \
    -e "PGDATABASE=${POSTGRES_DB}" \
    postgres psql -v ON_ERROR_STOP=1 "$@"
}

function wipeNeo4j() {
  log "wiping neo4j"
  echo "MATCH (n) DETACH DELETE n" | cypher
}

function bootstrapDatabase() {
  log "bootstrapping databases"
  cypher < "${QUERIES_DIR}/bootstrap.cypher"
  psql -d "${POSTGRES_DB}" < "${QUERIES_DIR}/bootstrap.sql"
  psql -d "${POSTGRES_DB}" <<'SQL'
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

function waitForServer() {
  local i
  for i in $(seq 1 90); do
    if compose exec -T elemo-server true >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  error "elemo-server did not become ready"
}

function loadDemoData() {
  local config_file="${ELEMO_PREFILL_CONFIG:-${ELEMO_CONFIG}}"
  local config_dir
  local config_base
  local configs_root
  local config_rel
  config_dir="$(dirname "${config_file}")"
  config_base="$(basename "${config_file}")"
  configs_root="${ELEMO_CONFIGS_DIR:-$(dirname "${config_dir}")}"
  config_rel="$(basename "${config_dir}")/${config_base}"

  log "loading demo data"
  if [ "${ELEMO_PREFILL_DOCKER:-}" = "1" ] || ! type go >/dev/null 2>&1; then
    checkInstalled "docker"
    # Overlay host configs so relative license/signing paths in config.yml resolve.
    docker run --rm \
      --network "${ELEMO_COMPOSE_NETWORK:-elemo-network}" \
      --volume "${ROOT_DIR}:/src" \
      --volume "${configs_root}:/src/configs:ro" \
      --workdir /src \
      -e GOPROXY=https://proxy.golang.org,direct \
      "${ELEMO_GOLANG_IMAGE:-golang:1.26-trixie}" \
      go run ./tools/workload-prefill \
        -config "/src/configs/${config_rel}" \
        -profile smoke \
        -yes
    return
  fi

  checkInstalled "go"
  go run "${TOOLS_DIR}/workload-prefill" \
    -config "${config_file}" \
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

function remintOAuth() {
  local add_out
  local py

  if [ "${WITH_OAUTH}" != "true" ]; then
    return 0
  fi

  if [ -z "${ELEMO_OAUTH_CALLBACK_URL:-}" ] \
    || [ -z "${ELEMO_OAUTH_SECRETS_FILE:-}" ] \
    || [ -z "${ELEMO_WEB_ENV:-}" ] \
    || [ -z "${ELEMO_API_BASE_URL:-}" ]; then
    error "--with-oauth requires ELEMO_OAUTH_CALLBACK_URL, ELEMO_OAUTH_SECRETS_FILE, ELEMO_WEB_ENV, ELEMO_API_BASE_URL"
  fi

  checkInstalled "python3"
  waitForServer
  log "reminting OAuth client"
  add_out="$(compose exec -T elemo-server \
    bin/elemo auth add-client --callback-url "${ELEMO_OAUTH_CALLBACK_URL}" 2>&1)"

  py="$(mktemp)"
  cat > "${py}" <<'PY'
import json
import re
import sys
from pathlib import Path

secrets_path, env_path, callback_url, api_base_url = sys.argv[1:5]
text = sys.stdin.read()
match = re.search(r"\{[^{}]*\"client-id\"[^{}]*\}", text, re.S)
if match is None:
    raise SystemExit("elemo auth add-client did not print client-id JSON")
payload = json.loads(match.group(0))
client_id = payload.get("client-id")
client_secret = payload.get("client-secret")
if not client_id or not client_secret:
    raise SystemExit("OAuth JSON missing client-id or client-secret")
secrets = Path(secrets_path)
secrets.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
secrets.write_text(
    json.dumps(
        {
            "client_id": client_id,
            "client_secret": client_secret,
            "callback_url": callback_url,
            "api_base_url": api_base_url,
        }
    )
    + "\n"
)
secrets.chmod(0o600)
env = Path(env_path)
lines = env.read_text().splitlines() if env.exists() else []
values = {
    "API_BASE_URL": api_base_url,
    "AUTH_CLIENT_ID": client_id,
    "AUTH_CLIENT_SECRET": client_secret,
}
seen = set()
out = []
for line in lines:
    key = line.split("=", 1)[0] if "=" in line else ""
    if key in values:
        out.append("%s=%s" % (key, values[key]))
        seen.add(key)
    else:
        out.append(line)
for key, value in values.items():
    if key not in seen:
        out.append("%s=%s" % (key, value))
env.parent.mkdir(parents=True, exist_ok=True)
env.write_text("\n".join(out) + "\n")
env.chmod(0o640)
PY
  printf '%s' "${add_out}" | python3 "${py}" \
    "${ELEMO_OAUTH_SECRETS_FILE}" \
    "${ELEMO_WEB_ENV}" \
    "${ELEMO_OAUTH_CALLBACK_URL}" \
    "${ELEMO_API_BASE_URL}"
  rm -f "${py}"

  if [ -n "${ELEMO_WEB_SERVICE:-}" ] && command -v systemctl >/dev/null 2>&1; then
    if systemctl cat "${ELEMO_WEB_SERVICE}" >/dev/null 2>&1; then
      log "restarting ${ELEMO_WEB_SERVICE}"
      systemctl restart "${ELEMO_WEB_SERVICE}"
    fi
  fi
}

requireYes "$@"
checkInstalled "docker"
resolveCredentials

if [ "${ELEMO_PREFILL_DOCKER:-}" != "1" ] && type go >/dev/null 2>&1; then
  :
elif [ "${ELEMO_PREFILL_DOCKER:-}" != "1" ]; then
  log "go not in PATH; workload-prefill will run in Docker"
fi

wipeNeo4j
emptyS3
bootstrapDatabase
wipePluginInstallations
loadDemoData
flushRedis
reindexSearch
reloadPluginRuntime
remintOAuth

success "demo data reset"
