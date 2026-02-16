package community

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var UnauthenticatedError = errors.New("request is unauthenticated")

type AuthenticationResult struct {
	UserId uuid.UUID
}

type Authenticator struct {
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{}
}

// Authenticate authenticates the request using the bearer token from the Authorization header.
// Returns UnauthenticatedError if request is not be authenticated.
func (a *Authenticator) Authenticate(ctx context.Context, authorizationHeader string) (*AuthenticationResult, error) {
	splitHeader := strings.SplitN(authorizationHeader, "Bearer ", 2)

	if len(splitHeader) != 2 {
		return &AuthenticationResult{}, UnauthenticatedError
	}

	authToken := splitHeader[1]

	// 1. Parse JWT without validation
	// 2. Get issuer (e.g. homeserver.com)
	// 3. Get public key from homeserver.com/.well-known/prochat.json
	// 4. Validate JWT

	accessTokenData, found, err := a.sessionStore.GetAccessTokenData(ctx, authorizationToken)
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("error getting access token data: %w", err)
	}
	if !found {
		return &AuthenticationResult{}, fmt.Errorf("access token not found: %w", UnauthenticatedError)
	}

	return &AuthenticationResult{
		UserId: accessTokenData.UserId,
	}, nil
}
