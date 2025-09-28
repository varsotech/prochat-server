package auth

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, authHandlers Handlers) {
	mux.HandleFunc("POST /api/v1/auth/login", authHandlers.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authHandlers.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandlers.Refresh)
}
