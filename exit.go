package main

// package main

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"net"
// 	"net/http"

// 	"go.uber.org/fx"
// 	"go.uber.org/fx/fxevent"
// 	"go.uber.org/zap"
// )

// type Route interface {
// 	http.Handler
// 	Pattern() string
// }

// type EchoHandler struct{ log *zap.Logger }

// func NewEchoHandler(log *zap.Logger) *EchoHandler {
// 	return &EchoHandler{log: log}
// }

// func (*EchoHandler) Pattern() string { return "/echo" }

// func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	if _, err := io.Copy(w, r.Body); err != nil {
// 		h.log.Warn("Error copying body", zap.Error(err))
// 	}
// }

// type HelloHandler struct{ log *zap.Logger }

// func NewHelloHandler(log *zap.Logger) *HelloHandler {
// 	return &HelloHandler{log: log}
// }

// func (*HelloHandler) Pattern() string { return "/hello" }

// func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		h.log.Error("Error reading request", zap.Error(err))
// 		http.Error(w, "internal error", http.StatusInternalServerError)
// 		return
// 	}
// 	fmt.Fprintf(w, "Hello, %s\n", body)
// }

// type MuxParams struct {
// 	fx.In
// 	Routes []Route `group:"routes"`
// }

// func NewServeMux(p MuxParams) *http.ServeMux {
// 	mux := http.NewServeMux()
// 	for _, r := range p.Routes {
// 		mux.Handle(r.Pattern(), r)
// 	}
// 	return mux
// }

// func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger) *http.Server {
// 	srv := &http.Server{Addr: ":8080", Handler: mux}

// 	lc.Append(fx.Hook{
// 		OnStart: func(ctx context.Context) error {
// 			ln, err := net.Listen("tcp", srv.Addr)
// 			if err != nil {
// 				return err
// 			}
// 			log.Info("http server starting", zap.String("addr", srv.Addr))
// 			go srv.Serve(ln)
// 			return nil
// 		},
// 		OnStop: srv.Shutdown,
// 	})

// 	return srv
// }

// func AsRoute(constructor any) any {
// 	return fx.Annotate(
// 		constructor,
// 		fx.As(new(Route)),
// 		fx.ResultTags(`group:"routes"`),
// 	)
// }

// func main() {
// 	fx.New(
// 		fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
// 			return &fxevent.ZapLogger{Logger: l}
// 		}),
// 		fx.Provide(
// 			zap.NewProduction,
// 			AsRoute(NewEchoHandler),
// 			AsRoute(NewHelloHandler),
// 			NewServeMux,
// 			NewHTTPServer,
// 		),
// 		fx.Invoke(func(*http.Server) {}),
// 	).Run()
// }
