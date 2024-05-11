[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8801/badge)](https://www.bestpractices.dev/projects/8801)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopcotech%2Felemo.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopcotech%2Felemo?ref=badge_shield&issueType=license)
[![Backend Build](https://github.com/opcotech/elemo/actions/workflows/build-backend.yml/badge.svg)](https://github.com/opcotech/elemo/actions/workflows/build-backend.yml)
[![Front-end Build](https://github.com/opcotech/elemo/actions/workflows/build-frontend.yml/badge.svg)](https://github.com/opcotech/elemo/actions/workflows/build-frontend.yml)
[![Maintainability](https://api.codeclimate.com/v1/badges/9164ce433e407e447946/maintainability)](https://codeclimate.com/github/opcotech/elemo/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/9164ce433e407e447946/test_coverage)](https://codeclimate.com/github/opcotech/elemo/test_coverage)

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
    <a href="https://discord.gg/sx9FPyXAdP">Join our Discord</a>
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

## Features

_The features listed below are part of the roadmap, but not necessarily implemented yet. The __fully__ implemented features are
marked with a checkmark._

- [ ] **Organizations:** Collaborate on projects across multiple organizations, whether it is your client, vendor, your
      subsidiary, or else.
- [x] **Roles:** Create roles for organizations, namespaces, or projects. Be flexible. You decide on what permissions
      the role has.
- [ ] **Namespaces:** Organize your projects into namespaces, create namespace-specific roles and forget about
      team-named projects as a workaround.
- [ ] **Projects:** Group issues and documents into projects and keep everything related at one place. No more
      unnecessary back-and-forth between tabs.
- [ ] **Issues:** Create issues, attach related files, link documents, and more. Everything you need for the
      implementation at one place.
- [ ] **Documents:** Create rich documents, link them to issues, projects, or even namespaces.
- [x] **Todo lists:** Track your Todo list within Elemo and create new items in the blink of an eye. No more "where did
      I put that note?!".
- [ ] **Boards:** Gain insights into the project's progress. No matter what project management methodology your team
      uses.
- [ ] **Releases:** Keep your releases where they belong to. Link releases to documents, issues, and so on.
- [ ] **Roadmaps:** Make sure you see the big picture and don't get lost in details.

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
   # Requirements: yq, jq, go, openssl, docker (with compose plugin), make, nvm
   ./scripts/setup.sh
   ```

3. Start the services

   ```shell
   # Start the backend services
   make start # or "make dev" for development
   ```

4. Now, you can try Elemo. Navigate to http://127.0.0.1:3000 and log in using the `demo@elemo.app` email
   and `AppleTree123` password.

## Contributing

We welcome contributions to the project, whether it is source code, documentation, bug reports, feature requests or
feedback. To get started with contributing:

- Have a look through GitHub issues labelled "good first issue".
- Read the [contributing guide](https://github.com/opcotech/elemo/blob/main/CONTRIBUTING.md).
- For details on building Elemo, see the
  related [Dockerfile](https://github.com/opcotech/elemo/blob/main/build/package/Dockerfile).
- Create a fork of Elemo and submit a pull request with your proposed changes.
