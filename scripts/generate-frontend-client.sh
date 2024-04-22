#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
PACKAGE_DIR="${ROOT_DIR}/web/packages/elemo-client"

pnpm dlx openapi-typescript-codegen \
    --input "${ROOT_DIR}/api/openapi/openapi.yaml" \
    --output "${PACKAGE_DIR}" \
    --exportSchemas true

pnpm exec prettier --plugin-search-dir "${PACKAGE_DIR}" --write "${PACKAGE_DIR}"