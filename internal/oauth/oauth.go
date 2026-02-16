package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	authhttp "github.com/varsotech/prochat-server/internal/auth/http"
	"github.com/varsotech/prochat-server/internal/html/components"
	"github.com/varsotech/prochat-server/internal/html/pages"
	prochatv1 "github.com/varsotech/prochat-server/internal/models/gen/prochat/v1"
	"github.com/varsotech/prochat-server/internal/oauth/clientmetadata"
)

func (o *Routes) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := o.authenticator.Authenticate(r)
	if errors.Is(err, authhttp.UnauthenticatedError) {
		// TODO: Should login redirection happen client side like now? Or should we do it in the login handler and validate that its same origin?
		http.Redirect(w, r, fmt.Sprintf("/login?redirectTo=%s", url.QueryEscape(r.URL.String())), http.StatusFound)
		return
	}
	if err != nil {
		slog.Error("failed to authenticate user", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// TODO: Better error handling
	formRes, err := o.parseForm(r)
	if err != nil {
		slog.Error("failed to parse form", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	rawQuery, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		slog.Error("failed to unescape raw query", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	if err := o.template.ExecuteTemplate(w, "AuthorizePage", pages.AuthorizePage{
		HeadInner: components.HeadInner{
			Title:       "Authorization",
			Description: "Authorization",
		},
		Name:        formRes.clientMetadata.Response.ClientName,
		ClientID:    formRes.clientMetadata.Response.ClientID,
		QueryString: rawQuery,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		slog.Error("failed to execute homepage", "error", err)
		return
	}
}

func (o *Routes) authorizeSubmitHandler(w http.ResponseWriter, r *http.Request) {
	authenticate, err := o.authenticator.Authenticate(r)
	if errors.Is(err, authhttp.UnauthenticatedError) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("failed to authenticate user", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	formRes, err := o.parseForm(r)
	if err != nil {
		slog.Error("failed to parse form", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	code, err := o.codeStore.InsertCode(r.Context(), authenticate.UserId, formRes.clientId, formRes.redirectUriParam)
	if err != nil {
		slog.Error("failed to insert code", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Redirect to desired endpoint
	targetUrl := fmt.Sprintf("%s?state=%s&code=%s", formRes.redirectUri, formRes.state, code)
	http.Redirect(w, r, targetUrl, http.StatusFound)
}

type form struct {
	responseType     string
	clientId         string
	state            string
	redirectUri      string
	redirectUriParam string
	clientMetadata   *clientmetadata.ClientMetadata
}

func (o *Routes) parseForm(r *http.Request) (*form, error) {
	q := r.URL.Query()

	responseType := q.Get("response_type")
	if responseType != "code" {
		slog.Debug("invalid response type", "type", responseType)
		return nil, fmt.Errorf("invalid response type parameter")
	}

	clientIdArg := q.Get("client_id")
	clientId, err := clientmetadata.NewClientID(clientIdArg, true)
	if err != nil {
		slog.Debug("client id is not valid", "error", err, "clientId", clientIdArg)
		return nil, fmt.Errorf("invalid client_id parameter: %w", err)
	}

	state := q.Get("state")
	if state == "" {
		slog.Debug("invalid state received")
		return nil, fmt.Errorf("invalid state parameter")
	}

	redirectUriParam := q.Get("redirect_uri")

	// TODO: Scopes
	//scope := q.Get("scope")

	// TODO: SHOULD respect HTTP cache headers [RFC9111] when caching client metadata
	// TODO: MAY define its own upper and/or lower bounds on an acceptable cache lifetime as well
	cacheTtl := 1 * time.Hour

	clientMetadata, err := o.clientMetadataResolver.ResolveClientMetadata(r.Context(), clientId, cacheTtl)
	if err != nil {
		slog.Debug("unable to resolve client metadata", "error", err)
		return nil, fmt.Errorf("unable to resolve client metadata: %w", err)
	}

	redirectUri := redirectUriParam
	if redirectUri == "" {
		// An empty Redirect URI in the request is only allowed for clients with a single Redirect URI
		if len(clientMetadata.Response.RedirectURIs) != 1 {
			slog.Error("client metadata is expected to have only one redirect uri", "client_id", clientId)
			return nil, fmt.Errorf("invalid redirect_uri parameter")
		}

		if len(clientMetadata.Response.RedirectURIs) == 0 {
			// Just in case, we validate it in the resolver
			slog.Error("client metadata has no redirect uri", "client_id", clientId)
			return nil, fmt.Errorf("invalid redirect_uri parameter")
		}

		redirectUri = clientMetadata.Response.RedirectURIs[0]
	}

	return &form{
		responseType:     responseType,
		clientId:         clientIdArg,
		state:            state,
		redirectUri:      redirectUri,
		redirectUriParam: redirectUriParam,
		clientMetadata:   clientMetadata,
	}, nil
}

func (o *Routes) tokenHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("token handler")

	// TODO: Move this
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	q := r.URL.Query()

	grantType := q.Get("grant_type")
	if grantType != "authorization_code" {
		slog.Debug("invalid grant_type", "grant_type", q.Get("grant_type"))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	code := q.Get("code")
	if code == "" {
		slog.Debug("invalid code parameter", "code", q.Get("code"))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	storedCode, err := o.codeStore.DeleteCode(r.Context(), code)
	if errors.Is(err, ErrCodeNotFound) {
		slog.Debug("code not found", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err != nil {
		slog.Debug("failed to retrieve and delete code", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if storedCode.Code != code {
		slog.Debug("invalid code parameter", "code", code, "stored_code", storedCode.Code)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	redirectUri := q.Get("redirect_uri")
	if storedCode.RedirectUriParam != redirectUri {
		slog.Debug("invalid redirect_uri parameter", "redirect_uri", redirectUri)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	clientId := q.Get("client_id")
	if storedCode.ClientId != clientId {
		slog.Debug("invalid client_id parameter", "client_id", clientId)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	issueTokenPairResult, err := o.tokenPairIssuer.IssueTokenPair(r.Context(), storedCode.UserId)
	if err != nil {
		slog.Error("failed to issue token pair", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(prochatv1.TokenResponse{
		AccessToken:           issueTokenPairResult.AccessToken,
		RefreshToken:          issueTokenPairResult.RefreshToken,
		AccessTokenExpiresIn:  issueTokenPairResult.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: issueTokenPairResult.RefreshTokenExpiresIn,
		TokenType:             "bearer",
		Scope:                 "", // TODO: Scopes
	})

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// Write data
	_, err = w.Write(data)
	if err != nil {
		slog.Error("failed to write response", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
