package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/auth/sessionstore"
)

var UnauthorizedError = fmt.Errorf("authentication is unauthorized")

type AuthenticateResult struct {
	UserId       uuid.UUID
	RefreshToken string
	AccessToken  string
}

type SessionAuthenticator struct {
	sessionStore *sessionstore.SessionStore
}

func NewAuthenticator(redisClient *redis.Client) *SessionAuthenticator {
	return &SessionAuthenticator{
		sessionStore: sessionstore.New(redisClient),
	}
}

// Authenticate authenticates the request using the access token from the cookies.
// Returns UnauthorizedError if user is not be authenticated.
func (a *SessionAuthenticator) Authenticate(r *http.Request) (AuthenticateResult, error) {
	accessTokenCookie, err := r.Cookie(accessTokenCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return AuthenticateResult{}, fmt.Errorf("no access token cookie: %w", UnauthorizedError)
	}
	if err != nil {
		return AuthenticateResult{}, err
	}

	accessTokenData, found, err := a.sessionStore.GetAccessTokenData(r.Context(), accessTokenCookie.Value)
	if err != nil {
		return AuthenticateResult{}, fmt.Errorf("error getting access token data: %w", err)
	}
	if !found {
		return AuthenticateResult{}, fmt.Errorf("access token not found: %w", UnauthorizedError)
	}

	return AuthenticateResult{
		UserId:       accessTokenData.UserId,
		RefreshToken: accessTokenData.RefreshToken,
		AccessToken:  accessTokenData.AccessToken,
	}, nil
}
