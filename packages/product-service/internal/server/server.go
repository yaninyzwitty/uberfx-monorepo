package server

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"

	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Params struct {
	fx.In

	Lifecycle      fx.Lifecycle
	Logger         *zap.Logger
	Config         *config.Config
	ProductService productsv1.ProductServiceServer
	Metrics        *grpcprom.ServerMetrics
}

// Module exports the gRPC server provider
// Creates and configures the gRPC server with registered services
var Module = fx.Module("server",
	fx.Provide(NewServer),
)

// loggingUnaryInterceptor logs only failed gRPC requests
func loggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Call the handler
		resp, err := handler(ctx, req)

		// Log only errors
		if err != nil {
			st, _ := status.FromError(err)
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.String("error", err.Error()),
				zap.String("code", st.Code().String()),
			)
		}

		return resp, err
	}
}

func NewServer(p Params) *grpc.Server {

	// handle Panics
	panicHandler := func(pan any) (err error) {
		return status.Errorf(codes.Internal, "panic: %v\n%s", pan, debug.Stack())

	}
	// Create server with logging interceptor
	s := grpc.NewServer(
		// otelgrpc stats
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor(p.Logger.Named("grpc_server")),
			p.Metrics.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(panicHandler)),
		),
	)

	// Register product service
	productsv1.RegisterProductServiceServer(s, p.ProductService)

	// Only add server reflection when debug mode is on
	if p.Config.ServerConfig.Debug {
		reflection.Register(s)
	}

	// Register health check service (used for Kubernetes probes)
	healthCheck := health.NewServer()
	healthpb.RegisterHealthServer(s, healthCheck)

	// Mark service as healthy
	healthCheck.SetServingStatus(
		productsv1.ProductService_ServiceDesc.ServiceName,
		healthpb.HealthCheckResponse_SERVING,
	)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := fmt.Sprintf(":%d", p.Config.ServerConfig.ProductServicePort)
			p.Logger.Info("Starting gRPC server", zap.String("addr", addr))

			l, err := net.Listen("tcp", addr)
			if err != nil {
				p.Logger.Error("Failed to listen", zap.Error(err), zap.String("addr", addr))
				return fmt.Errorf("failed to listen on %s: %w", addr, err)
			}

			serverError := make(chan error, 1)
			go func() {
				if err := s.Serve(l); err != nil && err != grpc.ErrServerStopped {
					serverError <- fmt.Errorf("server serve error: %w", err)
				}
			}()

			// we verify if the server started
			select {
			case err := <-serverError:
				return fmt.Errorf("server failed to start: %w", err)
			case <-time.After(100 * time.Millisecond):
				p.Logger.Info("grpc server started", zap.String("addr", addr))
			default:
			}

			return nil
		},

		OnStop: func(ctx context.Context) error {
			s.GracefulStop()
			return nil
		},
	})

	return s

}
