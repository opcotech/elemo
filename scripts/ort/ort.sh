#!/usr/bin/env bash

set -euo pipefail

# shellcheck source=scripts/common.sh
source "$(dirname "${BASH_SOURCE[0]:-$0}")/../common.sh"

# Pin image and ort-config together. Update both when bumping ORT.
# Minimal image includes Go, PNPM, and ScanCode. Digest for 92.4.0:
# sha256:770bdd803a77c6754ccf59baa78954754f71590fb27dcf7d5e5992f85ac445b3
ORT_IMAGE="${ORT_IMAGE:-ghcr.io/oss-review-toolkit/ort-minimal:92.4.0@sha256:770bdd803a77c6754ccf59baa78954754f71590fb27dcf7d5e5992f85ac445b3}"
ORT_CONFIG_REPO="${ORT_CONFIG_REPO:-https://github.com/oss-review-toolkit/ort-config.git}"
ORT_CONFIG_REVISION="${ORT_CONFIG_REVISION:-ec122d3426571afe84344a83d5e16c99cda20325}"
ORT_JAVA_OPTIONS="${ORT_JAVA_OPTIONS:--Xmx8192m}"
GO_LICENSES_VERSION="${GO_LICENSES_VERSION:-v1.6.0}"
# Full CI scans Elemo source only. Full dependency ScanCode is opt-in (too slow for CI).
ORT_SCAN_PACKAGE_TYPES="${ORT_SCAN_PACKAGE_TYPES:-PROJECT}"

ORT_DIR="${ROOT_DIR}/.ort"
ORT_CONFIG_DIR="${ORT_DIR}/config"
ORT_RESULTS_DIR="${ORT_DIR}/results"
ORT_OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
ORT_GITCONFIG="${ORT_DIR}/gitconfig"

ORT_CONTAINER_HOME="/home/ort"
ORT_CONTAINER_DATA_DIR="${ORT_CONTAINER_HOME}/.ort"
ORT_CONTAINER_PROJECT_DIR="/project"
ORT_CONTAINER_CONFIG_FILE="${ORT_CONTAINER_DATA_DIR}/config/config.yml"
ORT_CONTAINER_RULES_FILE="${ORT_CONTAINER_DATA_DIR}/config/evaluator.rules.kts"

overlay_policy() {
  mkdir -p "${ORT_CONFIG_DIR}"
  cp "${ORT_OVERLAY_DIR}/evaluator.rules.kts" "${ORT_CONFIG_DIR}/evaluator.rules.kts"
  cp "${ORT_OVERLAY_DIR}/config.yml" "${ORT_CONFIG_DIR}/config.yml"
  if [[ -f "${ORT_DIR}/go-license-curations.yml" ]]; then
    cp "${ORT_DIR}/go-license-curations.yml" "${ORT_CONFIG_DIR}/curations.yml"
  fi
}

usage() {
  cat <<'EOF'
Usage: scripts/ort/ort.sh [prepare|analyze|scan|evaluate|advise|report|pr|run]

  prepare   Fetch pinned ort-config and overlay Elemo policy files
  analyze   Run the ORT analyzer
  scan      Run ScanCode (default: Elemo projects only)
  evaluate  Run the ORT evaluator (policy rules)
  advise    Run the ORT advisor (OSV; reported, not a gate)
  report    Generate SPDX, CycloneDX, WebApp, NOTICE, and the legal/ bundle
  pr        prepare + analyze + evaluate (PR policy gate; no ScanCode)
  run       prepare + analyze + scan + evaluate + advise + report (default)

ScanCode on every dependency is too slow for CI. Default scan is --package-types PROJECT.
Set ORT_SCAN_PACKAGE_TYPES=PACKAGE,PROJECT for a full dependency scan.

Go module licenses come from go-licenses (module cache), npm from declared package.json licenses.
The evaluate step fails the command when policy violations reach ERROR.
EOF
}

