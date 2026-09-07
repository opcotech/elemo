#!/usr/bin/env bash

set -euo pipefail

# Idempotent local toolchain and dependency setup.
# Does not start services or modify databases.

if [ "${CI:-}" == "true" ]; then
  set -x
fi

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
# shellcheck source=common.sh
source "${ROOT_DIR}/scripts/common.sh"

readonly PLAYWRIGHT_BROWSERS="${PLAYWRIGHT_BROWSERS:-}"

checkInstalled "jq"
checkInstalled "pnpm"
checkInstalled "go"
checkInstalled "openssl"

log "installing locked Mise tools"
if [ -z "${MISE_BIN:-}" ] && ! type mise >/dev/null 2>&1; then
  error "mise is not on PATH; install Mise, then re-run bootstrap"
fi
# --locked applies to every active config. Global tools in
# ~/.config/mise/config.toml are not in this repo's mise.lock.
MISE_LOCKED_SCOPES="${MISE_LOCKED_SCOPES:-project}" mise install --locked

log "downloading Go modules"
go mod download

log "installing Go tools from go.mod"
go install tool

log "installing frontend dependencies"
pnpm --dir web install --frozen-lockfile

log "installing website dependencies"
pnpm --dir website install --frozen-lockfile

log "installing email toolchain dependencies"
pnpm --dir tools/email install --frozen-lockfile

generateConfigIfMissing

log "generating frontend API client"
pnpm --dir web generate

if [ -n "${PLAYWRIGHT_BROWSERS}" ]; then
  log "installing Playwright browsers: ${PLAYWRIGHT_BROWSERS}"
  # shellcheck disable=SC2206
  browsers=(${PLAYWRIGHT_BROWSERS})
  pnpm --dir web exec playwright install --with-deps "${browsers[@]}"
else
  log "installing Playwright browsers"
  pnpm --dir web exec playwright install --with-deps
fi

success "bootstrap finished; run \"scripts/dev-demo-init.sh --yes\" then \"mise run dev\""
