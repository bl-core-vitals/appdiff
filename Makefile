## Makefile style guide
## https://style-guides.readthedocs.io/en/latest/makefile.html

.DEFAULT_GOAL:=help
SHELL:=/usr/bin/env bash

help: ## Display all available target
	@echo ""
	@echo "Available tasks:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo ""
.PHONY: help

build: ## Build mac binary executable
	@go build -o mac-appdiff
.PHONY: build

build-linux: ## Build linux binary executable
	@GOARCH="amd64" GOOS="linux" go build -o linux-appdiff
.PHONY: build-linux
