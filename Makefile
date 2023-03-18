ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
ASSETS_DIR:=$(ROOT_DIR)/assets
TEMP_DIR:=$(ROOT_DIR)/tmp
API_DIR:=$(ROOT_DIR)/api
OPENAPI_DIR:=$(API_DIR)/openapi
OPENAPI_GEN_SERVER_DIR:=$(ROOT_DIR)/internal/transport/http/gen
COVERAGE_OUT := $(ROOT_DIR)/.coverage.out
COVERAGE_OUT_UNIT := $(ROOT_DIR)/.coverage.unit.out
COVERAGE_OUT_INTEGRATION := $(ROOT_DIR)/.coverage.integration.out
COVERAGE_HTML := $(ROOT_DIR)/coverage.html
GO_EXEC := $(shell which go)
GO_TEST_COVER := $(GO_EXEC) test -shuffle=on -race -cover -covermode=atomic
JAVA_EXEC := $(shell which java)

define integration-test
$(eval COVERAGE_OUT_PART := $(ROOT_DIR)/.coverage.part.$(shell uuidgen).out)
$(GO_TEST_COVER) -tags "integration" -coverprofile=$(COVERAGE_OUT_PART) -coverpkg=$(1) ./... && \
	if [ -f $(COVERAGE_OUT_PART) ]; then \
		cat $(COVERAGE_OUT_PART) >> $(COVERAGE_OUT_INTEGRATION); \
		rm $(COVERAGE_OUT_PART); \
	fi;
endef

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
	docker-compose -f deploy/docker/docker-compose.yml build

.PHONY: build.monitoring
build.monitoring: ## Build service
	docker-compose -f deploy/docker/docker-compose.monitoring.yml build

.PHONY: generate
generate: generate.openapi ## Generate code

.PHONY: generate.openapi
generate.openapi: ## Generate http server code from openapi spec
	swagger-cli bundle $(OPENAPI_DIR)/openapi.yaml --outfile $(OPENAPI_DIR)/openapi.final.yaml --type yaml
	oapi-codegen -config $(OPENAPI_DIR)/generator.config.yml -o $(OPENAPI_GEN_SERVER_DIR)/server.gen.go $(OPENAPI_DIR)/openapi.final.yaml

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
	docker-compose -f deploy/docker/docker-compose.yml down --rmi local --volumes --remove-orphans

.PHONY: destroy.monitoring
destroy.monitoring: ## Remove service resources
	docker-compose -f deploy/docker/docker-compose.monitoring.yml down --rmi local --volumes --remove-orphans

.PHONY: bench
bench: bench.backend ## Run all benchmarks

.PHONY: bench.backend
bench.backend: ## Run backend benchmarks
	$(GO_EXEC) test -bench=. -benchmem -benchtime=10s ./...

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
	@$(GO_TEST_COVER) -coverprofile=$(COVERAGE_OUT_UNIT) ./...

.PHONY: test.integration
test.integration: ## Run integration tests
	@rm -f $(COVERAGE_OUT_INTEGRATION)
	@echo "mode: atomic" > $(COVERAGE_OUT_INTEGRATION)
	$(eval PKGS := $(shell $(GO_EXEC) list ./... | egrep -v "(testutil|tools)" ))
	@$(foreach var,$(PKGS),$(call integration-test,$(var)))

.PHONY: coverage.combine
coverage.combine:
	@rm -f $(COVERAGE_OUT)
	@echo "mode: atomic" > $(COVERAGE_OUT)
	@for file in $(COVERAGE_OUT_UNIT) $(COVERAGE_OUT_INTEGRATION); do \
		cat $$file | egrep -v "(mode: atomic|testutil|tools|cmd|server\/graph)" >> $(COVERAGE_OUT); \
	done
	@rm -f $(COVERAGE_OUT_UNIT) $(COVERAGE_OUT_INTEGRATION)

.PHONY: coverage.html
coverage.html: coverage.combine ## Generate html coverage report from previous test run
	$(GO_EXEC) tool cover -html "$(COVERAGE_OUT)" -o "$(COVERAGE_HTML)"

.PHONY: coverage.stats
coverage.stats: coverage.combine ## Generate coverage stats from previous test run
	$(GO_EXEC) tool cover -func "$(COVERAGE_OUT)"

.PHONY: changelog
changelog: ## Generate changelog
	git cliff > CHANGELOG.md
