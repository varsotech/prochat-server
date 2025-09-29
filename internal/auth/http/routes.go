package http

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/auth/internal/authrepo"
	"github.com/varsotech/prochat-server/internal/auth/service"
	"net/http"
)

type Service struct {
	service  *service.Service
	authRepo *authrepo.Repo
}

func New(pgClient *pgxpool.Pool, redisClient *redis.Client) *Service {
	return &Service{
		service:  service.New(pgClient, redisClient),
		authRepo: authrepo.New(redisClient),
	}
}

func (o *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", o.login)
	mux.HandleFunc("POST /api/v1/auth/register", o.register)
	mux.HandleFunc("POST /api/v1/auth/refresh", o.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", o.logout)
}
