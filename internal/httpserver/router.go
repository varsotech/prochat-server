package httpserver

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
)

type Server struct {
	Ctx            context.Context
	PostgresClient *pgxpool.Pool
}

func (s Server) Serve() error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("handling login")
	})

	mux.HandleFunc("POST /api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("handling register")
	})

	slog.Info("http server is ready to accept connections")
	return http.ListenAndServe("localhost:8090", mux)
}
