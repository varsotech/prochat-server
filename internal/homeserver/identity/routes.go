package identity

import (
	"net/http"
)

type Routes struct {
	host       string
	publicKey  string
	privateKey string
}

func NewRoutes(host, publicKey, privateKey string) *Routes {
	return &Routes{
		host:       host,
		publicKey:  publicKey,
		privateKey: privateKey,
	}
}

func (s *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/prochat.json", s.wellKnown)

}
