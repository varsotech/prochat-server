package http

import (
	"errors"
	"fmt"
	"github.com/varsotech/prochat-server/internal/auth/internal/authrepo"
	"net/http"
)

func createCookie(name, value, path string, maxAge int) http.Cookie {
	return http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

var UnauthorizedError = fmt.Errorf("authentication is unauthorized")

// Authenticate authenticates the request using the access token from the cookies.
// Returns UnauthorizedError if user is not be authenticated.
func (s *Service) Authenticate(r *http.Request) (authrepo.AccessTokenData, error) {
	accessTokenCookie, err := r.Cookie(accessTokenCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		return authrepo.AccessTokenData{}, fmt.Errorf("no access token cookie: %w", UnauthorizedError)
	}
	if err != nil {
		return authrepo.AccessTokenData{}, err
	}

	accessTokenData, found, err := o.authRepo.GetAccessTokenData(r.Context(), accessTokenCookie.Value)
	if err != nil {
		return authrepo.AccessTokenData{}, fmt.Errorf("error getting access token data: %w", err)
	}
	if !found {
		return authrepo.AccessTokenData{}, fmt.Errorf("access token not found: %w", UnauthorizedError)
	}

	return accessTokenData, nil
}
