#!/bin/bash

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/../")"
OPENAPI_PATH="${ROOT_DIR}/../api/openapi/openapi.final.yaml"
API_PATH="${ROOT_DIR}/lib/api"

npx swagger-typescript-api \
  --add-readonly \
  --responses \
  --extract-enums \
  --extract-response-error \
  --extract-response-body \
  --extract-request-params \
  --extract-request-body \
  --api-class-name "Client" \
  --path "${OPENAPI_PATH}" \
  --output "${API_PATH}" \
  --name "api.ts";
