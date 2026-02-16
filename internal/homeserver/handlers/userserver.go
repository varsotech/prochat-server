package handlers

import (
	"context"

	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/homeserverdb"
	"google.golang.org/protobuf/proto"
)

func (h *Handlers) AddUserServerRequest(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message {
	var req homeserverv1.AddUserServerRequest
	err := proto.Unmarshal(message.Payload, &req)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	_, err = h.postgresClient.UpsertUserServer(ctx, homeserverdb.UpsertUserServerParams{
		UserID: auth.UserId,
		Host:   req.Host,
	})
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	resp := homeserverv1.AddUserServerResponse{}

	payload, err := proto.Marshal(&resp)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	return &homeserverv1.Message{
		Type:    homeserverv1.Message_TYPE_ADD_USER_SERVER,
		Payload: payload,
	}
}

func (h *Handlers) GetUserServersRequest(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message {
	var req homeserverv1.GetUserServersRequest
	err := proto.Unmarshal(message.Payload, &req)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	userServers, err := h.postgresClient.GetUserServers(ctx, auth.UserId)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	var servers []*homeserverv1.GetUserServersResponse_Server
	for _, userServer := range userServers {
		servers = append(servers, &homeserverv1.GetUserServersResponse_Server{
			Host: userServer,
		})
	}

	payload, err := proto.Marshal(&homeserverv1.GetUserServersResponse{
		Servers: servers,
	})
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	return &homeserverv1.Message{
		Type:    homeserverv1.Message_TYPE_GET_USER_SERVERS,
		Payload: payload,
	}
}
