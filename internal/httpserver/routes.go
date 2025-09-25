package httpserver

import (
	"github.com/varsotech/prochat-server/internal/auth"
	"net/http"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	authHandlers := auth.NewHandlers(s.postgresClient, s.redisClient)
	
	mux.HandleFunc("POST /api/v1/auth/login", authHandlers.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authHandlers.Register)
}
