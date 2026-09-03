.DEFAULT_GOAL := build

PLUGIN_DIR := $(realpath $(dir $(lastword $(MAKEFILE_LIST))))
ROOT_DIR := $(realpath $(PLUGIN_DIR)/../..)
BUILD_DIR ?= $(ROOT_DIR)/build

GO_EXEC ?= $(shell which go)
PNPM_EXEC ?= $(shell which pnpm)

PLUGIN_ID := $(shell awk '/^id:/{print $$2; exit}' '$(PLUGIN_DIR)/plugin.yaml')
PLUGIN_ZIP := $(BUILD_DIR)/plugins/$(PLUGIN_ID).zip

define log
	@echo "[\033[36mINFO\033[0m]\t$(1)" 1>&2;
endef

.PHONY: build
build:
	$(call log, build $(PLUGIN_ID) plugin)
	@mkdir -p '$(BUILD_DIR)/plugins' '$(PLUGIN_DIR)/backend'
	@CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm $(GO_EXEC) -C '$(PLUGIN_DIR)' build -buildmode=c-shared -o backend/plugin.wasm .
	@rm -f '$(PLUGIN_DIR)/backend/plugin.h'
	@cd '$(ROOT_DIR)/web' && NODE_ENV=production $(PNPM_EXEC) exec vite build --config '$(PLUGIN_DIR)/frontend/vite.config.ts'
	@rm -f '$(PLUGIN_ZIP)'
	@cd '$(PLUGIN_DIR)' && zip -q '$(PLUGIN_ZIP)' plugin.yaml backend/plugin.wasm frontend/index.js
	@echo '$(PLUGIN_ZIP)'
