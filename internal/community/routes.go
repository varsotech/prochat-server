package community

import (
	"context"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/communitydb"
)

type Authenticator interface {
	Authenticate(ctx context.Context, authorizationHeader string) (*AuthenticationResult, error)
}

type Handlers interface {
	Handle(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message
}

type TemplateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type Routes struct {
	authenticator Authenticator
	communityDb   *communitydb.Queries
}

func NewRoutes(postgresClient *pgxpool.Pool) *Routes {
	return &Routes{
		authenticator: NewIdentityAuthenticator(),
		communityDb:   communitydb.New(postgresClient),
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/community/server/join", o.joinServer)
	mux.HandleFunc("GET /api/v1/community/user_communities", o.getUserCommunitiesHandler)
}
