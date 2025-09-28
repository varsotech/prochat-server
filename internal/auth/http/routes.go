package http

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/auth/service"
	"net/http"
)

type Routes struct {
	service *service.Service
}

func NewRoutes(pgClient *pgxpool.Pool, redisClient *redis.Client) *Routes {
	return &Routes{
		service: service.New(pgClient, redisClient),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", o.Login)
	mux.HandleFunc("POST /api/v1/auth/register", o.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", o.Refresh)
}
