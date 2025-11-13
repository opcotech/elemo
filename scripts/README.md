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
