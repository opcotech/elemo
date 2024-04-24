#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
source "${ROOT_DIR}/scripts/common.sh";

CURRENT_TIME=$(date "+%s")
REPORT_FILE="${TMPDIR}/elemo-bug-report-${CURRENT_TIME}.md"

function formatOutput() {
    echo -e "**${1}**\n\n\`\`\`${2}\n${3}\n\`\`\`\n" >> "${REPORT_FILE}"
}

function promptInformation() {
    read -p "This script provides some automation to help with bug reports.
During the execution, the script will run Docker and several other tools
as necessary to start the service and get the necessary information. Please
make sure that Docker is running before executing this script! If Elemo was
never built before, getting its version will take a while.

The script may collect information you are not comfortable to share, such as
your OS username. Please ALWAYS check the output and remove any sensitive
information before filing the bug report.

The collected information is saved into ${REPORT_FILE}.

To continue, press ENTER."
}

function replaceUsername() {
    log "removing username from ${REPORT_FILE}"
    sed -i s/$(whoami)/[REDACTED]/g "${REPORT_FILE}"
}

function getOSInfo() {
    log "getting OS information" 
    formatOutput "OS information" "text" "$(uname -a)"
}

function getDockerInfo() {
    log "getting Docker information"
    formatOutput "Docker information" "text" "$(docker info)"
}

function getLocalGoVersion() {
    log "getting Go version"
    formatOutput "Local Go version" "text" "$(go version)"
}

function getElemoVersion() {
    if [ -z "$(docker images | grep elemo-server)" ]; then
        log "building Elemo"
        docker compose -f "${DOCKER_DEPLOY_DIR}/docker-compose.yml" build elemo-server 2>&1
    fi

    log "getting Elemo version"
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
checkInstalled "docker2"
checkInstalled "go"
checkInstalled "sed"

# Generate dev config if missing
generateConfigIfMissing

# Get bug report details
getOSInfo
getDockerInfo
getLocalGoVersion
getElemoVersion

# Trying to anonymize the resulting file
replaceUsername

success "the report file is available at ${REPORT_FILE}"
