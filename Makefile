.DEFAULT_GOAL := build

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR=$(ROOT_DIR)/build
TEMPLATES_DIR=$(ROOT_DIR)/templates
API_DIR:=$(ROOT_DIR)/api/openapi
API_SERVER_DIR:=$(ROOT_DIR)/internal/transport/http/api
EMAILS_DIR:=$(BUILD_DIR)/email

FRONTEND_DIR:=$(ROOT_DIR)/web
FRONTEND_CLIENT:=elemo-client
FRONTEND_CLIENT_DIR:=$(ROOT_DIR)/web/packages/$(FRONTEND_CLIENT)

BACKEND_COVER_OUT := $(ROOT_DIR)/.coverage.out
BACKEND_COVER_OUT_UNIT := $(ROOT_DIR)/.coverage.unit.out
BACKEND_COVER_OUT_INTEGRATION := $(ROOT_DIR)/.coverage.integration.out

PNPM_EXEC := $(shell which pnpm)
PNPM_RUN := $(PNPM_EXEC) run --prefix $(FRONTEND_DIR)
PNPM_EMAILS_RUN := $(PNPM_EXEC) run --prefix $(EMAILS_DIR)

GO_EXEC := $(shell which go)
GO_TEST_COVER := $(GO_EXEC) test -race -shuffle=on -cover -covermode=atomic -ldflags="-extldflags=-Wl,-ld_classic"
GO_TEST_IGNORE := "(mode: atomic|testutil|tools|cmd|http\/api)"

TMPDIR := $(shell echo "${TMPDIR:-/tmp}")

define log
	@echo "[\033[36mINFO\033[0m]\t$(1)" 1>&2;
endef

.PHONY: help
help: ## Show help message
	@echo "Available targets:";
	@grep -E '^[a-z.A-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}';

.PHONY: changelog
changelog: ## Update the changelog
	$(call log, updating changelog)

	@if [ -z "$(RELEASE_VERSION)" ]; then \
		git cliff > CHANGELOG.md; \
	else \
		git cliff --tag "v$(RELEASE_VERSION)" --unreleased --prepend CHANGELOG.md; \
	fi

.PHONY: release
release: ## Cut a new release
	$(if $(value RELEASE_VERSION),,$(error No RELEASE_VERSION set))

	$(call log, bumping front-end version)
	@jq '.version="$(RELEASE_VERSION)"' $(FRONTEND_CLIENT_DIR)/package.json > $(TMPDIR)/package.json.tmp && \
		mv $(TMPDIR)/package.json.tmp $(FRONTEND_CLIENT_DIR)/package.json;
	@jq '.version="$(RELEASE_VERSION)"' $(FRONTEND_DIR)/package.json > $(TMPDIR)/package.json.tmp && \
		mv $(TMPDIR)/package.json.tmp $(FRONTEND_DIR)/package.json;
	@$(PNPM_EXEC) update --prefix $(FRONTEND_CLIENT)

	@$(MAKE) changelog;
	
	$(call log, committing changelog)
	@git commit -sm "chore(changelog): update changelog for v$(RELEASE_VERSION)"

	$(call log, cutting new tag)
	@git tag -sm "chore(release): v$(RELEASE_VERSION)"

.PHONY: generate
generate: generate.email generate.server generate.client ## Generate resources

.PHONY: generate.server
generate.server: ## Generate API server
	$(call log, generating backend API server)
	@oapi-codegen -config $(API_DIR)/generator.config.yml -o $(API_SERVER_DIR)/server.go $(API_DIR)/openapi.yaml

.PHONY: generate.client
generate.client: ## Generate API client
	$(call log, generating front-end API client)
	@$(PNPM_RUN) generate 2>&1 >/dev/null

.PHONY: generate.email
generate.email: ## Generate HTML template emails
	# TODO: when deployed to production, we should use the actual S3 bucket and endpoint
	$(call log, compiling email templates)
	@$(PNPM_EMAILS_RUN) build --out $(TEMPLATES_DIR)/email \
		--access-key-id "access-key-id" \
		--secret-access-key "secret-access-key" \
		--region "us-east-1" \
		--s3-bucket elemo \
		--endpoint "http://0.0.0.0:4566" \
		--static-root "http://0.0.0.0:4566/elemo"

.PHONY: dep
dep: deb.backend dep.frontend ## Download and install backend and front-end dependencies

.PHONY: dep.backend
dep.backend: ## Download backend dependencies
	$(call log, download backend dependencies)
	@$(GO_EXEC) mod tidy
	@$(GO_EXEC) mod download

.PHONY: dep.frontend
dep.frontend: ## Install front-end dependencies
	$(call log, download and install front-end dependencies)
	@rm -rf $(FRONTEND_DIR)/node_modules
	@$(PNPM_EXEC) install --prefix $(FRONTEND_DIR)

.PHONY: build
build: build.backend build.frontend ## Build backend and front-end

