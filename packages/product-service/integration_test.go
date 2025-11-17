//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "go-fx-v1/gen/products/v1"
)

func TestProductServiceIntegration(t *testing.T) {
	// This test requires the service to be running
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skip("Product service not available for integration test")
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test health check or basic functionality
	_, err = client.GetProduct(ctx, &pb.GetProductRequest{Id: "test"})
	// We expect this to work or return a specific error
	assert.NotNil(t, err) // Adjust based on your actual implementation
}
