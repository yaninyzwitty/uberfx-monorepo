package router

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockRouteHandler struct {
	pattern string
}

func (m *mockRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	n, err := w.Write([]byte("test response"))
	if err != nil {
		// handle error, e.g., log it
		log.Printf("failed to write response: %v", err)
		return
	}
	_ = n // optional if you want to ignore bytes written
}

func (m *mockRouteHandler) Pattern() string {
	return m.pattern
}

func TestNewRouter(t *testing.T) {
	logger := zap.NewNop()

	routes := []RouteHandler{
		&mockRouteHandler{pattern: "/test"},
		&mockRouteHandler{pattern: "/health"},
	}

	params := Params{
		Logger: logger,
		Routes: routes,
	}

	mux := NewRouter(params)

	// Test that routes are registered
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}
