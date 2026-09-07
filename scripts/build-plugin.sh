#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="${MISE_PROJECT_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
PLUGIN_DIST_DIR="${ROOT_DIR}/dist/plugins"

log() {
  echo -e "INFO\t${1}" 1>&2
}

build_one() {
  local plugin_dir="${1}"

  if [ ! -d "${plugin_dir}" ]; then
    plugin_dir="${ROOT_DIR}/plugins/${1}"
  fi
  plugin_dir="$(cd "${plugin_dir}" && pwd)"

  if [ ! -f "${plugin_dir}/plugin.yaml" ]; then
    echo "ERROR\tno plugin.yaml in ${plugin_dir}" 1>&2
    exit 1
  fi

  local plugin_id
  plugin_id="$(awk '/^id:/{print $2; exit}' "${plugin_dir}/plugin.yaml")"
  local plugin_zip="${PLUGIN_DIST_DIR}/${plugin_id}.zip"

  log "build ${plugin_id} plugin"
  mkdir -p "${PLUGIN_DIST_DIR}" "${plugin_dir}/backend"
  CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go -C "${plugin_dir}" build -buildmode=c-shared -o backend/plugin.wasm .
  rm -f "${plugin_dir}/backend/plugin.h"
  (
    cd "${ROOT_DIR}/web"
    NODE_ENV=production pnpm exec vite build --config "${plugin_dir}/frontend/vite.config.ts"
  )
  rm -f "${plugin_zip}"
  (
    cd "${plugin_dir}"
    zip -q "${plugin_zip}" plugin.yaml backend/plugin.wasm frontend/index.js
  )
  echo "${plugin_zip}"
}

if [ "${#}" -eq 0 ]; then
  shopt -s nullglob
  for plugin_dir in "${ROOT_DIR}"/plugins/*/; do
    if [ -f "${plugin_dir}/plugin.yaml" ]; then
      build_one "${plugin_dir}"
    fi
  done
  exit 0
fi

build_one "${1}"