prepare() {
  checkInstalled git
  log "preparing ORT config at ${ORT_CONFIG_REVISION}"

  mkdir -p "${ORT_RESULTS_DIR}"

  if [[ -d "${ORT_CONFIG_DIR}/.git" ]]; then
    git -C "${ORT_CONFIG_DIR}" remote set-url origin "${ORT_CONFIG_REPO}"
  else
    rm -rf "${ORT_CONFIG_DIR}"
    mkdir -p "${ORT_CONFIG_DIR}"
    git -C "${ORT_CONFIG_DIR}" init --quiet
    git -C "${ORT_CONFIG_DIR}" remote add origin "${ORT_CONFIG_REPO}"
  fi

  git -C "${ORT_CONFIG_DIR}" fetch --quiet --depth 1 origin "${ORT_CONFIG_REVISION}"
  git -C "${ORT_CONFIG_DIR}" checkout --quiet --force FETCH_HEAD

  overlay_policy

  cat > "${ORT_GITCONFIG}" <<'EOF'
[url "https://github.com/"]
	insteadOf = ssh://git@github.com/
	insteadOf = git@github.com:
EOF

  success "ORT config ready"
}

host_go_cache_dirs() {
  local gomodcache gocache
  if command -v go >/dev/null 2>&1; then
    gomodcache="$(go env GOMODCACHE)"
    gocache="$(go env GOCACHE)"
  fi
  HOST_GOMODCACHE="${gomodcache:-${ORT_DIR}/gomodcache}"
  HOST_GOCACHE="${gocache:-${ORT_DIR}/gocache}"
}

run_ort() {
  checkInstalled docker

  host_go_cache_dirs

  # Caches live on host-owned mounts so the container can run as the invoking
  # user (PNPM stash + go list write to the project / module cache). Reuse the
  # host Go module cache when available so analyze does not re-download what
  # setup-go and go-licenses already populated.
  mkdir -p "${ORT_CONFIG_DIR}" "${ORT_RESULTS_DIR}" \
    "${ORT_DIR}/gopath" "${HOST_GOMODCACHE}" "${HOST_GOCACHE}"

  # Image user `ort` is uid 1000; CI (and local) often run as another uid.
  # Bind-mount a writable home so that uid can traverse /home/ort to reach
  # the nested .ort data mount (config, caches, corepack). Keep the temp dir
  # outside .ort to avoid a recursive bind.
  local container_home
  container_home="$(mktemp -d "${TMPDIR:-/tmp}/ort-home.XXXXXX")"
  mkdir -p "${container_home}/.cache/node/corepack" "${container_home}/.local"

  local exit_code=0
  docker run --rm \
    --user "$(id -u):$(id -g)" \
    -e HOME="${ORT_CONTAINER_HOME}" \
    -e COREPACK_HOME="${ORT_CONTAINER_HOME}/.cache/node/corepack" \
    -e COREPACK_ENABLE_DOWNLOAD_PROMPT=0 \
    -e JDK_JAVA_OPTIONS="${ORT_JAVA_OPTIONS}" \
    -e ORT_DATA_DIR="${ORT_CONTAINER_DATA_DIR}" \
    -e ORT_CONFIG_DIR="${ORT_CONTAINER_DATA_DIR}/config" \
    -e GIT_CONFIG_GLOBAL="${ORT_CONTAINER_DATA_DIR}/gitconfig" \
    -e GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}" \
    -e GOSUMDB="${GOSUMDB:-sum.golang.org}" \
    -e GOPATH="${ORT_CONTAINER_DATA_DIR}/gopath" \
    -e GOMODCACHE="${ORT_CONTAINER_DATA_DIR}/gomodcache" \
    -e GOCACHE="${ORT_CONTAINER_DATA_DIR}/gocache" \
    -v "${container_home}:${ORT_CONTAINER_HOME}" \
    -v "${ORT_DIR}:${ORT_CONTAINER_DATA_DIR}" \
    -v "${HOST_GOMODCACHE}:${ORT_CONTAINER_DATA_DIR}/gomodcache" \
    -v "${HOST_GOCACHE}:${ORT_CONTAINER_DATA_DIR}/gocache" \
    -v "${ROOT_DIR}:${ORT_CONTAINER_PROJECT_DIR}" \
    -w "${ORT_CONTAINER_PROJECT_DIR}" \
    "${ORT_IMAGE}" \
    --info \
    --config "${ORT_CONTAINER_CONFIG_FILE}" \
    "$@" || exit_code=$?

  rm -rf "${container_home}"
  return "${exit_code}"
}

# pnpm-workspace.yaml files ORT may mutate via `pnpm config set node-linker`.
ORT_WORKSPACE_FILES=(
  web/pnpm-workspace.yaml
  website/pnpm-workspace.yaml
  build/email/pnpm-workspace.yaml
)

