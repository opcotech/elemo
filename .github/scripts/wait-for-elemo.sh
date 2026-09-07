#!/usr/bin/env bash

set -euo pipefail

# Wait until the Elemo API heartbeat endpoint responds.
#
# Usage:
#   .github/scripts/wait-for-elemo.sh [base-url]
#
# Environment:
#   ELEMO_WAIT_ATTEMPTS  Number of attempts (default: 15)
#   ELEMO_WAIT_SLEEP     Seconds between attempts (default: 2)

BASE_URL="${1:-http://127.0.0.1:35478}"
ATTEMPTS="${ELEMO_WAIT_ATTEMPTS:-15}"
SLEEP_SECONDS="${ELEMO_WAIT_SLEEP:-2}"
COMPOSE_FILE="${ELEMO_COMPOSE_FILE:-deploy/docker/docker-compose.yml}"

attempt=0
while [ "${attempt}" -lt "${ATTEMPTS}" ]; do
  if curl -sf "${BASE_URL}/v1/system/heartbeat" > /dev/null 2>&1; then
    echo "elemo-server is ready"
    exit 0
  fi
  echo "Waiting for elemo-server to be ready... (attempt $((attempt + 1))/${ATTEMPTS})"
  sleep "${SLEEP_SECONDS}"
  attempt=$((attempt + 1))
done

echo "ERROR: elemo-server did not become ready within $((ATTEMPTS * SLEEP_SECONDS)) seconds"
if [ -f "${COMPOSE_FILE}" ]; then
  docker compose -f "${COMPOSE_FILE}" ps
  docker compose -f "${COMPOSE_FILE}" logs elemo-server
fi
exit 1
