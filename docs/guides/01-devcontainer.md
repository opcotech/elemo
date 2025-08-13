# Devcontainer Integration

Elemo provides a basic [devcontainer configuration](https://github.com/opcotech/elemo/blob/main/.devcontainer/devcontainer.json) that makes contributing easier. The configuration sets up a development environment completely by running the [`setup.sh`](https://github.com/opcotech/elemo/blob/main/scripts/setup.sh), which is the recommended way to setup the project whether in a devcontainer or not.

## Troubleshooting

Although `devcontainer`s can be awesome, they are not bulletproof and issues may happen.

> Q: The container setup shows an error saying the container is out of space during setup. What's wrong?

A: This is not us. It is literally the container running out of space while doing its own setup. Please do a (not full) rebuild or try with a devcontainer with bigger resource allocations.

> Q: Why are the ports is not accessible?

A: You may not running the services. Start backend services using `make start.backend` and web app using `make dev.frontend`.
