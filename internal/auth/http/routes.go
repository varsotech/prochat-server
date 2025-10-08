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

func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", s.loginHandler)
	mux.HandleFunc("POST /api/v1/auth/register", s.registerHandler)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refreshHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logoutHandler)
}
