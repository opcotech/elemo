# Scripts

Helper scripts used for building and/or testing the project.

## bug-report.sh

Collects some useful information for debugging and prints them on the screen.

## common.sh

The collection of common exports and functions, used by other scripts.

## generate-dev-config.sh

This script generates development configuration files and key.

## generate-frontend-client.sh

Takes the Open API specification in the `api/openapi` directory, and generates
a TypeScript client from it.

## reset-demo.sh

Wipes a running demo instance and reloads the smoke-profile seed via
`tools/workload-prefill`. Data stores stay up; `elemo-server` is restarted so
the plugin registry is empty.

Requires the Compose stack to already be up (`make start.backend`). Pass
`--yes` because the wipe is destructive:

```sh
./scripts/reset-demo.sh --yes
# or
make demo.reset
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

## setup.sh

This script prepares the whole development environment. Generates a new
developer configuration, creates certificates, register a new OAuth client,
sets the web credentials, creates a new user in the database to interact with.

The following environment variables customize CI setup while preserving the
default local behavior:

- `ELEMO_SKIP_IMAGE_BUILD=true` uses an existing local `elemo-server` image.
- `ELEMO_KEEP_BACKEND=true` leaves the Compose services running after setup.
- `PLAYWRIGHT_BROWSERS` is a space-separated list of browsers to install. When
  unset, Playwright installs all configured browsers.

For example, a CI job that has already loaded the server image can run:

```sh
ELEMO_SKIP_IMAGE_BUILD=true \
ELEMO_KEEP_BACKEND=true \
PLAYWRIGHT_BROWSERS=chromium \
./scripts/setup.sh
```

## cla-sync-signatures.py

Imports CLA sign-off comments from a pull request into `.github/cla.json` on **that PR's
branch**. The CLA Action only records PR committers; this backfills anyone else who posted
`I have read the CLA Document and I hereby sign the CLA`.

```sh
./scripts/cla-sync-signatures.py 425           # print merged JSON; do not push
./scripts/cla-sync-signatures.py 425 --push    # commit new signers onto the PR branch
```

## ort/

ORT runner, policy overlay, and Go license curation generator. See [`scripts/ort/ort.sh`](ort/ort.sh).

```sh
./scripts/ort/ort.sh           # prepare + analyze + scan + evaluate + advise + report
make ort                       # same
make ort.pr                    # prepare + analyze + evaluate (PR CI; no ScanCode)
./scripts/ort/ort.sh prepare   # clone pinned ort-config and overlay scripts/ort/
```

Results (evaluation JSON, SPDX, CycloneDX, WebApp HTML, NOTICE) go to `.ort/results/`, which is gitignored. The
`legal/` subdirectory (`LICENSE`, `LICENSE-COMMERCIAL`, `NOTICE`, SBOMs) is what GitHub Releases attach.
