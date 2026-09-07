#!/usr/bin/env bash

set -euo pipefail

# Run Docker Compose with the same discovery as other project scripts.
#
# Examples:
#   ./scripts/compose.sh version
#   ./scripts/compose.sh -f deploy/docker/docker-compose.yml up -d
#   DOCKER_COMPOSE_PLUGIN=/path/to/docker-compose ./scripts/compose.sh ps

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
# shellcheck source=common.sh
source "${ROOT_DIR}/scripts/common.sh"

dockerCompose "$@"
