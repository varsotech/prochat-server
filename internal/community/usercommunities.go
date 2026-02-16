package community

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	authhttp "github.com/varsotech/prochat-server/internal/homeserver/auth/http"
)

func (o *Routes) getUserCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
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
}
