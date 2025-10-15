package clientmetadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/varsotech/prochat-server/internal/imageproxy"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
)

var ErrBadRequest = errors.New("client id is not a valid accessible url to a client metadata document")

type Cache interface {
	SetClientMetadata(ctx context.Context, storedClientMetadata *ClientMetadata, ttl time.Duration) error
	GetClientMetadata(ctx context.Context, clientId string) (*ClientMetadata, bool, error)
}

type URLSigner interface {
	GenerateSignedURL(inputUrl string) string
}

type Resolver struct {
	client    *httputil.Client
	cache     Cache
	urlSigner URLSigner
}

func NewResolver(redisClient *redis.Client, imageProxyConfig *imageproxy.Config) *Resolver {
	return &Resolver{
		client:    httputil.NewClient(),
		cache:     NewCache(redisClient),
		urlSigner: imageproxy.NewSigner(imageProxyConfig),
	}
}

// ResolveClientMetadata fetches and validates OAuth client metadata document based on identifier in URL format.
// Returns ErrBadRequest for invalid or inaccessible URLs and invalid documents.
// See: https://datatracker.ietf.org/doc/draft-parecki-oauth-client-id-metadata-document/
func (r *Resolver) ResolveClientMetadata(ctx context.Context, clientID ClientID, ttl time.Duration) (*ClientMetadata, error) {
	// Aligning with the ATProto implementation, an exception is made for these localhost paths to enable local
	// development of clients. See: https://atproto.com/specs/oauth
	if clientID == "http://localhost" || clientID == "http://localhost/" {
		return &ClientMetadata{
			Response: &Response{
				ClientID:     clientID.String(),
				RedirectURIs: []string{"http://127.0.0.1/"},
				ClientName:   "Localhost",
			},
		}, nil
	}

	// Try retrieving from cache
	cachedClientMetadata, found, err := r.cache.GetClientMetadata(ctx, clientID.String())
	if err != nil {
		return nil, fmt.Errorf("failed retrieving cached client metadata: %w", err)
	}
	if found {
		slog.Debug("cache hit, retrieved client metadata from cache", "client_id", clientID)
		return cachedClientMetadata, nil
	}

	slog.Debug("cache miss, requesting client metadata", "client_id", clientID)

	req, err := http.NewRequestWithContext(ctx, "GET", clientID.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create http request: %w", ErrBadRequest, err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to execute http request: %w", ErrBadRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: invalid http status code: %d", ErrBadRequest, resp.StatusCode)
	}

	// SHOULD limit the response size when fetching the client metadata document.
	// Recommended maximum response size for client metadata documents is 5 kilobytes
	const maxBodySize = 5 * 1024 // 5 KB
	limitedBody := io.LimitReader(resp.Body, maxBodySize)

	var response Response
	if err := json.NewDecoder(limitedBody).Decode(&response); err != nil {
		return nil, fmt.Errorf("%w: invalid client metadata document: %w", ErrBadRequest, err)
	}

	err = response.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: invalid client metadata: %w", ErrBadRequest, err)
	}

	// MUST match the URL of the document using simple string comparison as defined in RFC3986 Section 6.2.1
	if response.ClientID != clientID.String() {
		return nil, fmt.Errorf("%w: client id does not match url used to fetch it", ErrBadRequest)
	}

	// SHOULD prefetch the file at logo_uri and cache it for the cache duration of the client metadata document
	var cachedLogoUrl string
	if response.LogoURI != nil && *response.LogoURI != "" {
		cachedLogoUrl = r.urlSigner.GenerateSignedURL(*response.LogoURI)
	}

	clientMetadata := ClientMetadata{
		Response:      &response,
		CachedLogoUrl: cachedLogoUrl,
	}

	err = r.cache.SetClientMetadata(ctx, &clientMetadata, ttl)
	if err != nil {
		// Fail here to avoid attacks skipping our cache forcing us to use a lot of bandwidth
		return nil, fmt.Errorf("failed to set client metadata in cache: %w", err)
	}

	return &clientMetadata, nil
}
