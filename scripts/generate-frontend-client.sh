#!/usr/bin/env bash

set -euo pipefail

if [ "$CI" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

function generateAPIClient() {
    pnpm dlx openapi-typescript-codegen \
        --input "${ROOT_DIR}/api/openapi/openapi.yaml" \
        --output "${PACKAGE_DIR}" \
        --exportSchemas true

    pnpm exec prettier --plugin-search-dir "${PACKAGE_DIR}" --write "${PACKAGE_DIR}"
}

# Run preflight
checkInstalled "pnpm"

# Generate front-end API client
generateAPIClient

success "the front-end client is generated successfully"
