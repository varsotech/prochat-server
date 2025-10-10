package oauth

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html/components"
	"github.com/varsotech/prochat-server/internal/html/pages"
	"github.com/varsotech/prochat-server/internal/oauth/clientmetadata"
)

func (o *Routes) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := o.authenticator.Authenticate(r)
	if errors.Is(err, authhttp.UnauthorizedError) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("failed to authenticate user", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	responseType := q.Get("response_type")
	if responseType != "code" {
		http.Error(w, "Invalid response_type parameter. Valid options: ['code']", http.StatusBadRequest)
		return
	}

	clientIdArg := q.Get("client_id")
	clientId, err := clientmetadata.NewClientID(clientIdArg, true)
	if err != nil {
		slog.Debug("client id is not valid", "error", err)
		http.Error(w, "Invalid client_id parameter", http.StatusBadRequest)
		return
	}

	state := q.Get("state")
	if state == "" {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	redirectUri := q.Get("redirect_uri")
	//scope := q.Get("scope")

	// TODO: SHOULD respect HTTP cache headers [RFC9111] when caching client metadata
	// TODO: MAY define its own upper and/or lower bounds on an acceptable cache lifetime as well
	cacheTtl := 1 * time.Hour

	clientMetadata, err := o.clientMetadataResolver.ResolveClientMetadata(r.Context(), clientId, cacheTtl)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if redirectUri == "" {
		// An empty Redirect URI in the request is only allowed for clients with a single Redirect URI
		if len(clientMetadata.Response.RedirectURIs) != 1 {
			http.Error(w, "Invalid redirect_uri parameter", http.StatusBadRequest)
			return
		}

		if len(clientMetadata.Response.RedirectURIs) == 0 {
			// Just in case, we validate it in the resolver
			slog.Error("client metadata has no redirect uri", "client_id", clientId)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		redirectUri = clientMetadata.Response.RedirectURIs[0]
	}

	if err := o.template.ExecuteTemplate(w, "AuthorizePage", pages.AuthorizePage{
		HeadInner: components.HeadInner{
			Title:       "Authorization",
			Description: "Authorization",
		},
		Name:     clientMetadata.Response.ClientName,
		ClientID: clientMetadata.Response.ClientID,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute homepage", "error", err)
		return
	}
}
