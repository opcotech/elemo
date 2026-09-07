#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="${MISE_PROJECT_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
COVER_OUT="${ROOT_DIR}/.coverage.out"
COVER_UNIT="${ROOT_DIR}/.coverage.unit.out"
COVER_INTEGRATION="${ROOT_DIR}/.coverage.integration.out"
IGNORE_PATTERN='mode: atomic|testutil|tools|cmd|http/api'

rm -f "${COVER_OUT}"
echo "mode: atomic" > "${COVER_OUT}"
for file in "${COVER_UNIT}" "${COVER_INTEGRATION}"; do
  grep -Ev "${IGNORE_PATTERN}" "${file}" >> "${COVER_OUT}"
done
rm -f "${COVER_UNIT}" "${COVER_INTEGRATION}"
go tool cover -func "${COVER_OUT}"
