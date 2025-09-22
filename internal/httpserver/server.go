package httpserver

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varsotech/prochat-server/internal/httpserver/routes/auth"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	Ctx            context.Context
	PostgresClient *pgxpool.Pool
}

func (s Server) Serve() error {
	mux := http.NewServeMux()

	auth.NewRoutes(s.PostgresClient).RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    "localhost:8090",
		Handler: mux,
		BaseContext: func(listener net.Listener) context.Context {
			return s.Ctx
		},
	}

	// Shutdown server when context is cancelled
	go func() {
		<-s.Ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("failed to shutdown http server", "error", err)
		}
	}()

	slog.Info("http server is ready to accept connections")
	return srv.ListenAndServe()
}
