SHELL := /bin/sh

APP_NAME := homelabwatch
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
IMAGE ?= $(APP_NAME):local

.PHONY: help web-install web-build test build run docker-build release-check release-snapshot clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

web-install: ## Install frontend dependencies with npm ci
	cd web && npm ci

web-build: ## Build the React frontend into web/dist
	cd web && npm run build

test: ## Run Go tests
	go test ./...

build: web-build ## Build the Go binary after compiling frontend assets
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) ./cmd/homelabwatch

run: ## Run the Go application locally
	go run ./cmd/homelabwatch

docker-build: ## Build the Docker image
	docker build -t $(IMAGE) .

release-check: ## Validate the GoReleaser configuration
	goreleaser check

release-snapshot: web-build test ## Build a local snapshot release into dist/
	goreleaser release --snapshot --clean

clean: ## Remove local build outputs
	rm -rf $(BIN_DIR)
	rm -rf web/dist
