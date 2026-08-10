#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
WEB_DIR="${ROOT_DIR}/web"
SNAPSHOT_DIR="$(mktemp -d)"

trap 'rm -rf "${SNAPSHOT_DIR}"' EXIT

cp -R "${WEB_DIR}/src/lib/client" "${SNAPSHOT_DIR}/client"
"${ROOT_DIR}/scripts/generate-frontend-client.sh"

if ! diff -ru "${SNAPSHOT_DIR}/client" "${WEB_DIR}/src/lib/client"; then
  echo "Generated frontend client is out of date" >&2
  exit 1
fi

echo "Generated frontend client is reproducible"
