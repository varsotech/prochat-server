package httpserver

import (
	"embed"
	"fmt"
	"github.com/varsotech/prochat-server/internal/httpserver/internal/auth"
	"github.com/varsotech/prochat-server/internal/httpserver/internal/html/pages"
	"html/template"
	"net/http"
)

//go:embed internal/html/components/*.gohtml internal/html/pages/*.gohtml
var templateFS embed.FS

func (s *Server) registerRoutes(mux *http.ServeMux) error {
	authHandlers := auth.NewHandlers(s.postgresClient, s.redisClient)

	mux.HandleFunc("POST /api/v1/auth/login", authHandlers.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authHandlers.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandlers.Refresh)

	templateManager, err := template.ParseFS(templateFS, "internal/html/components/*.gohtml", "internal/html/pages/*.gohtml")
	if err != nil {
		return fmt.Errorf("could not load templates: %w", err)
	}

	mux.HandleFunc("GET /login", pages.Login{TemplateManager: templateManager}.Handler)
	mux.HandleFunc("GET /register", pages.Register{TemplateManager: templateManager}.Handler)

	return nil
}
