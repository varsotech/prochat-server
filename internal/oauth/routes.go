package oauth

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/oauth/clientmetadata"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
	"html/template"
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
	clientMetadataResolver *clientmetadata.Resolver
	authenticator          Authenticator
	template               TemplateExecutor
}

func NewRoutes(redisClient *redis.Client, s3Client *s3.Client, authenticator Authenticator, httpClient *httputil.Client, template *template.Template) *Routes {
	clientMetadataCache := clientmetadata.NewCache(redisClient)
	imageStore := clientmetadata.NewImageStore(s3Client)
	return &Routes{
		clientMetadataResolver: clientmetadata.NewResolver(httpClient, clientMetadataCache, imageStore),
		authenticator:          authenticator,
		template:               template,
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/oauth/authorize", o.authorizeHandler)
}
