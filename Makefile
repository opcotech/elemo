ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
API_DIR:=$(ROOT_DIR)/api
OPENAPI_DIR:=$(API_DIR)/openapi
OPENAPI_GEN_SERVER_DIR:=$(ROOT_DIR)/internal/transport/http/api
COVERAGE_OUT := $(ROOT_DIR)/.coverage.out
COVERAGE_OUT_UNIT := $(ROOT_DIR)/.coverage.unit.out
COVERAGE_OUT_INTEGRATION := $(ROOT_DIR)/.coverage.integration.out
COVERAGE_HTML := $(ROOT_DIR)/coverage.html
GO_EXEC := $(shell which go)
GO_TEST_COVER := $(GO_EXEC) test -json -race -shuffle=on -cover -covermode=atomic
GO_TEST_IGNORE := "testutil|tools|cmd|http\/gen"

default: build

.PHONY: help
help: ## Show available targets
	@echo "Available targets:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: dep
dep: ## Download dependencies
	$(GO_EXEC) mod tidy
	$(GO_EXEC) mod download

.PHONY: dep-update
dep-update: ## Update dependencies
	$(GO_EXEC) get -t -u ./...
	$(GO_EXEC) mod tidy
	$(GO_EXEC) mod download

.PHONY: build
build: generate dep build.backend  ## Build the project

.PHONY: build.backend
build.backend: ## Build service
	docker-compose -f deploy/docker/docker-compose.yml build --no-cache

.PHONY: build.monitoring
build.monitoring: ## Build service
	docker-compose -f deploy/docker/docker-compose.monitoring.yml build

.PHONY: generate
generate: generate.openapi generate.email ## Generate code

.PHONY: generate.openapi
generate.openapi: ## Generate http server code from openapi spec
	mkdir -p $(OPENAPI_GEN_SERVER_DIR)
	oapi-codegen -config $(OPENAPI_DIR)/generator.config.yml -o $(OPENAPI_GEN_SERVER_DIR)/server.go $(OPENAPI_DIR)/openapi.yaml

.PHONY: generate.email
generate.email: ## Generate email templates
	mjml --config.minify=true --config.minifyOptions='{"minifyCSS": true}' --config.validationLevel=strict -r $(ROOT_DIR)/templates/email/*.mjml -o $(ROOT_DIR)/templates/email

.PHONY: start.backend
start.backend: ## Start service
	docker-compose -f deploy/docker/docker-compose.yml up --build -d
	docker-compose -f deploy/docker/docker-compose.yml logs -f

.PHONY: start.monitoring
start.monitoring: ## Start service
	docker-compose -f deploy/docker/docker-compose.monitoring.yml up --build -d
	docker-compose -f deploy/docker/docker-compose.monitoring.yml logs -f

.PHONY: stop.backend
stop.backend: ## Halt service
	docker-compose -f deploy/docker/docker-compose.yml stop

.PHONY: stop.monitoring
stop.monitoring: ## Halt service
	docker-compose -f deploy/docker/docker-compose.monitoring.yml stop

.PHONY: destroy.backend
destroy.backend: ## Remove service resources
	docker-compose -f deploy/docker/docker-compose.yml down --rmi local --volumes

.PHONY: destroy.monitoring
destroy.monitoring: ## Remove service resources
	docker-compose -f deploy/docker/docker-compose.monitoring.yml down --rmi local --volumes

.PHONY: bench
bench: bench.backend ## Run all benchmarks

.PHONY: bench.backend
bench.backend: ## Run backend benchmarks
	$(GO_EXEC) test -run=Bench -bench=. -benchmem -benchtime=10s ./...

.PHONY: format
format: dep ## Format source code
	@gofmt -l -s -w $(shell pwd)
	@goimports -w $(shell pwd)

.PHONY: lint
lint: lint.license lint.backend ## Run linters on the project

.PHONY: lint.backend
lint.backend: dep ## Run linters on the backend
	@golangci-lint run --timeout 5m

.PHONY: lint.license
lint.license: dep ## Check license headers
	@./scripts/extract-and-lint-licenses.sh

.PHONY: test
test: test.unit test.integration ## Run all tests

.PHONY: test.unit
test.unit: ## Run unit tests
	@rm -f $(COVERAGE_OUT_UNIT)
	@$(GO_TEST_COVER) -short -coverprofile=$(COVERAGE_OUT_UNIT) ./...

.PHONY: test.integration
test.integration: ## Run integration tests
	@rm -f $(COVERAGE_OUT_INTEGRATION)
	@$(GO_TEST_COVER) -timeout 3600s -run=Integration -coverprofile=$(COVERAGE_OUT_INTEGRATION) ./...

.PHONY: coverage.combine
coverage.combine:
	@rm -f $(COVERAGE_OUT)
	@echo "mode: atomic" > $(COVERAGE_OUT)
	@for file in $(COVERAGE_OUT_UNIT) $(COVERAGE_OUT_INTEGRATION); do \
		cat $$file | egrep -v "(mode: atomic|$(shell echo $GO_TEST_IGNORE))" >> $(COVERAGE_OUT); \
	done
	@rm -f $(COVERAGE_OUT_UNIT) $(COVERAGE_OUT_INTEGRATION)

.PHONY: coverage.html
coverage.html: ## Generate html coverage report from previous test run
	$(GO_EXEC) tool cover -html "$(COVERAGE_OUT)" -o "$(COVERAGE_HTML)"

.PHONY: coverage.stats
coverage.stats: ## Generate coverage stats from previous test run
	$(GO_EXEC) tool cover -func "$(COVERAGE_OUT)"

.PHONY: changelog
changelog: ## Generate changelog
	git cliff > CHANGELOG.md
