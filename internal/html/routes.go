package html

import (
	"embed"
	"fmt"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html/internal/pages"
	"html/template"
	"net/http"
)

type Routes struct {
	template        *template.Template
	authHTTPService *authhttp.Service
}

func NewRoutes(authHTTPService *authhttp.Service) (*Routes, error) {
	templateManager, err := template.ParseFS(templateFS, "internal/components/*.gohtml", "internal/pages/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("could not load templates: %w", err)
	}

	return &Routes{
		template:        templateManager,
		authHTTPService: authHTTPService,
	}, nil
}

//go:embed internal/components/*.gohtml internal/pages/*.gohtml
var templateFS embed.FS

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", pages.Login{Template: o.template}.Handler)
	mux.HandleFunc("GET /register", pages.Register{Template: o.template}.Handler)

	mux.HandleFunc("GET /", o.home)
}
