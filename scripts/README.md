# Scripts

Helper scripts used for building, generating, and supporting the project.
GitHub-owned helpers live in [`.github/scripts/`](../.github/scripts).

## common.sh

Shared exports and functions used by other scripts, including `dockerCompose`
which locates the Compose CLI plugin when `docker compose` is not on PATH.

## compose.sh

Runs Docker Compose through `dockerCompose` in `common.sh`. Mise tasks and
local scripts use this so Compose works when Docker Desktop has not registered
the plugin under `~/.docker/cli-plugins`.

```sh
./scripts/compose.sh version
./scripts/compose.sh -f deploy/docker/docker-compose.yml config
mise run start
```

Override the plugin path with `DOCKER_COMPOSE_PLUGIN` if Compose is installed
somewhere else.

## bootstrap.sh

Idempotent toolchain and dependency setup. Installs locked Mise tools from
this project's `mise.lock` (not tools declared in a user-global Mise config),
Go
modules, frontend and email dependencies, generates missing local config and
certificates, generates the frontend API client, and installs Playwright
browsers. It does not start services or modify databases.

```sh
mise bootstrap
./scripts/bootstrap.sh
```

`PLAYWRIGHT_BROWSERS` is a space-separated list of browsers to install. When
unset, Playwright installs all configured browsers.

## dev-demo-init.sh

Starts Compose services, wipes graph data, loads the ACME demo seed, mints
OAuth credentials into `web/.env` and `web/.env.test.local`, and leaves the
stack running. This is destructive and requires `--yes`.

```sh
./scripts/dev-demo-init.sh --yes
```

`ELEMO_SKIP_IMAGE_BUILD=true` uses an existing local `elemo-server` image.

## generate-dev-config.sh

Generates development configuration files and keys.

## dev-demo-reset.sh

Wipes a running demo instance and reloads the smoke-profile seed via
`tools/workload-prefill`. Data stores stay up; `elemo-server` is restarted so
the plugin registry is empty.

Requires the Compose stack to already be up (`mise run start`). Pass
`--yes` because the wipe is destructive:

```sh
./scripts/dev-demo-reset.sh --yes
```

Local development uses the Compose file in-tree, `neo4jsecret` / `pgsecret`,
and host `go`. Production (unpublished DB ports, vaulted passwords, no Go)
sets:

| Variable | Purpose |
| --- | --- |
| `ELEMO_COMPOSE_OVERRIDE` | Extra Compose file (vaulted auth, no published DB ports) |
| `ELEMO_COMPOSE_ENV_FILE` | Compose `--env-file` |
| `ELEMO_PREFILL_CONFIG` | `config.yml` with Docker DNS names (`neo4j`, `postgres`, …) |
| `ELEMO_CONFIGS_DIR` | Host directory mounted at `/src/configs` so relative license paths resolve |
| `ELEMO_PREFILL_DOCKER=1` | Run `workload-prefill` in `ELEMO_GOLANG_IMAGE` on `ELEMO_COMPOSE_NETWORK` |
| `NEO4J_PASSWORD` / `NEO4J_AUTH` | cypher-shell auth (defaults to `neo4jsecret`) |
| `POSTGRES_PASSWORD` | psql auth (defaults to `pgsecret`) |

`--with-oauth` remints `bin/elemo auth add-client` after the reset (needs
`ELEMO_OAUTH_CALLBACK_URL`, `ELEMO_OAUTH_SECRETS_FILE`, `ELEMO_WEB_ENV`,
`ELEMO_API_BASE_URL`). Set `ELEMO_WEB_SERVICE` to restart a systemd unit.

## build-plugin.sh

Builds a plugin zip (WASM backend + Vite frontend) into `dist/plugins/`. With
no arguments, builds every plugin under `plugins/`. Pass a directory name or
path to build one:

```sh
./scripts/build-plugin.sh
./scripts/build-plugin.sh timetracking
mise run build-plugin -- accounting
```

## generate-frontend-client.sh

Takes the OpenAPI specification in the `api/openapi` directory and generates
a TypeScript client from it.

## check-frontend-client-drift.sh

Regenerates the TypeScript client and fails if `web/src/lib/client` changed.
Used by `pnpm --dir web generate:check`.

## ort/

ORT runner, policy overlay, and Go license curation generator. See
[`scripts/ort/ort.sh`](ort/ort.sh).

```sh
./scripts/ort/ort.sh           # prepare + analyze + scan + evaluate + advise + report
./scripts/ort/ort.sh pr        # prepare + analyze + evaluate (PR CI; no ScanCode)
./scripts/ort/ort.sh prepare   # clone pinned ort-config and overlay scripts/ort/
```

Results (evaluation JSON, SPDX, CycloneDX, WebApp HTML, NOTICE) go to
`.ort/results/`, which is gitignored. The `legal/` subdirectory (`LICENSE`,
`LICENSE-COMMERCIAL`, `NOTICE`, SBOMs) is what GitHub Releases attach.

## GitHub-owned helpers

These live in [`.github/scripts/`](../.github/scripts) because they are
consumed by GitHub workflows and PR process, not local development:

- `combine-backend-coverage.sh` merges `.coverage.unit.out` and
  `.coverage.integration.out` into `.coverage.out`
- `wait-for-elemo.sh` waits until the API heartbeat endpoint responds
- `cla-sync-signatures.py` imports CLA sign-off comments from a pull request
  into `.github/cla.json` on that PR's branch

```sh
./.github/scripts/cla-sync-signatures.py 425           # print merged JSON; do not push
./.github/scripts/cla-sync-signatures.py 425 --push    # commit new signers onto the PR branch
```
