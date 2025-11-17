package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductServiceHandler_GetProduct_InvalidID(t *testing.T) {
	handler := &ProductServiceHandler{
		log:    zap.NewNop(),
		tracer: noop.NewTracerProvider().Tracer("test"),
	}

	req := &productsv1.GetProductRequest{Id: 0}
	_, err := handler.GetProduct(context.Background(), req)
	
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}
