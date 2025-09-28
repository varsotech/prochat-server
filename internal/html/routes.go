package html

import (
	"embed"
	"fmt"
	"github.com/varsotech/prochat-server/internal/html/internal/pages"
	"html/template"
	"net/http"
)

type Routes struct {
	template *template.Template
}

func NewRoutes() (*Routes, error) {
	templateManager, err := template.ParseFS(templateFS, "internal/html/components/*.gohtml", "internal/html/pages/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("could not load templates: %w", err)
	}

	return &Routes{
		template: templateManager,
	}, nil
}

//go:embed html/components/*.gohtml html/pages/*.gohtml
var templateFS embed.FS

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", pages.Login{Template: o.template}.Handler)
	mux.HandleFunc("GET /register", pages.Register{Template: o.template}.Handler)

	// TODO: Protect route with access token cookie
	mux.HandleFunc("GET /", pages.Home{Template: o.template}.Handler)
}
