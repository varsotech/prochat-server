package community

import (
	"context"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	authhttp "github.com/varsotech/prochat-server/internal/homeserver/auth/http"
	"github.com/varsotech/prochat-server/internal/homeserver/handlers"
	"github.com/varsotech/prochat-server/internal/homeserver/html"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
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

	authService  *authhttp.Service
	htmlService  *html.Routes
	oauthService *oauth.Routes
}

// NewRoutes exposes HTTP routes struct for the homeserver WebSocket API.
// These routes are accessed by clients with OAuth credentials.
func NewRoutes(redisClient *redis.Client, postgresClient *pgxpool.Pool) *Routes {
	return &Routes{
		authorizer: oauth.NewAuthorizer(),
		handlers:   handlers.New(postgresClient),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/community/user_communities", o.getUserCommunitiesHandler)
}
