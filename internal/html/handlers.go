package html

import (
	"errors"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html/internal/components"
	"github.com/varsotech/prochat-server/internal/html/internal/pages"
	"log/slog"
	"net/http"
)

func (o *Service) home(w http.ResponseWriter, r *http.Request) {
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

func (o *Service) login(w http.ResponseWriter, r *http.Request) {
	if err := o.template.ExecuteTemplate(w, "LoginPage", pages.HomePage{
		HeadInner: components.HeadInner{
			Title:       "Login",
			Description: "Login",
		},
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute login page", err)
		return
	}
}

func (o *Service) register(w http.ResponseWriter, r *http.Request) {
	if err := o.template.ExecuteTemplate(w, "RegisterPage", pages.HomePage{
		HeadInner: components.HeadInner{
			Title:       "Register",
			Description: "Register",
		},
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute register page", err)
		return
	}
}
