package grpcclient

import (
	"context"
	"fmt"

	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Params contains dependencies for gRPC client creation
type Params struct {
	fx.In

	Config    *config.Config
	Logger    *zap.Logger
	Lifecycle fx.Lifecycle
}

// Module exports the gRPC client provider
// Creates and configures the gRPC client connection to product service
var Module = fx.Module("grpcclient",
	fx.Provide(NewProductClient),
)

func NewProductClient(p Params) (productsv1.ProductServiceClient, error) {
	targetAddr := fmt.Sprintf(":%d", p.Config.ServerConfig.ProductServicePort)
	conn, err := grpc.NewClient(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		p.Logger.Error("failed to create new client", zap.String("error", err.Error()))
		return nil, err
	}

	// Register cleanup with fx lifecycle
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("closing grpc connection")
			return conn.Close()
		},
	})

	client := productsv1.NewProductServiceClient(conn)
	p.Logger.Info("gRPC client connected", zap.String("addr", targetAddr))
	return client, nil
}
