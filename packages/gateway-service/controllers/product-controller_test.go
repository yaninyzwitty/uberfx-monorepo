package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProductsRouteHandler_Patterns(t *testing.T) {
	controller := &ProductController{
		logger: zap.NewNop(),
	}
	
	handler := &ProductsRouteHandler{
		controller: controller,
	}

	patterns := handler.Patterns()
	assert.Contains(t, patterns, "/api/v1/products")
	assert.Contains(t, patterns, "/api/v1/products/")
}

func TestProductsRouteHandler_ServeHTTP_NotFound(t *testing.T) {
	controller := &ProductController{
		logger: zap.NewNop(),
	}
	
	handler := &ProductsRouteHandler{
		controller: controller,
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/invalid", nil)
	
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
}
