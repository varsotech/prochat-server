package community

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/varsotech/prochat-server/internal/homeserver/identity"
	homeserverv1 "github.com/varsotech/prochat-server/internal/models/gen/homeserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
	"google.golang.org/protobuf/encoding/protojson"
)

var UnauthenticatedError = errors.New("request is unauthenticated")

type AuthenticationResult struct {
	UserAddress string
}

type IdentityAuthenticator struct {
	httpClient *httputil.Client
}

func NewIdentityAuthenticator() *IdentityAuthenticator {
	return &IdentityAuthenticator{
		httpClient: httputil.NewClient(),
	}
}

// Authenticate authenticates the request using the bearer token from the Authorization header.
// Returns UnauthenticatedError if request is not be authenticated.
func (a *IdentityAuthenticator) Authenticate(ctx context.Context, authorizationHeader string) (*AuthenticationResult, error) {
	splitHeader := strings.SplitN(authorizationHeader, "Bearer ", 2)

	if len(splitHeader) != 2 {
		return &AuthenticationResult{}, UnauthenticatedError
	}

	authToken := splitHeader[1]

	// 1. Parse JWT without validation to get its issuer
	unverifiedIssuer, err := identity.GetUnverifiedIssuer(authToken)
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("failed to get unverified issuer: %w", err)
	}

	// 2. Parse issuer
	if !strings.HasPrefix(unverifiedIssuer, "http://") && !strings.HasPrefix(unverifiedIssuer, "https://") {
		unverifiedIssuer = "https://" + unverifiedIssuer
	}

	unverifiedIssuerUrl, err := url.Parse(unverifiedIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer url: %w", err)
	}

	// 3. Get public key from well known path
	wellKnown, err := a.getWellKnown(ctx, unverifiedIssuerUrl)
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("failed to get well known issuer: %w", err)
	}

	// 4. Validate JWT
	claims, err := identity.Parse(authToken, wellKnown.PublicKey)
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("failed to parse claims: %w", err)
	}

	// 5. Get issuer from JWT
	issuer, err := claims.Claims.GetIssuer()
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("failed to parse subject: %w", err)
	}

	if !strings.HasPrefix(unverifiedIssuer, "http://") && !strings.HasPrefix(issuer, "https://") {
		issuer = "https://" + issuer
	}

	issuerUrl, err := url.Parse(issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer url: %w", err)
	}

	// 6. Get subject from JWT
	subject, err := claims.Claims.GetSubject()
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("failed to parse issuer: %w", err)
	}

	userId, err := uuid.Parse(subject)
	if err != nil {
		return &AuthenticationResult{}, fmt.Errorf("invalid user id: %w", err)
	}

	// Build user address
	userAddress := fmt.Sprintf("%s@%s", userId.String(), issuerUrl.Host)
	return &AuthenticationResult{
		UserAddress: userAddress,
	}, nil
}

func (a *IdentityAuthenticator) getWellKnown(ctx context.Context, u *url.URL) (*homeserverv1.WellKnown, error) {
	wellKnownUrl := u.JoinPath("/.well-known/prochat.json").String()

	req, err := http.NewRequestWithContext(ctx, "GET", wellKnownUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error executing request: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var wellKnown homeserverv1.WellKnown
	err = protojson.Unmarshal(body, &wellKnown)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling well known response: %w", err)
	}

	return &wellKnown, nil
}
