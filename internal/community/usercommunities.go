package community

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	communityserverv1 "github.com/varsotech/prochat-server/internal/models/gen/communityserver/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (o *Routes) getUserCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := o.authenticator.Authenticate(r.Context(), r.Header.Get("Authorization"))
	if errors.Is(err, UnauthenticatedError) {
		slog.Info("community route unauthenticated", "error", err.Error())
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Info("failed to authenticate user to get user communities", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	member, err := o.communityDb.GetMemberByUserAddress(r.Context(), auth.UserAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		o.writeProtoJson(w, &communityserverv1.GetUserCommunitiesResponse{
			Communities: []*communityserverv1.GetUserCommunitiesResponse_Community{},
		})
		return
	}
	if err != nil {
		slog.Info("could not get member by user address", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// TODO: Retrieve real member communities
	_ = member

	response := communityserverv1.GetUserCommunitiesResponse{
		Communities: []*communityserverv1.GetUserCommunitiesResponse_Community{
			{
				Id:   "testid",
				Name: "Test Community",
			},
		},
	}

	o.writeProtoJson(w, &response)
}

func (o *Routes) writeProtoJson(w http.ResponseWriter, m proto.Message) {
	data, err := protojson.Marshal(m)
	if err != nil {
		slog.Info("failed to marshal response", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		slog.Error("failed to write response", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
