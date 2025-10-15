package oauth

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/imageproxy"
	"github.com/varsotech/prochat-server/internal/oauth/apitokenstore"
	"github.com/varsotech/prochat-server/internal/oauth/clientmetadata"
)

type ClientMetadataResolver interface {
	ResolveClientMetadata(ctx context.Context, clientID clientmetadata.ClientID, ttl time.Duration) (*clientmetadata.ClientMetadata, error)
}

type Authenticator interface {
	Authenticate(r *http.Request) (authhttp.AuthenticateResult, error)
}

type TokenPairIssuer interface {
	IssueTokenPair(ctx context.Context, userId uuid.UUID) (apitokenstore.IssueTokenPairResult, error)
}

type TemplateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type CodeStore interface {
	InsertCode(ctx context.Context, userId uuid.UUID, clientId, redirectUri string) (string, error)
	DeleteCode(ctx context.Context, code string) (*StoredCode, error)
}

type Routes struct {
	clientMetadataResolver ClientMetadataResolver
	authenticator          Authenticator
	tokenPairIssuer        TokenPairIssuer
	template               TemplateExecutor
	codeStore              CodeStore
}

func NewRoutes(redisClient *redis.Client, template *template.Template, imageProxyConfig *imageproxy.Config) *Routes {
	return &Routes{
		clientMetadataResolver: clientmetadata.NewResolver(redisClient, imageProxyConfig),
		authenticator:          authhttp.NewAuthenticator(redisClient),
		tokenPairIssuer:        apitokenstore.New(redisClient),
		template:               template,
		codeStore:              NewCodeStore(redisClient),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/oauth/authorize", o.authorizeHandler)
	mux.HandleFunc("POST /api/v1/oauth/authorize", o.authorizeSubmitHandler)
	mux.HandleFunc("POST /api/v1/oauth/token", o.tokenHandler)
}
