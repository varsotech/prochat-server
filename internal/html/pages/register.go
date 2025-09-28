package pages

import (
	"github.com/varsotech/prochat-server/internal/httpserver/internal/html/components"
	"html/template"
	"log/slog"
	"net/http"
)

type Register struct {
	Template *template.Template
}

type RegisterPage struct {
	HeadInner components.HeadInner
}

func (h Register) Handler(w http.ResponseWriter, r *http.Request) {
	if err := h.Template.ExecuteTemplate(w, "RegisterPage", LoginPage{
		HeadInner: components.HeadInner{
			Title:       "Register",
			Description: "Register",
		},
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("failed to execute homepage", err)
		return
	}
}
