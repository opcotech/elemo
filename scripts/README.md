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
