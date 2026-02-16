package homeserver

import (
	"context"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/oauth"
)

type Authorizer interface {
	Authorize(ctx context.Context, authorizationHeader string) (oauth.AuthorizeResult, error)
}

type Routes struct {
	authorizer Authorizer
}

// NewRoutes exposes HTTP routes struct for the homeserver WebSocket API.
// These routes are accessed by clients with OAuth credentials.
func NewRoutes(redisClient *redis.Client) *Routes {
	return &Routes{
		authorizer: oauth.NewAuthorizer(redisClient),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/homeserver/ws", o.ws)
}
