#!/usr/bin/env bash

set -euo pipefail

CONFIG_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/../configs/development")"

function generateCert() {
  local host="${1}"

  if [[ "${host}" == "docker" ]]; then
    host="0.0.0.0"
    local suffix=""
  else
    local suffix=".local"
  fi

  mkcert -install
  mkcert -cert-file "${CONFIG_DIR}/cert${suffix}.gen.pem" -key-file "${CONFIG_DIR}/key${suffix}.gen.pem" "${host}"
}

function generateSigningKey() {
  openssl genrsa -out "${CONFIG_DIR}/signing-key.gen.pem" 2048
  openssl req -new -x509 -days 3650 \
    -key "${CONFIG_DIR}/signing-key.gen.pem" \
    -out "${CONFIG_DIR}/signing-cert.gen.pem"
}

function generateLicenseKey() {
  go run tools/license-generator/main.go \
    -validity-period 3650 \
    -email info@example.com \
    -organization "ACME Inc." \
    -private-key configs/test/generator.key \
    -license "${CONFIG_DIR}/license.gen.key" \
    -quota "users=999,organizations=999,documents=999,namespaces=999,roles=999"
}

function generateConfigFile() {
  local host="${1}"

  if [[ "${host}" == "docker" ]]; then
    host="0.0.0.0"
    local suffix=""
    local redis_host="redis"
    local neo4j_host="neo4j"
    local postgres_host="postgres"
    local otel_collector_host="otel_collector"
  else
    local suffix=".local"
    local redis_host=host
    local neo4j_host=host
    local postgres_host=host
    local otel_collector_host=host
  fi


  cat <<EOF > "${CONFIG_DIR}/config${suffix}.gen.yml"
log:
  level: "info"

license:
  file: "configs/development/license.gen.key"

tls:
  cert_file: "configs/development/cert${suffix}.gen.pem"
  key_file: "configs/development/key${suffix}.gen.pem"

server:
  address: "${host}:35478"
  read_timeout: 10
  write_timeout: 5
  request_throttle_limit: 350
  request_throttle_backlog: 750
  request_throttle_timeout: 10
  cors:
    enabled: true
    allowed_origins:
      - "http://127.0.0.1:3000"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allowed_headers:
      - "*"
    allow_credentials: false
    max_age: 86400
  session:
    cookie_name: "elemo_session"
    max_age: 86400
    is_secure: false

cache_database:
  host: ${redis_host}
  port: 6379
  username: ""
  password: ""
  database: "0"
  dial_timeout: 3
  read_timeout: 2
  write_timeout: 2
  pool_size: 100
  max_idle_connections: 25
  min_idle_connections:  5
  connection_max_idle_time: 250
  connection_max_lifetime: 300

graph_database:
  host: ${neo4j_host}
  port: 7687
  username: neo4j
  password: neo4jsecret
  name: neo4j
  max_transaction_retry_time: 3
  max_connection_pool_size: 100
  max_connection_lifetime: 300
  connection_acquisition_timeout: 60
  socket_connect_timeout: 5
  socket_keepalive: true
  fetch_size: 150

relational_database:
  host: ${postgres_host}
  port: 5432
  username: elemo
  password: pgsecret
  name: elemo
  max_connections: 100
  max_connection_lifetime: 300
  max_connection_idle_time: 10
  min_connections: 5

metrics_server:
  address: "${host}:35479"
  read_timeout: 10
  write_timeout: 5

tracing:
  service_name: 'elemo'
  collector_endpoint: '${otel_collector_host}:4318'
  trace_ratio: 0.75
EOF
}

generateCert "docker"
generateCert "192.168.0.23"
generateSigningKey
generateLicenseKey
generateConfigFile "docker"
generateConfigFile "127.0.0.1"