snapshot_workspace_files() {
  local snapshot_dir="$1"
  local f
  for f in "${ORT_WORKSPACE_FILES[@]}"; do
    if [[ -f "${ROOT_DIR}/${f}" ]]; then
      mkdir -p "${snapshot_dir}/$(dirname "${f}")"
      cp "${ROOT_DIR}/${f}" "${snapshot_dir}/${f}"
    fi
  done
}

restore_workspace_files() {
  local snapshot_dir="$1"
  local f
  for f in "${ORT_WORKSPACE_FILES[@]}"; do
    if [[ -f "${snapshot_dir}/${f}" ]]; then
      cp "${snapshot_dir}/${f}" "${ROOT_DIR}/${f}"
    fi
  done
}

analyze() {
  host_go_cache_dirs
  log "running ORT analyzer (GOMODCACHE=${HOST_GOMODCACHE})"
  local exit_code=0
  local snapshot_dir
  snapshot_dir="$(mktemp -d "${TMPDIR:-/tmp}/ort-workspace.XXXXXX")"
  snapshot_workspace_files "${snapshot_dir}"
  # Go licenses are applied via $ORT_CONFIG_DIR/curations.yml (DefaultFile) and again
  # at evaluate with --package-curations-file. Analyze has no --package-curations-file.
  overlay_policy
  run_ort analyze \
    -i "${ORT_CONTAINER_PROJECT_DIR}" \
    -o "${ORT_CONTAINER_DATA_DIR}/results" \
    -f JSON \
    --repository-configuration-file "${ORT_CONTAINER_PROJECT_DIR}/.ort.yml" \
    || exit_code=$?
  restore_workspace_files "${snapshot_dir}"
  rm -rf "${snapshot_dir}"
  chmod -R a+rX "${ORT_RESULTS_DIR}" || true

  if [[ "${exit_code}" -eq 0 ]]; then
    success "analyzer finished"
    return 0
  fi

  if [[ "${exit_code}" -eq 2 ]]; then
    log "analyzer finished with unresolved issues (not a policy gate); see ${ORT_RESULTS_DIR}/analyzer-result.json"
    return 0
  fi

  error "analyzer failed with exit code ${exit_code}"
}

curate_go_licenses() {
  checkInstalled go
  checkInstalled python3

  log "concluding Go module licenses with go-licenses ${GO_LICENSES_VERSION}"
  mkdir -p "${ORT_DIR}/bin" "${ORT_RESULTS_DIR}"

  local gobin="${ORT_DIR}/bin"
  if [[ ! -x "${gobin}/go-licenses" ]]; then
    GOBIN="${gobin}" go install "github.com/google/go-licenses@${GO_LICENSES_VERSION}"
  fi

  local modules_file licenses_file
  modules_file="$(mktemp "${TMPDIR:-/tmp}/ort-go-modules.XXXXXX")"
  licenses_file="$(mktemp "${TMPDIR:-/tmp}/ort-go-licenses.XXXXXX")"

  {
    go list -m all
    go -C "${ROOT_DIR}/tools/openapi-assemble" list -m all
    go -C "${ROOT_DIR}/tools/pre-mailer" list -m all
  } | awk 'NF >= 2 { print $1, $2 }' | sort -u > "${modules_file}"

  # Ensure module cache has LICENSE files for transitive / test-only modules.
  go mod download all >/dev/null 2>&1 || true
  go -C "${ROOT_DIR}/tools/openapi-assemble" mod download all >/dev/null 2>&1 || true
  go -C "${ROOT_DIR}/tools/pre-mailer" mod download all >/dev/null 2>&1 || true

  : > "${licenses_file}"
  "${gobin}/go-licenses" csv ./... >> "${licenses_file}" 2>/dev/null || true
  (cd "${ROOT_DIR}/tools/openapi-assemble" && "${gobin}/go-licenses" csv ./... >> "${licenses_file}" 2>/dev/null) || true
  (cd "${ROOT_DIR}/tools/pre-mailer" && "${gobin}/go-licenses" csv ./... >> "${licenses_file}" 2>/dev/null) || true

  GOMODCACHE="$(go env GOMODCACHE)" \
    python3 "${ORT_OVERLAY_DIR}/go-licenses-curations.py" \
    "${modules_file}" \
    "${licenses_file}" \
    "${ORT_DIR}/go-license-curations.yml"

  # DefaultFile provider reads $ORT_CONFIG_DIR/curations.yml. Keep a copy there so
  # curations apply even when a custom File provider entry is ignored.
  mkdir -p "${ORT_CONFIG_DIR}"
  cp "${ORT_DIR}/go-license-curations.yml" "${ORT_CONFIG_DIR}/curations.yml"

  rm -f "${modules_file}" "${licenses_file}"
}

