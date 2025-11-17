.PHONY: help build test lint clean docker-build docker-push

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build all services
build: ## Build all services
	@echo "Building all services..."
	cd packages/product-service && go build -o ../../bin/product-service ./main.go
	cd packages/gateway-service && go build -o ../../bin/gateway-service ./main.go

# Run tests for all modules
test: ## Run tests for all modules
	@echo "Running tests..."
	go work sync
	cd packages/product-service && go test -v -race ./...
	cd packages/gateway-service && go test -v -race ./...
	cd packages/shared && go test -v -race ./...

# Run linter for all modules
lint: ## Run golangci-lint for all modules
	@echo "Running linter..."
	golangci-lint run ./packages/product-service/...
	golangci-lint run ./packages/gateway-service/...
	golangci-lint run ./packages/shared/...

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	go clean -cache

# Build Docker images
docker-build: ## Build Docker images for all services
	@echo "Building Docker images..."
	docker build -f packages/product-service/Dockerfile -t product-service:latest .
	docker build -f packages/gateway-service/Dockerfile -t gateway-service:latest .

# Push Docker images (requires DOCKER_USERNAME env var)
docker-push: docker-build ## Build and push Docker images
	@if [ -z "$(DOCKER_USERNAME)" ]; then echo "DOCKER_USERNAME not set"; exit 1; fi
	docker tag product-service:latest $(DOCKER_USERNAME)/product-service:latest
	docker tag gateway-service:latest $(DOCKER_USERNAME)/gateway-service:latest
	docker push $(DOCKER_USERNAME)/product-service:latest
	docker push $(DOCKER_USERNAME)/gateway-service:latest

# Setup development environment
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go work sync
	go mod download

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go work sync
	cd packages/product-service && go test -v -coverprofile=coverage.out ./...
	cd packages/gateway-service && go test -v -coverprofile=coverage.out ./...
	cd packages/shared && go test -v -coverprofile=coverage.out ./...

# Run integration tests
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go work sync
	cd packages/product-service && go test -v -tags=integration ./...
	cd packages/gateway-service && go test -v -tags=integration ./...
