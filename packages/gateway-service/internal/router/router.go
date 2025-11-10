package router

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RouteHandler represents an HTTP handler with a pattern
type RouteHandler interface {
	http.Handler
	Pattern() string
}

// MultiPatternRoute is an optional interface for handlers that want to register multiple patterns
type MultiPatternRoute interface {
	RouteHandler
	Patterns() []string
}

// Params contains dependencies for router creation
type Params struct {
	fx.In

	Logger *zap.Logger
	Routes []RouteHandler `group:"routes"`
}

// Module exports the router provider
// Creates and configures the HTTP router with registered route handlers
var Module = fx.Module("router",
	fx.Provide(NewRouter),
)

// NewRouter creates a new HTTP ServeMux and registers all route handlers
func NewRouter(p Params) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range p.Routes {
		var patterns []string
		// Check if route implements MultiPatternRoute interface
		if multiPatternRoute, ok := route.(MultiPatternRoute); ok {
			patterns = multiPatternRoute.Patterns()
		} else {
			patterns = []string{route.Pattern()}
		}

		// Register all patterns for this route
		for _, pattern := range patterns {
			mux.Handle(pattern, route)
			p.Logger.Info("registered route", zap.String("pattern", pattern))
		}
	}

	return mux
}
