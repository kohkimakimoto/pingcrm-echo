.DEFAULT_GOAL := help

# Load .env file. see https://lithic.tech/blog/2020-05/makefile-dot-env
ifneq (,$(wildcard ./.env))
  include .env
  export
endif

# Environment
GO111MODULE := on
PATH := $(CURDIR)/dev/scripts:$(CURDIR)/dev/.external-tools/bin:$(PATH)
SHELL := bash
VERSION := 0.1.0
COMMIT_HASH := $(shell git rev-parse HEAD)
BUILD_LDFLAGS = -s -w \
                -X github.com/kohkimakimoto/pingcrm-echo/app.CommitHash=$(COMMIT_HASH) \
                -X github.com/kohkimakimoto/pingcrm-echo/app.Version=$(VERSION)

# Output help message
# see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[/0-9a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: format
format: ## Format code
	@find . -print | grep --regex '.*\.go' | xargs goimports -w -l -local "github.com/kohkimakimoto/pingcrm-echo"

.PHONY: deps
deps: ## Install go modules
	@go mod tidy

.PHONY: build/dev
build/dev: ## build dev binary
	@if [[ ! -e node_modules ]]; then yarn install; fi
	@npm run build
	@go build -o="dev/build/outputs/dev/pingcrm-echo" .

.PHONY: build/release
build/release: ## build release binaries
	@if [[ ! -e node_modules ]]; then yarn install; fi
	@npm run build
	@rm -rf dev/build/outputs/release && rm -rf dev/build/outputs/archives
	@gox -os="linux darwin" -arch="amd64 arm64" -ldflags="${BUILD_LDFLAGS}" -output "dev/build/outputs/release/pingcrm-echo_{{.OS}}_{{.Arch}}" .
	@mkdir -p dev/build/outputs/archives
	@cd dev/build/outputs && cp -f release/pingcrm-echo_darwin_amd64 archives/pingcrm-echo && cd archives && zip pingcrm-echo_darwin_amd64.zip pingcrm-echo && rm pingcrm-echo
	@cd dev/build/outputs && cp -f release/pingcrm-echo_darwin_arm64 archives/pingcrm-echo && cd archives && zip pingcrm-echo_darwin_arm64.zip pingcrm-echo && rm pingcrm-echo
	@cd dev/build/outputs && cp -f release/pingcrm-echo_linux_amd64 archives/pingcrm-echo && cd archives && zip pingcrm-echo_linux_amd64.zip pingcrm-echo && rm pingcrm-echo
	@cd dev/build/outputs && cp -f release/pingcrm-echo_linux_arm64 archives/pingcrm-echo && cd archives && zip pingcrm-echo_linux_arm64.zip pingcrm-echo && rm pingcrm-echo

.PHONY: clean
clean: ## clean build outputs
	@rm -rf dev/build/outputs
	@rm -rf public/dist

.PHONY: dev/tools/install
dev/tools/install: ## Install dev tools
	@export GOBIN=$(CURDIR)/dev/.external-tools/bin && \
		go install golang.org/x/tools/cmd/goimports@latest && \
		go install github.com/mitchellh/gox@latest && \
		go install github.com/axw/gocov/gocov@latest && \
		go install github.com/matm/gocov-html@latest && \
		go install github.com/cosmtrek/air@latest

.PHONY: dev/tools/clean
dev/tools/clean: ## Clean installed tools
	@rm -rf $(CURDIR)/dev/.external-tools

.PHONY: dev/start
dev/start: prepare-tmp-dir ## start dev server
	@if [[ ! -e public/dist ]]; then yarn build; fi
	@process-starter.py --run "npm run watch" "air"

.PHONY: test
test: ## Test go code
	@go test -timeout 30m ./...

.PHONY: test/verbose
test/verbose: ## Run all tests with verbose outputting.
	@go test -timeout 30m -v ./...

.PHONY: test/coverage
test/coverage: prepare-tmp-dir ## Run all tests with coverage report outputting.
	@gocov test ./... | gocov-html > dev/.tmp/coverage-report.html

.PHONY: deploy/appengine
deploy/appengine: guard-GCLOUD_ACCOUNT guard-GCLOUD_PROJECT clean ## deploy the app to appengine
	@if [[ ! -e app.yaml ]]; then echo "ERROR: require app.yaml file. run 'cp app.example.yaml app.yaml' and update SECRET config" && exit 1; fi
	@npm run build
	@gcloud app deploy app.yaml --account=${GCLOUD_ACCOUNT} --project=${GCLOUD_PROJECT} --quiet --promote

# This is a utility for checking variable definition
# see https://lithic.tech/blog/2020-05/makefile-wildcards/
guard-%:
	@if [[ -z '${${*}}' ]]; then echo 'ERROR: variable $* not set' && exit 1; fi

prepare-tmp-dir:
	@if [[ ! -e dev/.tmp ]]; then mkdir dev/.tmp; fi
