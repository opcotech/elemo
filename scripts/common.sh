#!/usr/bin/env bash

set -euo pipefail

export ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
export CMD_DIR="${ROOT_DIR}/cmd/elemo"
export CONFIG_DIR="${ROOT_DIR}/configs/development"
export DOCKER_DEPLOY_DIR="${ROOT_DIR}/deploy/docker"
export PACKAGE_DIR="${ROOT_DIR}/web/packages/elemo-client"
export QUERIES_DIR="${ROOT_DIR}/assets/queries"
export SCRIPTS_DIR="${ROOT_DIR}/scripts"
export TOOLS_DIR="${ROOT_DIR}/tools"
export WEB_DIR="${ROOT_DIR}/web"

export ELEMO_CONFIG="${CONFIG_DIR}/config.local.gen.yml"

function checkInstalled() {
  local program="${1}"

  if ! type "${program}" 2>&1 > /dev/null; then
    echo "Couldn't find ${program} in your PATH. Make sure it is installed."
    exit 1
  fi
}

function waitAndPrint() {
  echo "waiting ${1} seconds to let the services boot"
  sleep "${1}"
}

function backupCopyFile() {
  local backupFile="${1}"
  local copyFile="${2}"

  [[ -f "${backupFile}" ]] && mv "${backupFile}" "${backupFile}.bkp"
  cp "${copyFile}" "${backupFile}"
}

function generateConfigIfMissing() {
  if [ ! -f "${ELEMO_CONFIG}" ]; then
      bash "${SCRIPTS_DIR}/generate-dev-config.sh"
  fi
}
