package httpserver

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/auth"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	ctx            context.Context
	postgresClient *pgxpool.Pool
	redisClient    *redis.Client
	port           string
}

func New(ctx context.Context, port string, postgresClient *pgxpool.Pool, redisClient *redis.Client) *Server {
	return &Server{
		ctx:            ctx,
		postgresClient: postgresClient,
		redisClient:    redisClient,
		port:           port,
	}
}

func (s Server) Serve() error {
	mux := http.NewServeMux()

	authRoutes := auth.NewRoutes(s.postgresClient, s.redisClient)

	mux.HandleFunc("POST /api/v1/auth/login", authRoutes.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authRoutes.Register)

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

	slog.Info("http server is ready to accept connections")
	return srv.ListenAndServe()
}
