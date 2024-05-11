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
- [Code Quality and Tests](#code-quality-and-tests)
- [Updating The APIs](#updating-the-apis)
- [Writing Commit Messages](#writing-commit-messages)
- [Code Review](#code-review)
- [Coding Style](#coding-style)
- [Web Component Design](#web-component-design)
- [Developer's Certificate of Origin](#developers-certificate-of-origin)

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

_Note: All contributions will be licensed under the project's license._

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
- **Update the CHANGELOG** for all enhancements and bug fixes. Include the corresponding issue number if one exists, and
  your GitHub username. (example: "- Fixed crash in profile view. #123 @jessesquires")
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

The project is using Makefile for the backend and standard pnpm tooling for the front-end. To start backend-related
services, execute `make start.backend`. In order to start the front-end, execute `pnpm dev` in the `web/` directory.

Below, you can find more useful make targets to run (`make <target>`):

```shell
build.backend                  # Build backend images
build.frontend                 # Build front-end app
build                          # Build backend and front-end
changelog                      # Update the changelog
clean                          # Destroys all backend resources and cleans up untracked files
dep.backend                    # Download backend dependencies
dep.frontend                   # Install front-end dependencies
dep                            # Download and install backend and front-end dependencies
destroy.backend                # Destroy all backend resources
dev.frontend                   # Start front-end for development
dev                            # Start backend and front-end for development
format.backend                 # Run formatters for the backend
format.frontend                # Run formatters for the front-end
format                         # Run formatters for the backend and front-end
generate.client                # Generate API client
generate.email                 # Generate HTML emails from MJML templates
generate.server                # Generate API server
generate                       # Generate resources
help                           # Show help message
lint.backend                   # Run linters for the backend
lint.frontend                  # Run linters for the front-end
lint                           # Run linters for the backend and front-end
release                        # Cut a new release
start.backend                  # Start backend services
start.frontend                 # Start front-end app
start                          # Start backend and front-end
stop.backend                   # Stop backend service
stop                           # Stop backend services
test.k6                        # Run k6 tests
test.backend.bench             # Run backend benchmarks
test.backend.coverage          # Combine unit and integration test coverage
test.backend.integration       # Run backend integration tests
test.backend.unit              # Run backend unit tests
test.backend                   # Run all backend tests
test.frontend                  # Run all front-end tests
test                           # Run all k6, backend and front-end tests
```

## Code Quality and Tests

The project ensures code quality and code coverage in multiple ways. Besides third-party online tools, with the lack of
completeness, `gofmt` `go-imports`, `golangci-lint`, `go test`, `k6`, `playwright` and `eslint` are used to keep up with
industry standards.

Although front-end unit tests are not created yet, linters and some end-to-end tests are available. In order to run
end-to-end tests, you have to have the necessary browser drivers installed. The easiest way to install them, is using
playwright. When the drivers are installed, you can start the end-to-end tests.

```shell
# Install playwright dependencies
pnpm --prefix web install
pnpm --prefix web exec playwright install

# Run linters
make lint           # Run linters for the backend and front-end, or
make lint.backend   # Run linters for the backend, or
make lint.frontend  # Run linters for the front-end

# Run tests
make test                     # Run all backend and front-end tests, or
make test.backend.bench       # Run backend benchmarks, or
make test.backend.coverage    # Combine unit and integration test coverage, or
make test.backend.integration # Run backend integration tests, or
make test.backend.unit        # Run backend unit tests, or
make test.backend             # Run all backend tests, or
make test.frontend            # Run all front-end tests
```

The external tests, such as load tests, smoke tests, stress tests, etc., are
defined in the `tests` directory. To run these tests, you need to install `k6`
first. Then, execute the following:

```shell
make test.k6 # Run k6 tests
```

## Updating The APIs

The APIs are defined in `/api/openapi/openapi.yaml`. To reduce the possibility
of human error and ensure the API is called properly, both the server and client
code is generated.

After updating the API specification, you have to regenerate the server and
client code. To do so, execute the following:

```shell
make generate.server # Generate API server
make generate.client # Generate API client
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

## Web Component Design

The project uses [Storybook](https://storybook.js.org/) for building UI components and pages in isolation. To run
Storybook locally, execute `pnpm storybook` in the `web` directory.

Then, navigate to the local Storybook instance: http://127.0.0.1:6006.

## Developer's Certificate of Origin

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I have the right to submit it under the open source
license indicated in the file; or

(b) The contribution is based upon previous work that, to the best of my knowledge, is covered under an appropriate open
source license and I have the right under that license to submit that work with modifications, whether created in whole
or in part by me, under the same open source license (unless I am permitted to submit under a different license), as
indicated in the file; or

(c) The contribution was provided directly to me by some other person who certified (a), (b) or (c) and I have not
modified it.

(d) I understand and agree that this project and the contribution are public and that a record of the contribution (
including all personal information I submit with it, including my sign-off) is maintained indefinitely and may be
redistributed consistent with this project or the open source license(s) involved.
