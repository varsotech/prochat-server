package oauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/homeserver/oauth/apitokenstore"
)

var UnauthorizedError = errors.New("request is unauthorized")

type AuthorizeResult struct {
	UserId       uuid.UUID
	RefreshToken string
	AccessToken  string
}

type Authorizer struct {
	sessionStore *apitokenstore.TokenStore
}

func NewAuthorizer(redisClient *redis.Client) *Authorizer {
	return &Authorizer{
		sessionStore: apitokenstore.New(redisClient),
	}
}

// Authorize authorizes the request using the bearer token from the Authorization header.
// Returns UnauthorizedError if request is not be authorized.
func (a *Authorizer) Authorize(ctx context.Context, authorizationHeader string) (*AuthorizeResult, error) {
	splitHeader := strings.SplitN(authorizationHeader, "Bearer ", 2)

	if len(splitHeader) != 2 {
		return &AuthorizeResult{}, UnauthorizedError
	}

	authorizationToken := splitHeader[1]

	accessTokenData, found, err := a.sessionStore.GetAccessTokenData(ctx, authorizationToken)
	if err != nil {
		return &AuthorizeResult{}, fmt.Errorf("error getting access token data: %w", err)
	}
	if !found {
		return &AuthorizeResult{}, fmt.Errorf("access token not found: %w", UnauthorizedError)
	}

	return &AuthorizeResult{
		UserId:       accessTokenData.UserId,
		RefreshToken: accessTokenData.RefreshToken,
		AccessToken:  accessTokenData.AccessToken,
	}, nil
}
