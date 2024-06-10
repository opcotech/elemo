#!/usr/bin/env bash

set -euo pipefail

if [ "${CI:-}" == "true" ]; then
  set -x
fi

export TMPDIR="${TMPDIR:-/tmp}"
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

normal=""
red=""
green=""
cyan=""

if test -t 1; then
  ncolors=$(tput colors)
  if test -n "$ncolors" && test $ncolors -ge 8; then
    normal="$(tput sgr0)"
    red="$(tput setaf 1)"
    green="$(tput setaf 2)"
    cyan="$(tput setaf 6)"
  fi
fi

function log() {
    echo -e "${cyan}INFO${normal}\t${1}" 1>&2;
}

function success() {
    echo -e "${green}DONE${normal}\t${1}" 1>&2;
}

function error() {
    echo -e "${red}ERROR${normal}\t${1}" 1>&2;
    exit 1;
}

function checkInstalled() {
  local program="${1}"

  if ! type "${program}" 2>&1 > /dev/null; then
    error "couldn't find ${program} in your PATH. Make sure it is installed."
  fi
}

function waitAndPrint() {
  log "waiting ${1} seconds to let the services boot"
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
