package httputil

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
	ctx        context.Context
	port       string
	registrars []Registrar
}

func NewServer(ctx context.Context, port string, registrars ...Registrar) *Server {
	return &Server{
		ctx:        ctx,
		port:       port,
		registrars: registrars,
	}
}

func (s *Server) Serve() error {
	mux := http.NewServeMux()

	for _, reg := range s.registrars {
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
