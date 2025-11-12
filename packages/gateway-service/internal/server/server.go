package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
	Config    *config.Config
	Mux       *http.ServeMux
}

// Module exports the http server provider
// Creates and configures the http server with registered services
var Module = fx.Module("server",
	fx.Provide(NewHTTPServer),
)

func NewHTTPServer(p Params) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.Config.ServerConfig.GatewayPort),
		Handler: p.Mux,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			p.Logger.Info("http server starting", zap.String("addr", srv.Addr))
			// go srv.Serve(ln)
			srvError := make(chan error, 1)
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					srvError <- fmt.Errorf("server serve error: %w", err)
				}
			}()

			select {
			case <-srvError:
				return fmt.Errorf("server failed to start: %w", err)
			case <-time.After(100 * time.Millisecond):
				p.Logger.Info("http server started", zap.String("addr", srv.Addr))

			default:
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("http server stopping")
			return srv.Shutdown(ctx)
		},
	})
	return srv

}
