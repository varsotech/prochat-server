package homeserver

import (
	"context"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/homeserver/auth/http"
	"github.com/varsotech/prochat-server/internal/homeserver/html"
	"github.com/varsotech/prochat-server/internal/homeserver/identity"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	"github.com/varsotech/prochat-server/internal/homeserver/websocket"
	"github.com/varsotech/prochat-server/internal/imageproxy"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
)

type Authorizer interface {
	Authorize(ctx context.Context, authorizationHeader string) (*oauth.AuthorizeResult, error)
}

type Handlers interface {
	Handle(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message
}

type TemplateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type Routes struct {
	authorizer Authorizer
	handlers   Handlers

	authService     *authhttp.Routes
	htmlService     *html.Routes
	oauthService    *oauth.Routes
	identityService *identity.Routes
}

// NewRoutes exposes HTTP routes struct for the homeserver WebSocket API.
// These routes are accessed by clients with OAuth credentials.
func NewRoutes(redisClient *redis.Client, postgresClient *pgxpool.Pool, htmlTemplate TemplateExecutor, imageProxyConfig *imageproxy.Config, host, identityPrivateKey, identityPublicKey string) *Routes {
	return &Routes{
		authorizer:      oauth.NewAuthorizer(redisClient),
		handlers:        websocket.New(postgresClient, host, identityPrivateKey),
		authService:     authhttp.New(postgresClient, redisClient, host),
		htmlService:     html.NewRoutes(htmlTemplate, redisClient),
		oauthService:    oauth.NewRoutes(redisClient, htmlTemplate, imageProxyConfig),
		identityService: identity.NewRoutes(host, identityPublicKey, identityPrivateKey),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	o.authService.RegisterRoutes(mux)
	o.htmlService.RegisterRoutes(mux)
	o.oauthService.RegisterRoutes(mux)
	o.identityService.RegisterRoutes(mux)

	mux.HandleFunc("GET /api/v1/homeserver/ws", o.ws)
}
