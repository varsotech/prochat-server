package pages

import (
	"github.com/varsotech/prochat-server/internal/html/internal/components"
	"html/template"
	"log/slog"
	"net/http"
)

type Home struct {
	Template *template.Template
}

type HomePage struct {
	HeadInner components.HeadInner
}

func (h Home) Handler(w http.ResponseWriter, r *http.Request) {
	if err := h.Template.ExecuteTemplate(w, "HomePage", LoginPage{
		HeadInner: components.HeadInner{
			Title:       "Home",
			Description: "Home",
		},
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("failed to execute homepage", err)
		return
	}
}