.PHONY: build.backend
build.backend: ## Build backend images
	$(call log, build backend images)
	@docker compose -f deploy/docker/docker-compose.yml build --no-cache

.PHONY: build.frontend
build.frontend: ## Build front-end app
	$(call log, build front-end app)
	@$(PNPM_RUN) build

.PHONY: dev
dev: start.backend dev.frontend ## Start backend and front-end for development

.PHONY: dev.frontend
dev.frontend: dep.frontend ## Start front-end for development
	$(call log, starting front-end app)
	@$(PNPM_RUN) dev

.PHONY: start
start: start.backend start.frontend ## Start backend and front-end

.PHONY: start.backend
start.backend: ## Start backend services
	$(call log, starting backend services)
	@docker compose -f deploy/docker/docker-compose.yml up -d --force-recreate

.PHONY: start.frontend
start.frontend: build.frontend ## Start front-end app
	$(call log, starting front-end app)
	@$(PNPM_RUN) start

.PHONY: stop
stop: stop.backend ## Stop backend services

.PHONY: stop.backend
stop.backend: ## Stop backend service
	$(call log, stopping backend services)
	@docker compose -f deploy/docker/docker-compose.yml stop
	
.PHONY: test
test: test.backend test.frontend test.k6 ## Run all k6, backend and front-end tests

.PHONY: test.backend
test.backend: test.backend.unit test.backend.integration test.backend.bench test.backend.coverage ## Run all backend tests

.PHONY: test.backend.bench
test.backend.bench: ## Run backend benchmarks
	$(call log, execute backend benchmarks)
	@$(GO_EXEC) test -run=Bench -bench=. -benchmem -benchtime=10s ./...

.PHONY: test.backend.unit
test.backend.unit: ## Run backend unit tests
	$(call log, execute backend unit tests)
	@rm -f $(BACKEND_COVER_OUT_UNIT)
	@$(GO_TEST_COVER) -short -coverprofile=$(BACKEND_COVER_OUT_UNIT) ./...

.PHONY: test.backend.integration
test.backend.integration: ## Run backend integration tests
	$(call log, execute backend integration tests)
	@rm -f $(BACKEND_COVER_OUT_INTEGRATION)
	@$(GO_TEST_COVER) -timeout 900s -run=Integration -coverprofile=$(BACKEND_COVER_OUT_INTEGRATION) ./...

.PHONY: test.backend.coverage
test.backend.coverage: ## Combine unit and integration test coverage
	$(call log, combine backend test coverage)
	@rm -f $(BACKEND_COVER_OUT)
	@echo "mode: atomic" > $(BACKEND_COVER_OUT)
	@for file in $(BACKEND_COVER_OUT_UNIT) $(BACKEND_COVER_OUT_INTEGRATION); do \
		cat $$file | egrep -v ${GO_TEST_IGNORE} >> $(BACKEND_COVER_OUT); \
	done
	@rm -f $(BACKEND_COVER_OUT_UNIT) $(BACKEND_COVER_OUT_INTEGRATION)
	@$(GO_EXEC) tool cover -func "$(BACKEND_COVER_OUT)"

.PHONY: test.frontend
test.frontend: test.frontend.e2e ## Run all front-end tests

.PHONY: test.frontend.e2e
test.frontend.e2e: ## Run front-end end-to-end tests
	$(call log, execute front-end end-to-end tests)
	@$(MAKE) start.backend
	@$(PNPM_RUN) test:e2e
	@trap "$(MAKE) stop.backend" EXIT

.PHONY: test.k6
test.k6: ## Run k6 tests
	$(call log, execute k6 tests)
	@$(MAKE) start.backend
	@k6 run $(ROOT_DIR)/tests/main.js
	@trap "$(MAKE) stop.backend" EXIT

.PHONY: lint
lint: lint.backend lint.frontend ## Run linters for the backend and front-end

.PHONY: lint.backend
lint.backend: ## Run linters for the backend
	$(call log, run backend linters)
	@golangci-lint run --timeout 5m

.PHONY: lint.frontend
lint.frontend: ## Run linters for the front-end
	$(call log, run front-end linters)
	@$(PNPM_RUN) lint

.PHONY: format
format: format.backend format.frontend ## Run formatters for the backend and front-end

.PHONY: format.backend
format.backend: ## Run formatters for the backend
	$(call log, run backend formatters)
	@gofmt -l -s -w $(shell pwd)
	@goimports -w $(shell pwd)

.PHONY: format.frontend
format.frontend: ## Run formatters for the front-end
	$(call log, run front-end formatters)
	@$(PNPM_RUN) format

.PHONY: destroy.backend
destroy.backend: stop.backend ## Destroy all backend resources
	$(call log, removing docker resources)
	@docker compose -f deploy/docker/docker-compose.yml down --rmi local --volumes

.PHONY: clean
clean: destroy.backend ## Destroys all backend resources and cleans up untracked files
	$(call log, removing untracked files)
	@git clean -xd --force
