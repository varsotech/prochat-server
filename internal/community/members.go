package community

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	communityserverv1 "github.com/varsotech/prochat-server/internal/models/gen/communityserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/communitydb"
)

func (o *Routes) joinServer(w http.ResponseWriter, r *http.Request) {
	auth, err := o.authenticator.Authenticate(r.Context(), r.Header.Get("Authorization"))
	if errors.Is(err, UnauthenticatedError) {
		slog.Info("community route unauthenticated", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Info("failed to authenticate user to join server", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_, err = o.communityDb.UpsertMember(r.Context(), communitydb.UpsertMemberParams{
		ID:          uuid.New(),
		UserAddress: auth.UserAddress,
	})
	if err != nil {
		slog.Info("failed to upsert member", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	o.writeProtoJson(w, &communityserverv1.JoinServerResponse{})
}
