# Scripts

Helper scripts used for building and/or testing the project.

## extract-and-lint-licenses.sh

This script extracts the licenses of all dependencies and checks if they are
approved and compatible with each other. The script requires `jq`, `yq`,
`flict`, and `license_finder`.

## generate-dev-config.sh

This script generates development configuration files and key. It requires
`mkcert`, `openssl`, and `go`
