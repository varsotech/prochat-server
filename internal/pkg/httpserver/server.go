package httpserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Registrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

type Server struct {
	ctx  context.Context
	port string
}

func New(ctx context.Context, port string) *Server {
	return &Server{
		ctx:  ctx,
		port: port,
	}
}

func (s *Server) Serve(registrars ...Registrar) error {
	mux := http.NewServeMux()

	for _, reg := range registrars {
		reg.RegisterRoutes(mux)
	}

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
		BaseContext: func(listener net.Listener) context.Context {
			return s.ctx
		},
	}

	// Shutdown server when context is cancelled
	go func() {
		<-s.ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("failed to shutdown http server", "error", err)
		}
	}()

	slog.Info("http server is ready to accept connections", "port", s.port)
	return srv.ListenAndServe()
}
