#!/usr/bin/env bash

set -euo pipefail

CI="${CI:-'false'}"

if [ "$CI" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

function generateAPIClient() {
    cd "${WEB_DIR}"
    pnpm exec openapi-ts
}

# Run preflight
checkInstalled "pnpm"

# Generate front-end API client
generateAPIClient

success "the front-end client is generated successfully"
