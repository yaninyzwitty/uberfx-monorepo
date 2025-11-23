package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PromHTTPParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
	Registry  *prometheus.Registry
	Config    *config.Config
}

func NewPromHTTP(p PromHTTPParams) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics",
		promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.Config.ServerConfig.PromHTTPAddr),
		Handler: mux,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}

			go func() {
				if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
					p.Logger.Error("promhttp failed", zap.Error(err))
				}
			}()

			p.Logger.Info("Prometheus /metrics started", zap.String("addr", server.Addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})

	return server
}

var PromHTTPModule = fx.Provide(NewPromHTTP)
