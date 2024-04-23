#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

function formatOutput() {
    echo -e "**${1}**\n\n\`\`\`${2}\n${3}\n\`\`\`\n"
}

function promptInformation() {
    read -p "This script provides some automation to help with bug reports.
During the execution, the script will run Docker and several other tools
as necessary to start the service and get the necessary information. Please
sure that Docker is running before executing this script!

The script may collect information you are not comfortable to share, such as
your OS username. Please ALWAYS check the output and remove any sensitive
information before filing the bug report.

To continue, press ENTER."
    clear;
}

function getOSInfo() {
    formatOutput "OS information" "text" "$(uname -a)"
}

function getLocalGoVersion() {
    formatOutput "Local Go version" "text" "$(go version)"
}

function getElemoVersion() {
    formatOutput "Elemo version" "json" $(
        docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" run \
            -T \
            --rm \
            --entrypoint 'bin/elemo version' elemo-server 2>&1 | grep -i version
    )
}

# Prompt some warning
promptInformation

# Run preflight
checkInstalled "docker"
checkInstalled "go"

# Generate dev config if missing
generateConfigIfMissing

# Get bug report details
getOSInfo
getLocalGoVersion
getElemoVersion
