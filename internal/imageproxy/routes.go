package imageproxy

import (
	"net/http"

	"github.com/varsotech/prochat-server/internal/pkg/filestore"
	"github.com/varsotech/prochat-server/internal/pkg/httputil"
)

type URLSigner interface {
	GenerateSignature(inputUrl string) string
}

type Routes struct {
	urlSigner  URLSigner
	httpClient *httputil.Client
	fileStore  filestore.FileStore
}

type Config struct {
	FileStore            filestore.FileStore
	ImageProxyBaseUrl    string
	ImageProxySecretKey  string
	ImageProxySecretSalt string
}

func NewRoutes(fileStore filestore.FileStore, config *Config) *Routes {
	return &Routes{
		urlSigner:  NewSigner(config),
		httpClient: httputil.NewClient(),
		fileStore:  fileStore,
	}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/imageproxy/external/", o.externalHandler)
}
