package metrics

import (
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"go.uber.org/fx"
)

type GRPCMetricsResult struct {
	fx.Out
	ServerMetrics *grpcprom.ServerMetrics
}

type GRPCMetricsParams struct {
	fx.In
	Registry *prometheus.Registry
}

func NewGRPCMetrics(p GRPCMetricsParams) GRPCMetricsResult {
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(),
	)

	p.Registry.MustRegister(srvMetrics)

	return GRPCMetricsResult{
		ServerMetrics: srvMetrics,
	}
}

var GRPCMetricsModule = fx.Provide(NewGRPCMetrics)
