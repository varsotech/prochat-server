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

	"github.com/varsotech/prochat-server/internal/pkg/httputil"
)

var ErrNotAccessible = errors.New("client not accessible")

type Resolver struct {
	client     *httputil.Client
	cache      *Cache
	imageStore *ImageStore
}

func NewResolver(httpClient *httputil.Client, cache *Cache, imageStore *ImageStore) *Resolver {
	return &Resolver{
		client:     httpClient,
		cache:      cache,
		imageStore: imageStore,
	}
}

// ResolveClientMetadata fetches and validates OAuth client metadata document based on identifier in URL format.
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
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to execute http request: %w", ErrNotAccessible, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: invalid http status code: %d", ErrNotAccessible, resp.StatusCode)
	}

	// SHOULD limit the response size when fetching the client metadata document.
	// Recommended maximum response size for client metadata documents is 5 kilobytes
	const maxBodySize = 5 * 1024 // 5 KB
	limitedBody := io.LimitReader(resp.Body, maxBodySize)

	var response Response
	if err := json.NewDecoder(limitedBody).Decode(&response); err != nil {
		return nil, fmt.Errorf("invalid client metadata document: %w", err)
	}

	err = response.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid client metadata: %w", err)
	}

	// MUST match the URL of the document using simple string comparison as defined in RFC3986 Section 6.2.1
	if response.ClientID != clientID.String() {
		return nil, fmt.Errorf("client id does not match url used to fetch it")
	}

	// SHOULD prefetch the file at logo_uri and cache it for the cache duration of the client metadata document
	var cachedLogoUrl string
	if response.LogoURI != nil && *response.LogoURI != "" {
		cachedLogoUrl, err = r.imageStore.Store(ctx, clientID.String(), *response.LogoURI)
		if err != nil {
			return nil, fmt.Errorf("failed storing image to cache: %w", err)
		}
	}

	clientMetadata := ClientMetadata{
		Response:      &response,
		CachedLogoUrl: cachedLogoUrl,
	}

	err = r.cache.SetClientMetadata(ctx, &clientMetadata, ttl)
	if err != nil {
		// Best effort
		slog.Error("failed to set client metadata in cache", "error", err, "client_id", clientID)
	}

	return &clientMetadata, nil
}
