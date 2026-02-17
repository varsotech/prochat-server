package html

import (
	"errors"
	"log/slog"
	"net/http"

	authhttp "github.com/varsotech/prochat-server/internal/homeserver/auth/http"
	"github.com/varsotech/prochat-server/internal/homeserver/html/components"
	"github.com/varsotech/prochat-server/internal/homeserver/html/pages"
)

func (o *Routes) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	_, err := o.authenticator.Authenticate(r)
	if errors.Is(err, authhttp.UnauthenticatedError) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("failed to authenticate user", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := o.templateExecutor.ExecuteTemplate(w, "HomePage", pages.HomePage{
		HeadInner: components.HeadInner{
			Title:       "Home",
			Description: "Home",
		},
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute homepage", "error", err)
		return
	}
}

func (o *Routes) login(w http.ResponseWriter, r *http.Request) {
	if err := o.templateExecutor.ExecuteTemplate(w, "LoginPage", pages.HomePage{
		HeadInner: components.HeadInner{
			Title:       "Login",
			Description: "Login",
		},
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute login page", "error", err)
		return
	}
}

func (o *Routes) register(w http.ResponseWriter, r *http.Request) {
	if err := o.templateExecutor.ExecuteTemplate(w, "RegisterPage", pages.HomePage{
		HeadInner: components.HeadInner{
			Title:       "Register",
			Description: "Register",
		},
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute register page", "error", err)
		return
	}
}
