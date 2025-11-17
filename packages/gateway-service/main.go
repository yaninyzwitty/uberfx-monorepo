package main

import (
	"context"
	"net/http"

	"github.com/yaninyzwitty/go-fx-v1/packages/gateway-service/controllers"
	grpcclient "github.com/yaninyzwitty/go-fx-v1/packages/gateway-service/internal/grpc-client"
	"github.com/yaninyzwitty/go-fx-v1/packages/gateway-service/internal/router"
	"github.com/yaninyzwitty/go-fx-v1/packages/gateway-service/internal/server"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/telemetry"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l}
		}),
		// Core dependencies
		fx.Provide(zap.NewProduction),

		// Shared modules (order matters: config first, then dependencies)
		config.Module,
		telemetry.Module,

		// Gateway service modules
		grpcclient.Module,  // gRPC client must be provided before controllers
		controllers.Module, // Controllers depend on gRPC client
		router.Module,      // Router depends on controllers (route handlers)
		server.Module,      // Server depends on router (mux)

		// Lifecycle hooks
		fx.Invoke(logStartup),
		fx.Invoke(func(*http.Server) {}), // Ensure server is created
	)

	app.Run()
}

// logStartup logs application startup and shutdown events
func logStartup(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Info("Gateway service starting", zap.String("service", "gateway"))
			return nil
		},
		OnStop: func(context.Context) error {
			logger.Info("Gateway service stopping", zap.String("service", "gateway"))
			return nil
		},
	})
}
