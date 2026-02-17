package websocket

import (
	"context"

	"github.com/varsotech/prochat-server/internal/homeserver/identity"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"google.golang.org/protobuf/proto"
)

func (h *Handlers) GetIdentityToken(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message {
	claims := identity.NewClaims(h.host, auth.UserId)

	token, err := claims.Sign(h.identityPrivateKey)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	payload, err := proto.Marshal(&homeserverv1.GetIdentityTokenResponse{
		Token: token,
	})
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	return &homeserverv1.Message{
		Payload: payload,
	}
}
