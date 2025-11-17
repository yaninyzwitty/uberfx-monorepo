# Go Microservices Monorepo

A modern microservices architecture built with Go, featuring dependency injection with Uber FX, gRPC communication, and comprehensive observability.

## Architecture Overview

This monorepo contains a microservices ecosystem with the following components:

- **Product Service**: gRPC-based service for product management
- **Gateway Service**: HTTP REST API gateway that communicates with backend services
- **Shared Libraries**: Common utilities, database connections, and telemetry

## Services

### Product Service
- **Port**: gRPC server (configurable)
- **Purpose**: Manages product CRUD operations
- **Technology**: Go + gRPC + PostgreSQL
- **Features**: Product creation, retrieval, listing, and deletion

### Gateway Service  
- **Port**: HTTP server (configurable)
- **Purpose**: REST API gateway for client applications
- **Technology**: Go + HTTP + gRPC client
- **Features**: HTTP-to-gRPC translation, request routing

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: Uber FX (Dependency Injection)
- **Communication**: gRPC, HTTP REST
- **Database**: PostgreSQL
- **Code Generation**: Protocol Buffers, SQLC
- **Observability**: OpenTelemetry, Zap logging
- **Build Tools**: Docker, Make
- **CI/CD**: GitHub Actions

## Project Structure

```
├── packages/
│   ├── gateway-service/     # HTTP REST API gateway
│   ├── product-service/     # gRPC product service
│   └── shared/             # Shared libraries and utilities
├── proto/                  # Protocol Buffer definitions
├── gen/                   # Generated code from protobuf
├── scripts/               # Build and utility scripts
├── .github/workflows/     # CI/CD pipelines
├── go.work               # Go workspace configuration
├── Makefile             # Build automation
└── README.md           # This file
```

## Quick Start

### Prerequisites

- Go 1.25 or later
- Docker and Docker Compose
- PostgreSQL (for local development)
- Protocol Buffers compiler (`protoc`)
- Make

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-fx-v1
   ```

2. **Setup development environment**
   ```bash
   make setup
   ```

3. **Build all services**
   ```bash
   make build
   ```

4. **Run tests**
   ```bash
   make test
   ```

### Running Services

#### Using Docker (Recommended)
```bash
# Build Docker images
make docker-build

# Run with docker-compose (if available)
docker-compose up
```

#### Local Development
```bash
# Terminal 1: Start Product Service
cd packages/product-service
go run main.go

# Terminal 2: Start Gateway Service  
cd packages/gateway-service
go run main.go
```

## Configuration

Services use YAML configuration files:

- `packages/product-service/config.yaml`
- `packages/gateway-service/config.yaml`

Environment variables can override configuration values. See `.env.example` files for reference.

## API Documentation

### Product Service (gRPC)

The Product Service exposes the following gRPC methods:

- `GetProduct(id)` - Retrieve a single product
- `ListProducts(page_size, page_token)` - List products with pagination
- `CreateProduct(name, description, price, currency, stock_quantity)` - Create a new product
- `DeleteProduct(id)` - Delete a product

### Gateway Service (HTTP REST)

The Gateway Service provides REST endpoints that proxy to backend services:

- `GET /api/products/{id}` - Get product by ID
- `GET /api/products` - List products with pagination
- `POST /api/products` - Create a new product
- `DELETE /api/products/{id}` - Delete a product

## Development

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build all services
make test              # Run tests for all modules
make test-coverage     # Run tests with coverage
make test-integration  # Run integration tests
make lint              # Run golangci-lint
make clean             # Clean build artifacts
make docker-build      # Build Docker images
make docker-push       # Build and push Docker images
make setup             # Setup development environment
```

### Code Generation

This project uses code generation for:

- **Protocol Buffers**: Generate Go code from `.proto` files
- **SQLC**: Generate type-safe Go code from SQL queries

Regenerate code after making changes:
```bash
# Generate protobuf code
buf generate

# Generate SQLC code (if applicable)
sqlc generate
```

### Testing

The project includes multiple test types:

- **Unit Tests**: `make test`
- **Integration Tests**: `make test-integration`
- **Coverage Reports**: `make test-coverage`

### Linting

Code quality is enforced with golangci-lint:
```bash
make lint
```

## Deployment

### Docker

Each service has its own Dockerfile optimized for production:

```bash
# Build images
make docker-build

# Push to registry (set DOCKER_USERNAME env var)
DOCKER_USERNAME=your-username make docker-push
```

### CI/CD

GitHub Actions workflows handle:

- **Unit Tests**: Run on every PR and push
- **Linting**: Code quality checks
- **Docker Build**: Multi-service container builds
- **Integration Tests**: End-to-end testing

## Observability

### Logging
- Structured logging with Zap
- Configurable log levels
- Request correlation IDs

### Metrics & Tracing
- OpenTelemetry integration
- Distributed tracing across services
- Custom metrics collection

### Health Checks
- Service health endpoints
- Database connectivity checks
- Dependency health monitoring

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linting (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Write meaningful commit messages
- Add tests for new functionality
- Update documentation as needed

## Troubleshooting

### Common Issues

**Build Failures**
```bash
# Clean and rebuild
make clean
make setup
make build
```

**Test Failures**
```bash
# Run tests with verbose output
go test -v ./...
```

**Docker Issues**
```bash
# Clean Docker cache
docker system prune -a
make docker-build
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support:
- Create an issue in the GitHub repository
- Check existing documentation
- Review the troubleshooting section

---

**Built with ❤️ using Go and modern microservices patterns**
