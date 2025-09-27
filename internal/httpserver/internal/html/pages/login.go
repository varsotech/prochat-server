package pages

import (
	"github.com/varsotech/prochat-server/internal/httpserver/internal/html/components"
	"html/template"
	"log/slog"
	"net/http"
)

type Login struct {
	TemplateManager *template.Template
}

type LoginPage struct {
	HeadInner components.HeadInner
}

func (h Login) Handler(w http.ResponseWriter, r *http.Request) {
	if err := h.TemplateManager.ExecuteTemplate(w, "LoginPage", LoginPage{
		HeadInner: components.HeadInner{
			Title:       "Login",
			Description: "Login",
		},
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("failed to execute homepage", err)
		return
	}
}
