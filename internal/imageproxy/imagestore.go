package imageproxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"path"

	"github.com/varsotech/prochat-server/internal/pkg/filestore"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
)

type ImageStore struct {
	fileStore  filestore.FileStore
	httpClient *httputil.Client
}

func NewImageStore(fileStore filestore.FileStore, httpClient *httputil.Client) *ImageStore {
	return &ImageStore{
		fileStore:  fileStore,
		httpClient: httpClient,
	}
}

func (i *ImageStore) Store(ctx context.Context, clientId, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid http status code: %d", resp.StatusCode)
	}

	_, err = i.fileStore.PutObject(ctx, i.getImageKey(clientId), resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to s3 bucket")
	}

	return clientId, nil
}

func (i *ImageStore) getImageKey(clientId string) string {
	// Hash the client ID to avoid storing a big key for a very long URL
	clientIdHash := sha256.Sum256([]byte(clientId))
	return path.Join("clientmetadatalogos", fmt.Sprintf("%x", clientIdHash))
}
