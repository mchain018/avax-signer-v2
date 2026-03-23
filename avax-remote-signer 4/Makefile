SHELL := /bin/bash

.PHONY: help build test clean install proto lint fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the signer binary
	@echo "Building avax-remote-signer..."
	go build -o avax-remote-signer ./cmd/signer
	@echo "Build complete: ./avax-remote-signer"

install: ## Install the signer binary
	@echo "Installing avax-remote-signer..."
	go install ./cmd/signer
	@echo "Installed to: $$(go env GOPATH)/bin/signer"

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/signer.proto

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f avax-remote-signer
	rm -f coverage.out coverage.html
	go clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

genkey: ## Generate a test key (for development only)
	@echo "Generating test key..."
	mkdir -p keys
	./avax-remote-signer genkey -o ./keys/test-validator-key.json
	@echo "Test key generated: ./keys/test-validator-key.json"

run-dev: build ## Run signer in development mode
	@echo "Starting signer in development mode..."
	./avax-remote-signer --config config.example.yaml

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t avax-remote-signer:latest .

docker-run: docker-build ## Run signer in Docker
	@echo "Running signer in Docker..."
	docker run -it --rm \
		-v $(PWD)/keys:/keys \
		-v $(PWD)/config.yaml:/config.yaml \
		-p 9090:9090 \
		avax-remote-signer:latest \
		--config /config.yaml

check: fmt lint test ## Run all checks (format, lint, test)

.DEFAULT_GOAL := help
