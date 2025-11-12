# GitHub Workflows

This directory contains GitHub Actions workflows for the Go microservices monorepo.

## Workflows

### 1. CI (`ci.yml`)

Main continuous integration workflow that runs on push to `main`/`develop` branches:

- Runs linting and tests
- Builds and pushes Docker images (only on push to main/develop)
- Publishes test results

### 2. Pull Request (`pull-request.yml`)

Runs on pull requests to `main`/`develop` branches:

- Runs linting and tests
- Publishes test results
- Does NOT build Docker images

### 3. Lint (`lint.yml`)

Reusable workflow for running golangci-lint:

- Auto-detects all Go modules in the repository
- Runs golangci-lint with caching for performance
- Uses Go 1.23 and golangci-lint v1.61.0

### 4. Unit Tests (`unit-test.yml`)

Reusable workflow for running unit tests:

- Auto-detects all Go modules in the repository
- Runs tests with race detection and coverage
- Uploads coverage reports as artifacts

### 5. Docker Build (`docker-build.yml`)

Builds and pushes Docker images:

- Only builds services that have changed (using path filters)
- Builds multi-platform images (linux/amd64, linux/arm64)
- Uses GitHub Actions cache for faster builds
- Pushes to Docker Hub with `latest` and commit SHA tags

## Configuration

### Required Secrets

- `DOCKER_PASSWORD`: Docker Hub password/token

### Required Variables

- `DOCKER_USERNAME`: Docker Hub username

### Path Filters

The Docker build workflow monitors these paths for changes:

- `packages/product-service/**` → builds product-service
- `packages/gateway-service/**` → builds gateway-service
- `packages/shared/**` → builds all services (shared dependency)
- `gen/**` → builds all services (generated code)
- `go.work*` → builds all services (workspace changes)

## Local Development

Use the provided Makefile for local development:

```bash
# Build all services
make build

# Run tests
make test

# Run linter
make lint

# Build Docker images
make docker-build

# Build and push Docker images
DOCKER_USERNAME=your-username make docker-push
```

## Notes

- All workflows use Go 1.25
- Docker builds use the repository root as context (required for Go workspace)
- Workflows automatically detect Go modules and services with Dockerfiles
- Caching is enabled for Go modules, build cache, and Docker layers
