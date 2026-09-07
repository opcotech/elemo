# Contributing Guidelines

_Pull requests, bug reports, and all other forms of contribution are welcomed and highly encouraged!_

### Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Asking Questions](#asking-questions)
- [Types of Contributions](#types-of-contributions)
- [Opening an Issue](#opening-an-issue)
- [Feature Requests](#feature-requests)
- [Triaging Issues](#triaging-issues)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Running the services](#running-the-services)
- [Releases](#releases)
- [Code Quality and Tests](#code-quality-and-tests)
- [Updating The APIs](#updating-the-apis)
- [Writing Commit Messages](#writing-commit-messages)
- [Code Review](#code-review)
- [Coding Style](#coding-style)
- [Repository Conventions](#repository-conventions)
- [Web Component Design](#web-component-design)
- [Contributor License Agreement](#contributor-license-agreement)

> **This guide serves to set clear expectations for everyone involved with the project so that we can improve it
> together, while also creating a welcoming space for everyone to participate. Following these guidelines will help
> ensure a positive experience for contributors and maintainers.**

## Code of Conduct

Please review our [Code of Conduct](https://github.com/opcotech/elemo/blob/main/CODE_OF_CONDUCT.md). It is in effect at
all times. We expect it to be honored by everyone who contributes to this project. Acting like an asshole will not be
tolerated.

## Asking Questions

We utilize GitHub discussions as a place for our community to get together.
The [Q&A topic](https://github.com/opcotech/elemo/discussions/categories/q-a) serves as a dedicated place to ask
questions related to the project.

## Types of Contributions

There are many ways to contribute, but here are some examples to give you inspiration:

- Reporting issues
- Opening feature requests
- Fixing bugs
- Implementing new functionalities
- Participating in discussions
- Writing documentation

As you see, even a non-technical person can contribute. Use this opportunity to give back to the community by
contributing.

## Opening an Issue

Before [creating an issue](https://help.github.com/en/github/managing-your-work-on-github/creating-an-issue), check if
you are using the latest version of the project. If you are not up-to-date, see if updating fixes your issue first. When
opening a new issue, make sure you found no answers in the discussions or existing (open or closed) issues.

### Reporting Security Issues

Review our [Security Policy](https://github.com/opcotech/elemo/blob/main/SECURITY.md). **Do not** file a public issue
for or start a discussion about security vulnerabilities.

### Bug Reports and Other Issues

A great way to contribute to the project is to send a detailed issue when you encounter a problem. We always appreciate
a well-written, thorough bug report. Believe in Karma, open an issue that you would like to receive too.

- **Review
  the [documentation](https://github.com/opcotech/elemo/blob/main), [existing issues](https://github.com/opcotech/elemo/issues)
  and [discussions](https://github.com/opcotech/elemo/discussions)** before opening a new issue.
- **Do not open a duplicate issue!** Search through existing issues to see if your issue has previously been reported.
  If your issue exists, comment with any additional information you have. You may simply note "I have this problem too",
  which helps prioritize the most common problems and requests.
- **Prefer using [reactions](https://github.blog/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/)**, not
  comments, if you simply want to "+1" an existing issue.
- **Fully complete the provided issue template.** The bug report template requests all the information we need to
  quickly and efficiently address your issue. Be clear, concise, and descriptive. Provide as much information as you
  can, including steps to reproduce, stack traces, compiler errors, library versions, OS versions, and screenshots (if
  applicable).
- **Use [GitHub-flavored Markdown](https://help.github.com/en/github/writing-on-github/basic-writing-and-formatting-syntax).**
  Especially put code blocks and console outputs in backticks (`````````). This improves readability.
- **Provide as many _relevant_ details as you can**. This help the work of others and reduces the back-and-forth.

## Feature Requests

Feature requests are welcome! While we will consider all requests, we cannot guarantee your request will be accepted. We
want to avoid [feature creep](https://en.wikipedia.org/wiki/Feature_creep). Your idea may be great, but also
out-of-scope for the project. If accepted, we cannot make any commitments regarding the timeline for implementation and
release. However, you are welcome to submit a pull request to help!

- **Do not open a duplicate feature request.** Search for existing feature requests first. If you find your feature (or
  one very similar) previously requested, comment on that issue.
- **Fully complete the provided issue template.** The feature request template asks for all necessary information for us
  to begin a productive conversation.
- **Be precise about the proposed outcome** of the feature and how it relates to existing features. Include
  implementation details if possible.

## Triaging Issues

You can triage issues which may include reproducing bug reports or asking for additional information, such as version
numbers or reproduction instructions. Any help you can provide to quickly resolve an issue is very much appreciated!

## Submitting Pull Requests

We **love** pull requests!
Before [forking the repo](https://help.github.com/en/github/getting-started-with-github/fork-a-repo)
and [creating a pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/proposing-changes-to-your-work-with-pull-requests)
for non-trivial changes, it is usually best to
first [start a discussion](https://github.com/opcotech/elemo/discussions/categories/ideas) to discuss your intended
approach for solving the problem in the comments for an existing issue.

_Note: All contributions are subject to the [Contributor License Agreement](CLA.md)._

- **Smaller is better.** Submit **one** pull request per bug fix or feature. A pull request should contain isolated
  changes pertaining to a single bug fix or feature implementation. **Do not** refactor or reformat code that is
  unrelated to your change. It is better to **submit many small pull requests** rather than a single large one. Enormous
  pull requests will take enormous amounts of time to review, or may be rejected altogether.
- **Coordinate bigger changes.** For large and non-trivial changes, open an issue to discuss a strategy with the
  maintainers. Otherwise, you risk doing a lot of work for nothing!
- **Prioritize understanding over cleverness.** Write code clearly and concisely. Remember that source code usually gets
  written once and read often. Ensure the code is clear to the reader. The purpose and logic should be obvious to a
  reasonably skilled developer, otherwise you should add a comment that explains it.
- **Follow existing coding style and conventions.** Keep your code consistent with the style, formatting, and
  conventions in the rest of the code base. When possible, these will be enforced with a linter. Consistency makes it
  easier to review and modify in the future.
- **Include test coverage.** Add unit tests or UI tests when possible. Follow existing patterns for implementing tests.
- **Add documentation.** Document your changes with code doc comments or in existing guides.
- **Look up the existing [ADRs](https://adr.github.io/) before changing a questionable piece of code.** Code is
  opinionated and may not fit your preferred coding practices. However, almost everything have a good reason.
- **Sign the CLA.** Pull requests cannot be merged until you agree to the
  [Contributor License Agreement](CLA.md) (the bot will prompt you).
- **Use conventional commits.** The changelog is generated automatically from conventional commit messages by
  [Release Please](https://github.com/googleapis/release-please). Prefer `feat` and `fix` for user-facing changes;
  include the issue number in the commit body or PR description when one exists.
- **Use the repo's default branch.** Branch from
  and [submit your pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork)
  to the repo's default branch. This is the `main` branch.
- [Resolve any merge conflicts](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/resolving-a-merge-conflict-on-github)
  that occur.
- **Promptly address any CI failures**. If your pull request fails to build or pass tests, please push another commit to
  fix it.
- When writing comments, use properly constructed sentences, including punctuation.
- Although we are not all natives, please try your best to **provide documentation and comments in grammatically correct
  English**.

## Running the Services

The project uses [Mise](https://mise.jdx.dev/) for tool versions and a focused set of
developer tasks. Install Mise, then run `mise trust && mise bootstrap` in the
repository root so Go, Node, pnpm, and the other pinned tools are on your PATH and
project dependencies are installed. Bootstrap does not start services or modify
databases. When you want a running ACME demo, run `./scripts/dev-demo-init.sh --yes`,
then start Compose plus the frontend with `mise run dev`.

List every task with `mise tasks`. Conventional entry points first (`mise run <task>`):

```shell
build                          # Build backend image and frontend
test                           # Run backend and frontend correctness tests
check                          # Formatting checks, lint, typecheck, and unit tests
dev                            # Start Compose services and the frontend dev server
```

Selective build, test, quality, generation, and service tasks:

```shell
build-backend                  # Build backend images
build-frontend                 # Build front-end app
build-plugin                   # Build plugin zips (all, or one plugin by name)

test-backend                   # Run backend unit and integration tests
test-backend-unit              # Run backend unit tests
test-backend-integration       # Run backend integration tests
test-frontend                  # Run frontend unit and end-to-end tests
test-frontend-unit             # Run frontend unit tests
test-frontend-e2e              # Run frontend end-to-end tests
test-k6                        # Run k6 tests
test-storybook                 # Build Storybook and run a11y verification stories

lint                           # Lint backend and frontend
lint-backend                   # Lint backend
lint-frontend                  # Lint the web app and website
lint-web                       # Lint the web app
lint-website                   # Lint the website
format                         # Format backend and frontend
format-backend                 # Format backend
format-frontend                # Format the web app and website
format-web                     # Format the web app
format-website                 # Format the website
format-check                   # Check backend and frontend formatting
format-check-backend           # Check backend formatting
format-check-frontend          # Check web app and website formatting
format-check-web               # Check web app formatting
format-check-website           # Check website formatting

generate                       # Generate OpenAPI, server, client, and email artifacts
generate-openapi               # Assemble OpenAPI spec from split sources
generate-server                # Generate API server
generate-client                # Generate API client
generate-client-check          # Fail if generated TypeScript client drifted
generate-email                 # Generate HTML emails from MJML templates

start                          # Start Compose services
stop                           # Stop Compose services
destroy                        # Stop Compose services and remove local images and volumes
monitoring-start               # Start Jaeger/Prometheus/Grafana monitoring stack
monitoring-stop                # Stop Jaeger/Prometheus/Grafana monitoring stack
```

Dependency installation, environment setup, demo initialization, license scanning,
coverage combination, typecheck, and benchmarks are not Mise tasks. Use the native
commands:

```shell
mise bootstrap
./scripts/dev-demo-init.sh --yes
pnpm --dir web install
pnpm --dir website install
pnpm --dir tools/email install
go mod download
pnpm --dir web typecheck
./scripts/dev-demo-reset.sh --yes
go run ./tools/workload-prefill -config configs/development/config.local.gen.yml -yes
./scripts/ort/ort.sh pr
./scripts/ort/ort.sh run
./.github/scripts/combine-backend-coverage.sh
go test -run=Bench -bench=. -benchmem -benchtime=10s ./...
```

Former Makefile targets that are not Mise tasks:

- `generate.backend.check` — run `mise run generate-server`, then
  `git diff --exit-code` and `mise run format-check-backend`.
- `test.k6.work-item-latency` — start the stack and run k6 with
  `tests/k6/config/work-item-latency.json` (see [`tests/k6/README.md`](tests/k6/README.md)).
- `clean` — `mise run stop` or `mise run destroy`, and delete local
  coverage files (`.coverage.*.out`) as needed.
- `demo.prefill` — `go run ./tools/workload-prefill` as listed above.

## Releases

Releases are automated with [Release Please](https://github.com/googleapis/release-please). Conventional commits on
`main` drive a release pull request that updates `CHANGELOG.md` and package versions. Merging that PR creates a draft
GitHub Release and tag. The release workflow then runs ORT on the tag and attaches a **legal bundle** (`LICENSE`,
`LICENSE-COMMERCIAL`, generated `NOTICE`, SPDX and CycloneDX SBOMs). NOTICE and SBOMs are generated, not committed.
Publish the draft when ready.

## Code Quality and Tests

The project ensures code quality and code coverage in multiple ways. Besides third-party online tools, with the lack of
completeness, `gofumpt`, `goimports`, `golangci-lint`, `gotestsum`, `k6`, `playwright` and `biome` are used to keep up with
industry standards. Backend formatters are Go tools (`go tool gofumpt`, `go tool goimports`) declared in `go.mod`.

License and dependency policy is enforced with [ORT](https://oss-review-toolkit.org/ort/). Pull request CI runs
`./scripts/ort/ort.sh pr` (analyzer + evaluator) so dependency-license policy is gated quickly. Pushes to `main`, nightly
runs, and releases run `./scripts/ort/ort.sh run`, which ScanCodes **Elemo source** (not every dependency), concludes Go licenses
from `go-licenses`, and produces reports. Evaluation is against Apache-2.0 (the future license of FSL-1.1-ALv2).
It fails on strong copyleft inbound and restrictive, unknown, or unlicensed dependencies. Full runs also fail on
copyleft in project source. Reports land in `.ort/results/` (gitignored), including `.ort/results/legal/` for
GitHub Release assets.

Although front-end unit tests exist (`mise run test-frontend-unit`), linters and
end-to-end tests are also available. In order to run end-to-end tests, you have
to have the necessary browser drivers installed. The easiest way to install them
is using playwright. When the drivers are installed, you can start the
end-to-end tests.

```shell
# Install playwright dependencies
pnpm --dir web install
pnpm --dir web exec playwright install --with-deps

# Fast quality gate (formatting, lint, typecheck, unit tests)
mise run check

# Run linters
mise run lint            # Run linters for the backend and front-end, or
mise run lint-backend    # Run linters for the backend, or
mise run lint-frontend   # Run linters for the front-end

# License / dependency policy (requires Docker)
./scripts/ort/ort.sh pr              # Analyze and evaluate (PR CI policy gate; no ScanCode)
./scripts/ort/ort.sh run             # Analyze, scan Elemo source, evaluate, advise, and report
./scripts/ort/ort.sh prepare         # Fetch pinned ORT config (needed once, or after pin bumps)
./scripts/ort/ort.sh analyze         # Dependency analysis only
./scripts/ort/ort.sh scan            # ScanCode on Elemo projects (full CI default)
ORT_SCAN_PACKAGE_TYPES=PACKAGE,PROJECT ./scripts/ort/ort.sh scan  # ScanCode on every dependency (slow; not CI)
./scripts/ort/ort.sh evaluate        # Policy evaluation (fails on ERROR violations)
./scripts/ort/ort.sh report          # SPDX, CycloneDX, WebApp, NOTICE, and legal/ bundle

# Run tests
mise run test                        # Run all backend and front-end correctness tests, or
mise run test-backend-integration    # Run backend integration tests, or
mise run test-backend-unit           # Run backend unit tests, or
mise run test-backend                # Run all backend tests, or
mise run test-frontend-unit          # Run front-end unit tests, or
mise run test-frontend-e2e           # Run front-end end-to-end tests, or
mise run test-frontend               # Run all front-end tests
go test -run=Bench -bench=. -benchmem -benchtime=10s ./...  # Backend benchmarks
./.github/scripts/combine-backend-coverage.sh                       # Combine unit and integration coverage
```

The external tests, such as load tests, smoke tests, stress tests, etc., are
defined in the `tests/k6` directory. k6 is installed by Mise. Then, execute the following:

```shell
mise run test-k6 # Run k6 tests
```

## Updating The APIs

The APIs are defined as YAML fragments in `/api/openapi/src/`. Those files are
assembled into `/api/openapi/openapi.yaml` (do not edit the bundle by hand).
To reduce the possibility of human error and ensure the API is called properly,
both the server and client code is generated.

After updating the API specification, you have to assemble the spec, regenerate
the server and client code, then confirm there is no unexpected drift:

```shell
mise run generate-openapi         # Assemble api/openapi/openapi.yaml from src/
mise run generate-server          # Assemble, then generate API server and Go mocks/enums
mise run generate-client          # Assemble, then generate API client
mise run generate-client-check    # Fail if generated TypeScript client changed
```

## Writing Commit Messages

Please [write a great commit message](https://chris.beams.io/posts/git-commit/).

1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Do not capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line (example: "Fix networking issue")
6. Wrap the body at about 72 characters
7. Use the body to explain **why**, _not what and how_ (the code shows that!)

We use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) format, which is enforced by linters. An
example excellent commit could look like this:

```
fix: short summary of changes in 50 chars or less in total

Add a more detailed explanation here, if necessary. Possibly give
some background about the issue being fixed, etc. The body of the
commit message can be several paragraphs. Further paragraphs come
after blank lines and please do proper word-wrap.

Wrap it to about 72 characters or so. In some contexts,
the first line is treated as the subject of the commit and the
rest of the text as the body. The blank line separating the summary
from the body is critical (unless you omit the body entirely);
various tools like `log`, `shortlog` and `rebase` can get confused
if you run the two together.

Explain the problem that this commit is solving. Focus on why you
are making this change as opposed to how or what. The code explains
how or what. Reviewers and your future self can read the patch,
but might not understand why a particular solution was implemented.
Are there side effects or other unintuitive consequences of this
change? Here's the place to explain them.

 - Bullet points are okay, too
 - A hyphen or asterisk should be used for the bullet, preceded
   by a single space, with blank lines in between

Note the fixed or relevant GitHub issues at the end:

Resolves: #123
See also: #456, #789
```

## Code Review

- **Self-review your code before submitting it.** This helps to reduce review cycles. Also, you may find something that
  you would change after all.
- **Read relevant review guides** before requesting a review and adjust your code if
  necessary. [Code review guide](https://github.com/golang/go/wiki/CodeReviewComments) and
  [concurrency review guide](https://github.com/golang/go/wiki/CodeReviewConcurrency) are excellent resources.
- **Review the code, not the author.** Look for and suggest improvements without disparaging or insulting the author.
  Provide **actionable feedback** and explain your reasoning.
- **You are not your code.** When your code is critiqued, questioned, or constructively criticized, remember that you
  are not your code. Do not take code review personally.
- **Always do your best.** No one writes bugs on purpose. Do your best, and learn from your mistakes.
- Kindly note any violations to the guidelines specified in this document.

## Coding Style

Consistency is the most important. Following the existing style, formatting, and naming conventions of the file you are
modifying and of the overall project. Failure to do so will result in a prolonged review process that has to focus on
updating the superficial aspects of your code, rather than improving its functionality and performance.

When possible, style and format will be enforced with a linter.

## Repository Conventions

These conventions are Elemo-specific. Follow them for new code, and prefer them
when touching existing code.

### Backend

- Keep the partial-hexagonal layout: models own domain types, repositories own
  persistence, services own application logic, and HTTP controllers map requests
  to commands. See [ADR 0001](docs/ADRs/0001.software-architecture.md),
  [ADR 0020](docs/ADRs/0020.query-projections-and-relationship-fetching.md), and
  [ADR 0021](docs/ADRs/0021.scoped-rebac-authorization.md).
- Wire dependencies manually. Service constructors take required collaborators
  as arguments. Use options only for logger and tracer (`WithLogger`,
  `WithTracer`).
- Repository reads use typed query structs, `Compile()`, and
  `Neo4jExecuteReadPlan`. Name scoped lists `ListFor{Scope}` (`ListForLibrary`,
  `ListForNamespace`, `ListForOrganization`, `ListForUser`). Look up a single
  entity with `Get` or `GetByKey`. PostgreSQL and S3 repositories stay
  store-specific.
- Context carries request metadata (authenticated user, tracing). Filters,
  sorting, and list options are explicit parameters, not context values.
- Service errors are `Err{Entity}{Operation}` (`List`, not `GetAll`). Wrap the
  operation sentinel first, then inspect with `errors.Is` / `errors.As`. HTTP
  handlers classify through `classifyServiceError`.
- Do not import generated OpenAPI types from `internal/transport/http/api`
  outside the HTTP layer except through the existing mapping.

### Frontend

- App code imports API artifacts through `@/lib/api/{types,sdk,query-options,mutation-options,schemas}`.
  Do not import `@/lib/client` from application or component code.
- `Issue` is the API type. `WorkItem` is the work UI view model. The UI entity
  key is `work-item`. Convert at `web/src/lib/work/issue-adapter.ts`.
- Data routes load with `queryClient.fetchQuery` and wrap loaders in
  `withRouteErrors`. Settings tables use `SettingsResourceTable`. Standard forms
  use generated Zod schemas, `useFormMutation`, and `DialogForm`. Destructive
  flows use `useDeleteMutation` or entity lifecycle.
- Optimistic patching is reserved for high-frequency issue field updates.
  Documents invalidate queries instead. Mock data remains only for domains
  without a product API.

### Tests and generation

- Place unit tests next to the package (`*_test.go`). Integration tests live in
  `*_integration_test.go` and require the testcontainers environment.
- Regenerate mocks, enumer output, and the OpenAPI server with
  `mise run generate-server`. Check drift with `mise run generate-server` followed
  by `git diff --exit-code`, `mise run format-check-backend`, and
  `mise run generate-client-check`. The front-end client uses
  `mise run generate-client` / `generate-client-check`.
- Front-end unit tests: `mise run test-frontend-unit`. End-to-end tests:
  `mise run test-frontend-e2e`.

## Web Component Design

The project uses [Storybook](https://storybook.js.org/) for building UI components and pages in isolation. To run
Storybook locally, execute `pnpm storybook` in the `web` directory.

Then, navigate to the local Storybook instance: http://127.0.0.1:6006.

## Contributor License Agreement

Contributions are accepted under the [Contributor License Agreement](CLA.md). That agreement lets Open Code
Technologies FZC license your contribution under FSL-1.1-ALv2 and under commercial terms. Signing also
grants those licenses for your prior Elemo contributions.

The CLA bot comments on pull requests with the exact reply needed to sign. Signatures are committed
onto the PR branch as [`.github/cla.json`](.github/cla.json) (main cannot be pushed to directly). The
bot only records PR committers; to import other matching comments, run
`./.github/scripts/cla-sync-signatures.py <pr> --push`. Corporate-owned work also requires the Corporate CLA
section of that document.
