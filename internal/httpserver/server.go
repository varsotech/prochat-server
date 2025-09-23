package httpserver

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varsotech/prochat-server/internal/auth"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	Ctx            context.Context
	PostgresClient *pgxpool.Pool
}

func New(ctx context.Context, postgresClient *pgxpool.Pool) *Server {
	return &Server{
		Ctx:            ctx,
		PostgresClient: postgresClient,
	}
}

func (s Server) Serve() error {
	mux := http.NewServeMux()

	authRoutes := auth.NewRoutes(s.PostgresClient)

	mux.HandleFunc("POST /api/v1/auth/login", authRoutes.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authRoutes.Register)

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
