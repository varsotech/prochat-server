package html

import (
	"embed"
	"fmt"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"html/template"
	"net/http"
)

type Service struct {
	template        *template.Template
	authHTTPService *authhttp.Service
}

func NewRoutes(authHTTPService *authhttp.Service) (*Service, error) {
	templateManager, err := template.ParseFS(templateFS, "internal/components/*.gohtml", "internal/pages/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("could not load templates: %w", err)
	}

	return &Service{
		template:        templateManager,
		authHTTPService: authHTTPService,
	}, nil
}

//go:embed internal/components/*.gohtml internal/pages/*.gohtml
var templateFS embed.FS

func (o *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", o.login)
	mux.HandleFunc("GET /register", o.register)
	mux.HandleFunc("GET /", o.home)
}