rewrite_ssh_git_urls() {
  local file="$1"
  python3 - "$file" <<'PY'
from pathlib import Path
import sys
path = Path(sys.argv[1])
text = path.read_text()
path.write_text(
    text.replace("ssh://git@github.com/", "https://github.com/").replace("git@github.com:", "https://github.com/")
)
PY
}

scan() {
  log "running ORT scanner (ScanCode, package types: ${ORT_SCAN_PACKAGE_TYPES})"
  if [[ ! -f "${ORT_RESULTS_DIR}/analyzer-result.json" ]]; then
    error "analyzer-result.json not found; run analyze first"
  fi

  rewrite_ssh_git_urls "${ORT_RESULTS_DIR}/analyzer-result.json"
  overlay_policy
  local exit_code=0
  # shellcheck disable=SC2086
  run_ort scan \
    -i "${ORT_CONTAINER_DATA_DIR}/results/analyzer-result.json" \
    -o "${ORT_CONTAINER_DATA_DIR}/results" \
    -f JSON \
    --scanners ScanCode \
    --project-scanners ScanCode \
    --package-types "${ORT_SCAN_PACKAGE_TYPES}" \
    || exit_code=$?
  chmod -R a+rX "${ORT_RESULTS_DIR}" || true

  if [[ "${exit_code}" -eq 0 ]]; then
    success "scanner finished"
    return 0
  fi

  if [[ "${exit_code}" -eq 2 ]]; then
    log "scanner finished with unresolved issues (not a policy gate); see ${ORT_RESULTS_DIR}/scan-result.json"
    return 0
  fi

  error "scanner failed with exit code ${exit_code}"
}

ort_eval_input() {
  if [[ -f "${ORT_RESULTS_DIR}/scan-result.json" ]]; then
    echo "scan-result.json"
  else
    echo "analyzer-result.json"
  fi
}

evaluate() {
  overlay_policy
  local input="${1:-}"
  local curations_args=()
  if [[ -n "${input}" ]]; then
    if [[ ! -f "${ORT_RESULTS_DIR}/${input}" ]]; then
      error "${input} not found; run analyze first"
    fi
  else
    input="$(ort_eval_input)"
  fi
  log "running ORT evaluator (${input})"
  if [[ -f "${ORT_DIR}/go-license-curations.yml" ]]; then
    curations_args+=(--package-curations-file "${ORT_CONTAINER_DATA_DIR}/go-license-curations.yml")
  fi
  local exit_code=0
  run_ort evaluate \
    -i "${ORT_CONTAINER_DATA_DIR}/results/${input}" \
    -o "${ORT_CONTAINER_DATA_DIR}/results" \
    -f JSON \
    -r "${ORT_CONTAINER_RULES_FILE}" \
    --repository-configuration-file "${ORT_CONTAINER_PROJECT_DIR}/.ort.yml" \
    ${curations_args[@]+"${curations_args[@]}"} \
    || exit_code=$?
  chmod -R a+rX "${ORT_RESULTS_DIR}" || true

  if [[ "${exit_code}" -eq 0 ]]; then
    success "evaluator finished with no ERROR violations"
    return 0
  fi

  if [[ "${exit_code}" -eq 2 ]]; then
    log "evaluator reported ERROR policy violations; see ${ORT_RESULTS_DIR}/evaluation-result.json"
    return 2
  fi

  error "evaluator failed with exit code ${exit_code}"
}

advise() {
  log "running ORT advisor"
  local input="${ORT_RESULTS_DIR}/evaluation-result.json"
  if [[ ! -f "${input}" ]]; then
    input="${ORT_RESULTS_DIR}/analyzer-result.json"
  fi

  local exit_code=0
  run_ort advise \
    -i "${ORT_CONTAINER_DATA_DIR}/results/$(basename "${input}")" \
    -o "${ORT_CONTAINER_DATA_DIR}/results" \
    -f JSON \
    --advisors OSV \
    || exit_code=$?
  chmod -R a+rX "${ORT_RESULTS_DIR}" || true

  if [[ "${exit_code}" -eq 0 || "${exit_code}" -eq 2 ]]; then
    success "advisor finished (exit ${exit_code}; issues are not a CI gate)"
    return 0
  fi

  error "advisor failed with exit code ${exit_code}"
}

