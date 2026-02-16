package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/homeserverdb"
)

type handlerFunc = func(context.Context, *oauth.AuthorizeResult, *homeserverv1.Message) *homeserverv1.Message

type Handlers struct {
	postgresClient *homeserverdb.Queries
	handlerMap     map[homeserverv1.Message_Type]handlerFunc
}

func New(postgresClient *pgxpool.Pool) *Handlers {
	h := Handlers{
		postgresClient: homeserverdb.New(postgresClient),
	}

	h.handlerMap = map[homeserverv1.Message_Type]handlerFunc{
		homeserverv1.Message_TYPE_ADD_USER_SERVER:  h.AddUserServerRequest,
		homeserverv1.Message_TYPE_GET_USER_SERVERS: h.GetUserServersRequest,
	}

	return &h
}

func (h *Handlers) Handle(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message {
	handler, ok := h.handlerMap[message.Type]
	if !ok {
		return &homeserverv1.Message{
			Type: message.Type,
			Error: &homeserverv1.Message_Error{
				Message: "Unknown message type",
			},
		}
	}

	return handler(ctx, auth, message)
}
