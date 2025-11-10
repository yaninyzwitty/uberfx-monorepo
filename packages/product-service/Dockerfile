# Multi-stage build for product-service
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

# Copy product-service module
COPY packages/product-service/ ./packages/product-service/

# Build the application
# Using -C to change directory, but building from workspace root for workspace dependencies
WORKDIR /workspace/packages/product-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service ./main.go

# Stage 2: Runtime
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /workspace/packages/product-service/product-service .

# Copy config file if needed
COPY --from=builder /workspace/packages/product-service/config.yaml ./config.yaml

# Expose gRPC port
EXPOSE 50051

# Run the service
CMD ["./product-service"]

