package http

import (
	"github.com/varsotech/prochat-server/internal/auth/service"
	"net/http"
)

type Routes struct {
	service *service.Service
}

func NewRoutes(service *service.Service) *Routes {
	return &Routes{
		service: service,
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", o.Login)
	mux.HandleFunc("POST /api/v1/auth/register", o.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", o.Refresh)
}
