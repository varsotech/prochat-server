package html

import (
	"errors"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html/internal/components"
	"github.com/varsotech/prochat-server/internal/html/internal/pages"
	"log/slog"
	"net/http"
)

func (o *Routes) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	_, err := o.authHTTPService.Authenticate(r)
	if errors.Is(err, authhttp.UnauthorizedError) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("failed to authenticate user", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := o.template.ExecuteTemplate(w, "HomePage", pages.HomePage{
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
