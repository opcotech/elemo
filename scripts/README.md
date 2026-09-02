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

Wipes a running demo instance and reloads the ACME seed from
`assets/queries/demo.cypher`. Services are left running.

Clears Neo4j, Redis, Meilisearch, S3 objects, Postgres `user_tokens`,
`notifications`, and `oauth2_tokens`. Does **not** touch `oauth2_clients`, so
web `AUTH_CLIENT_ID` / `AUTH_CLIENT_SECRET` stay valid and you do not need to
re-register an OAuth client or restart the stack.

Requires the Compose stack to already be up (`make start.backend`). Pass
`--yes` because the wipe is destructive:

```sh
./scripts/reset-demo.sh --yes
# or
make demo.reset
```

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
./scripts/ort/ort.sh prepare   # clone pinned ort-config and overlay scripts/ort/
```

Results (evaluation JSON, SPDX, CycloneDX, WebApp HTML, NOTICE) go to `.ort/results/`, which is gitignored. The
`legal/` subdirectory (`LICENSE`, `LICENSE-COMMERCIAL`, `NOTICE`, SBOMs) is what GitHub Releases attach.
