package websocket

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/varsotech/prochat-server/internal/homeserver/identity"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth"
	communityserverv1 "github.com/varsotech/prochat-server/internal/models/gen/communityserver/v1"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/homeserverdb"
	"google.golang.org/protobuf/encoding/protojson"
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
		Payload: payload,
	}
}

func (h *Handlers) GetUserCommunitiesRequest(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message {
	var req homeserverv1.GetUserCommunitiesRequest
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

	identityClaims := identity.NewClaims(h.host, auth.UserId)

	identityJwt, err := identityClaims.Sign(h.identityPrivateKey)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	var userCommunities []*homeserverv1.GetUserCommunitiesResponse_Community
	for _, server := range userServers {
		communities, err := h.getUserCommunitiesForServer(ctx, identityJwt, server)
		if err != nil {
			return &homeserverv1.Message{
				Error: &homeserverv1.Message_Error{
					Message: err.Error(),
				},
			}
		}

		for _, community := range communities.Communities {
			userCommunities = append(userCommunities, &homeserverv1.GetUserCommunitiesResponse_Community{
				Id:   community.Id,
				Name: community.Name,
			})
		}
	}

	payload, err := proto.Marshal(&homeserverv1.GetUserCommunitiesResponse{
		Communities: userCommunities,
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

func (h *Handlers) getUserCommunitiesForServer(ctx context.Context, identityJWT string, server string) (*communityserverv1.GetUserCommunitiesResponse, error) {
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "https://" + server
	}

	u, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("invalid server url: %s", server)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.JoinPath("/api/v1/community/user_communities").String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+identityJWT)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user communities request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var protoResp communityserverv1.GetUserCommunitiesResponse
	err = protojson.Unmarshal(body, &protoResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &protoResp, nil
}

func (h *Handlers) JoinCommunityServer(ctx context.Context, auth *oauth.AuthorizeResult, message *homeserverv1.Message) *homeserverv1.Message {
	var req homeserverv1.JoinCommunityServerRequest
	err := proto.Unmarshal(message.Payload, &req)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	identityClaims := identity.NewClaims(h.host, auth.UserId)

	identityJwt, err := identityClaims.Sign(h.identityPrivateKey)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	_, err = h.joinCommunityServer(ctx, identityJwt, req.Host)
	if err != nil {
		return &homeserverv1.Message{
			Error: &homeserverv1.Message_Error{
				Message: err.Error(),
			},
		}
	}

	payload, err := proto.Marshal(&homeserverv1.AddUserServerResponse{})
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

func (h *Handlers) joinCommunityServer(ctx context.Context, identityJWT string, server string) (*communityserverv1.JoinServerResponse, error) {
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "https://" + server
	}

	u, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("invalid server url: %s", server)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.JoinPath("/api/v1/community/server/join").String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+identityJWT)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user communities request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var protoResp communityserverv1.JoinServerResponse
	err = protojson.Unmarshal(body, &protoResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &protoResp, nil
}
