#!/usr/bin/env bash

set -euo pipefail

# Format and apply safe lint fixes to paths passed by pre-commit.
# Go files use gofumpt/goimports; web/ and website/ files use Biome.
#
# Examples:
#   mise run pre-commit-files -- internal/model/user.go
#   mise run pre-commit-files -- web/src/lib/utils.ts website/src/pages/index.astro
#   ./scripts/pre-commit-files.sh cmd/elemo/main.go

ROOT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]:-$0}")/..")"
# shellcheck source=common.sh
source "${ROOT_DIR}/scripts/common.sh"

cd "${ROOT_DIR}"

web_files=()
website_files=()
go_files=()

is_generated_go() {
  local path="${1}"
  case "${path}" in
    *_gen.go) return 0 ;;
    internal/transport/http/api/*) return 0 ;;
  esac
  return 1
}

for path in "${@}"; do
  case "${path}" in
    web/*)
      web_files+=("${path#web/}")
      ;;
    website/*)
      website_files+=("${path#website/}")
      ;;
    *.go)
      if ! is_generated_go "${path}"; then
        go_files+=("${path}")
      fi
      ;;
  esac
done

if [ "${#web_files[@]}" -gt 0 ]; then
  pnpm --dir web exec biome check --write --files-ignore-unknown=true \
    --no-errors-on-unmatched -- "${web_files[@]}"
fi

if [ "${#website_files[@]}" -gt 0 ]; then
  pnpm --dir website exec biome check --write --files-ignore-unknown=true \
    --no-errors-on-unmatched -- "${website_files[@]}"
fi

if [ "${#go_files[@]}" -gt 0 ]; then
  go tool gofumpt -l -w "${go_files[@]}"
  go tool goimports -local github.com/opcotech/elemo -w "${go_files[@]}"
fi
