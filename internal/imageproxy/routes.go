package imageproxy

import (
	"net/http"
)

type Routes struct {
}

func NewRoutes() *Routes {
	return &Routes{}
}

func (o *Routes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/imageproxy/external/", o.externalHandler)
}
