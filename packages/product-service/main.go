package main

import (
	"context"

	"github.com/yaninyzwitty/go-fx-v1/packages/product-service/internal/controllers"
	"github.com/yaninyzwitty/go-fx-v1/packages/product-service/internal/server"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/database"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/sonyflake"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
		database.Module,
		sonyflake.Module,

		// Product service modules
		controllers.Module,
		server.Module,

		// Lifecycle hooks
		fx.Invoke(logStartup),
		fx.Invoke(func(*grpc.Server) {}), // Ensure server is created
	)

	app.Run()
}

func logStartup(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Info("Product service server starting", zap.String("service", "product"))
			return nil
		},
		OnStop: func(context.Context) error {
			logger.Info("Product service server stopping", zap.String("service", "product"))
			return nil
		},
	})
}
