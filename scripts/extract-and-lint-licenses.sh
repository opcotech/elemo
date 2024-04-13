#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/../")"
LICENSE_DECISIONS_FILE="${ROOT_DIR}/.license-decisions.yml"

# The following licenses are not included in the SPDX license list, but are
# used by some of the dependencies. We need to add them to the license
# decisions file to avoid the license check to fail. Also, we need to add
# them to the license list file to avoid the license list check to fail.
INVALID_LICENSES="(Public Domain)"

function extractLicenses() {
  license_finder \
    --log-directory="/tmp/lf/" \
    --decisions-file="${LICENSE_DECISIONS_FILE}" \
    --project-path="${1}"
}

# shellcheck disable=SC2001
# shellcheck disable=SC2155
function lintLicenseCompatibility() {
  local licenses="$(yq '.[] | select(index(":permit"))[1]' < "${LICENSE_DECISIONS_FILE}")"
  local andExpression=$(echo "${licenses}" | egrep -v "${INVALID_LICENSES}" | xargs -n1 -I{} echo -n "{} and " | sed 's/ and $//')
  local candidates=$(flict outbound-candidate "${andExpression}")

  if [ "$(echo "${candidates}" | jq '. | length > 0')" != "true" ]; then
    echo "No OSS license candidates found for ${andExpression}"
  else
    echo "Possible OSS license candidates: ${candidates}"
  fi
}

extractLicenses "${ROOT_DIR}" # Backend service
# extractLicenses "${ROOT_DIR}/web" # Web application
lintLicenseCompatibility
