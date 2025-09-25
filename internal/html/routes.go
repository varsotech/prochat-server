package html

import (
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/auth/sessionstore"
	"io"
	"net/http"
)

type Authenticator interface {
	Authenticate(r *http.Request) (authhttp.AuthenticateResult, error)
}

type TemplateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type Routes struct {
	templateExecutor TemplateExecutor
	authenticator    Authenticator
	sessionStore     *sessionstore.SessionStore
}

func NewRoutes(template TemplateExecutor, authenticator Authenticator) *Routes {
	return &Routes{
		templateExecutor: template,
		authenticator:    authenticator,
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", o.home)
	mux.HandleFunc("GET /login", o.login)
	mux.HandleFunc("GET /register", o.register)
}
