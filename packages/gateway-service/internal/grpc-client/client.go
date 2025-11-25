package grpcclient

import (
	"context"
	"fmt"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/prometheus/client_golang/prometheus"
	productsv1 "github.com/yaninyzwitty/go-fx-v1/gen/products/v1"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"
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

// NewProductClient creates a grpc connection and returns a typed client.
// It wires timeout, prometheus client metrics (with exemplars & labels), logging and OTEL stats.
func NewProductClient(p Params) (productsv1.ProductServiceClient, error) {
	// build target address (same pattern you used before)
	targetAddr := fmt.Sprintf("localhost:%d", p.Config.ServerConfig.ProductServicePort)

	// Create client metrics (not registered to any registry - metrics still work without registration)
	clMetrics := grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{
				0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120,
			}),
		),
	)
	// Note: Metrics work without registration, but won't be exposed via Prometheus HTTP endpoint
	// If you need Prometheus scraping, provide a Registry and register clMetrics

	// exemplar provider from trace span context
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// labels provider — extract "user-id" from outgoing gRPC metadata
	labelsFromContext := func(ctx context.Context) prometheus.Labels {
		labels := prometheus.Labels{"user_id": "unknown"}
		if md, ok := grpcMetadata.FromOutgoingContext(ctx); ok {
			if vs := md.Get("user-id"); len(vs) > 0 && vs[0] != "" {
				labels["user_id"] = vs[0]
			}
		}
		return labels
	}

	// simple zap adapter for the middleware logging.Logger
	zapLogger := func(l *zap.Logger) logging.Logger {
		return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
			// convert middleware level to zap level-like behavior
			switch lvl {
			case logging.LevelDebug:
				l.Debug(msg, zap.Any("fields", fields))
			case logging.LevelWarn:
				l.Warn(msg, zap.Any("fields", fields))
			case logging.LevelError:
				l.Error(msg, zap.Any("fields", fields))
			default:
				l.Info(msg, zap.Any("fields", fields))
			}
		})
	}(p.Logger)

	// timeout interceptor for unary calls
	timeoutUnary := timeout.UnaryClientInterceptor(10 * time.Second)

	// build dial options and interceptors
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// OTEL stats handler
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		// Chain of unary interceptors (the order matters)
		grpc.WithChainUnaryInterceptor(
			timeoutUnary,
			clMetrics.UnaryClientInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext),
				grpcprom.WithLabelsFromContext(labelsFromContext),
			),
			logging.UnaryClientInterceptor(zapLogger),
		),
		// Chain of stream interceptors
		grpc.WithChainStreamInterceptor(
			clMetrics.StreamClientInterceptor(
				grpcprom.WithExemplarFromContext(exemplarFromContext),
				grpcprom.WithLabelsFromContext(labelsFromContext),
			),
			logging.StreamClientInterceptor(zapLogger),
		),
	}

	// Dial
	conn, err := grpc.NewClient(targetAddr, dialOpts...)
	if err != nil {
		p.Logger.Error("failed to dial product service", zap.Error(err), zap.String("target", targetAddr))
		return nil, err
	}

	// Register cleanup with fx lifecycle to close the connection on shutdown
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("closing grpc connection", zap.String("addr", targetAddr))
			return conn.Close()
		},
	})

	client := productsv1.NewProductServiceClient(conn)
	p.Logger.Info("gRPC client connected", zap.String("addr", targetAddr))
	return client, nil
}
