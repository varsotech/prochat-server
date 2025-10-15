package html

import (
	"io"
	"net/http"

	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
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
}

func NewRoutes(template TemplateExecutor, redisClient *redis.Client) *Routes {
	return &Routes{
		templateExecutor: template,
		authenticator:    authhttp.NewAuthenticator(redisClient),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", o.home)
	mux.HandleFunc("GET /login", o.login)
	mux.HandleFunc("GET /register", o.register)
}
