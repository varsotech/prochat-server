package http

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/auth/service"
)

type Authenticator interface {
	Authenticate(r *http.Request) (AuthenticateResult, error)
}

type Service struct {
	service       *service.Service
	authenticator Authenticator
}

func New(pgClient *pgxpool.Pool, redisClient *redis.Client, authenticator Authenticator) *Service {
	return &Service{
		service:       service.New(pgClient, redisClient),
		authenticator: authenticator,
	}
}

func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", s.loginHandler)
	mux.HandleFunc("POST /api/v1/auth/register", s.registerHandler)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refreshHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logoutHandler)
}
