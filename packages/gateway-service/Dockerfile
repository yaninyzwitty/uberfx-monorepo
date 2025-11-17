# Multi-stage build for gateway-service
# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory to workspace root
WORKDIR /workspace

# Copy go.work and go.work.sum files
COPY go.work go.work.sum ./

# Copy gen module (generated protobufs)
COPY gen/ ./gen/

# Copy shared module
COPY packages/shared/ ./packages/shared/
# Copy product-service module (needed for go.work)
COPY packages/product-service/ ./packages/product-service/
# Copy gateway-service module
COPY packages/gateway-service/ ./packages/gateway-service/

# Build the application
# Using -C to change directory, but building from workspace root for workspace dependencies
WORKDIR /workspace/packages/gateway-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway-service ./main.go

# Stage 2: Runtime
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /workspace/packages/gateway-service/gateway-service .

# Copy config file if needed
COPY --from=builder /workspace/packages/gateway-service/config.yaml ./config.yaml

# Expose port
EXPOSE 8080

# Run the service
CMD ["./gateway-service"]

