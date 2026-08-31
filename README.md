[![ORT](https://img.shields.io/badge/compliance-ORT-blue)](https://oss-review-toolkit.org/ort/)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8801/badge)](https://www.bestpractices.dev/projects/8801)
[![Build](https://github.com/opcotech/elemo/actions/workflows/build.yml/badge.svg)](https://github.com/opcotech/elemo/actions/workflows/build.yml)
[![Codecov](https://codecov.io/gh/opcotech/elemo/graph/badge.svg?token=1E0JG98ESD)](https://codecov.io/gh/opcotech/elemo)

<br />
<div align="center">
  <h3 align="center">Elemo</h3>

  <p align="center">
    The next-generation project management platform.
    <br />
    <a href="https://github.com/opcotech/elemo/tree/main/docs"><strong>Explore the docs</strong></a>
    ·
    <a href="https://github.com/opcotech/elemo/blob/main/CONTRIBUTING.md#web-component-design"><strong>Check Storybook</strong></a>
    <br />
    <br />
    Join our <a href="https://discord.gg/sx9FPyXAdP">Discord</a> or <a href="https://join.slack.com/t/elemo-workspace/shared_invite/zt-3a6w9jb46-4uGjtkcqBN9BqBD50xl8eA">Slack</a>
    ·
    <a href="https://github.com/opcotech/elemo/issues/new?assignees=&labels=bug%2Ctriage-needed&projects=&template=BUG-REPORT.yml">Bug report</a>
    ·
    <a href="https://github.com/opcotech/elemo/issues/new?assignees=&labels=question%2Cenhancement%2Ctriage-needed&projects=&template=FEATURE-REQUEST.yml">Feature request</a>
  </p>
</div>

## About

Elemo is a project management platform which aims to help developers to ship faster, project managers to have better
project overview, and community members to be involved in the lifecycle of projects.

Elemo is not trying to reinvent the wheel, though it is introducing new abstractions in project management to allow any
size of company an easy use.

## Try Elemo

Setting up the development environment is an easy and straightforward process, however, you will need to run the code
on Linux, MacOS, or Windows WSL2. Follow the steps below and get ready to contribute:

1. Clone the repository

   ```shell
   # Clone the repository and change directory
   git clone https://github.com/opcotech/elemo.git && cd elemo
   ```

2. Set up and configure the services using an automated setup script

   ```shell
   # Make sure you have all the development requirements installed, then run the setup script.
   # Requirements: yq, jq, go, openssl, docker (with compose plugin), make, nvm, gotestsum
   ./scripts/setup.sh
   ```

3. Start the services

   ```shell
   # Start the backend services
   make start # or "make dev" for development
   ```

4. Now, you can try Elemo. Navigate to http://127.0.0.1:3000 and log in using the `demo@elemo.example` email
   and `AppleTree123` password.

## Contributing

We welcome contributions to the project, whether it is source code, documentation, bug reports, feature requests or
feedback. To get started with contributing:

- Have a look through GitHub issues labelled "good first issue".
- Read the [contributing guide](https://github.com/opcotech/elemo/blob/main/CONTRIBUTING.md).
- For details on building Elemo, see the
  related [Dockerfile](https://github.com/opcotech/elemo/blob/main/build/package/Dockerfile).
- Create a fork of Elemo and submit a pull request with your proposed changes.

## Codespaces/Dev Containers

Configuration for codespaces/dev containers are available. For more information check out the related [documentation](https://github.com/opcotech/elemo/blob/main/docs/guides/01-devcontainer.md).

Plugin authors should start with the [plugin developer guide](docs/guides/plugins.md).

## License

Elemo is **source-available** under [FSL-1.1-ALv2](LICENSE):

- You may read, modify, and self-host Elemo for **internal** use.
- You may not offer Elemo (or substantially similar functionality) to others as a **competing commercial product or hosted service** for two years after each version is published. After that, that version becomes Apache 2.0.
- **Elemo Cloud** is Licensor’s hosted service and is governed by its own terms, not by FSL.

A [commercial license](LICENSE-COMMERCIAL) is available if you need Competing Use rights (OEM, white-label, or third-party hosting). Contact **info@opcotech.com**.

Git tags published before this FSL switch remain available under AGPL-3.0 as originally licensed.