assemble_legal_bundle() {
  local legal_dir="${ORT_RESULTS_DIR}/legal"
  local notice_src=""

  mkdir -p "${legal_dir}"
  cp "${ROOT_DIR}/LICENSE" "${legal_dir}/LICENSE"
  cp "${ROOT_DIR}/LICENSE-COMMERCIAL" "${legal_dir}/LICENSE-COMMERCIAL"

  if [[ -f "${ORT_RESULTS_DIR}/NOTICE_DEFAULT" ]]; then
    notice_src="${ORT_RESULTS_DIR}/NOTICE_DEFAULT"
  elif [[ -f "${ORT_RESULTS_DIR}/NOTICE_DEFAULT.txt" ]]; then
    notice_src="${ORT_RESULTS_DIR}/NOTICE_DEFAULT.txt"
  else
    shopt -s nullglob
    local notice_candidates=("${ORT_RESULTS_DIR}"/NOTICE*)
    shopt -u nullglob
    if (( ${#notice_candidates[@]} > 0 )); then
      notice_src="${notice_candidates[0]}"
    fi
  fi
  if [[ -z "${notice_src}" || ! -f "${notice_src}" ]]; then
    error "ORT NOTICE file not found in ${ORT_RESULTS_DIR}; expected NOTICE_DEFAULT from PlainTextTemplate"
  fi
  cp "${notice_src}" "${legal_dir}/NOTICE"

  if [[ ! -f "${ORT_RESULTS_DIR}/bom.spdx.yml" || ! -f "${ORT_RESULTS_DIR}/bom.cyclonedx.json" ]]; then
    error "ORT SBOM files not found; expected bom.spdx.yml and bom.cyclonedx.json in ${ORT_RESULTS_DIR}"
  fi
  cp "${ORT_RESULTS_DIR}/bom.spdx.yml" "${legal_dir}/bom.spdx.yml"
  cp "${ORT_RESULTS_DIR}/bom.cyclonedx.json" "${legal_dir}/bom.cyclonedx.json"
}

report() {
  log "running ORT reporter"
  local input="${ORT_RESULTS_DIR}/advisor-result.json"
  if [[ ! -f "${input}" ]]; then
    input="${ORT_RESULTS_DIR}/evaluation-result.json"
  fi
  if [[ ! -f "${input}" ]]; then
    input="${ORT_RESULTS_DIR}/analyzer-result.json"
  fi

  run_ort report \
    -i "${ORT_CONTAINER_DATA_DIR}/results/$(basename "${input}")" \
    -o "${ORT_CONTAINER_DATA_DIR}/results" \
    -f CycloneDX,SpdxDocument,WebApp,PlainTextTemplate \
    --repository-configuration-file "${ORT_CONTAINER_PROJECT_DIR}/.ort.yml"
  chmod -R a+rX "${ORT_RESULTS_DIR}" || true
  assemble_legal_bundle
  chmod -R a+rX "${ORT_RESULTS_DIR}/legal" || true
  success "reporter finished; results in ${ORT_RESULTS_DIR} (legal bundle in ${ORT_RESULTS_DIR}/legal)"
}

run_pr() {
  prepare
  curate_go_licenses
  analyze

  local evaluate_exit=0
  # Force analyzer-result.json so a leftover local scan-result.json cannot
  # change the PR policy-gate input.
  evaluate analyzer-result.json || evaluate_exit=$?

  if [[ "${evaluate_exit}" -ne 0 ]]; then
    exit "${evaluate_exit}"
  fi
}

run_all() {
  prepare
  curate_go_licenses
  analyze
  scan

  local evaluate_exit=0
  evaluate || evaluate_exit=$?

  advise
  report

  if [[ "${evaluate_exit}" -ne 0 ]]; then
    exit "${evaluate_exit}"
  fi
}

command="${1:-run}"
case "${command}" in
  -h|--help|help)
    usage
    ;;
  prepare)
    prepare
    ;;
  curate)
    curate_go_licenses
    ;;
  analyze)
    analyze
    ;;
  scan)
    scan
    ;;
  evaluate)
    evaluate
    ;;
  advise)
    advise
    ;;
  report)
    report
    ;;
  pr)
    run_pr
    ;;
  run)
    run_all
    ;;
  *)
    usage
    error "unknown command '${command}'"
    ;;
esac
