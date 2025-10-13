package oauth

import (
	"context"
	"html/template"
	"io"
	"net/http"

	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/oauth/clientmetadata"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
)

type Authenticator interface {
	Authenticate(r *http.Request) (authhttp.AuthenticateResult, error)
}

type TemplateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type RemoteFileStore interface {
	Store(ctx context.Context, key string) (string, error)
}

type Routes struct {
	clientMetadataResolver *clientmetadata.Resolver
	authenticator          Authenticator
	template               TemplateExecutor
}

func NewRoutes(redisClient *redis.Client, authenticator Authenticator, httpClient *httputil.Client, template *template.Template) *Routes {
	clientMetadataCache := clientmetadata.NewCache(redisClient)
	return &Routes{
		clientMetadataResolver: clientmetadata.NewResolver(httpClient, clientMetadataCache),
		authenticator:          authenticator,
		template:               template,
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/oauth/authorize", o.authorizeHandler)
}
