package http

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/homeserver/auth/service"
)

type Authenticator interface {
	Authenticate(r *http.Request) (AuthenticateResult, error)
}

type Routes struct {
	service       *service.Service
	authenticator Authenticator
}

func New(pgClient *pgxpool.Pool, redisClient *redis.Client, host string) *Routes {
	return &Routes{
		service:       service.New(pgClient, redisClient, host),
		authenticator: NewAuthenticator(redisClient),
	}
}

func (s *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", s.loginHandler)
	mux.HandleFunc("POST /api/v1/auth/register", s.registerHandler)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refreshHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logoutHandler)
}
