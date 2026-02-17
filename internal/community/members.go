package community

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	communityserverv1 "github.com/varsotech/prochat-server/internal/models/gen/communityserver/v1"
	"github.com/varsotech/prochat-server/internal/pkg/communitydb"
	"google.golang.org/protobuf/encoding/protojson"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var req communityserverv1.JoinServerRequest
	err = protojson.Unmarshal(body, &req)
	if err != nil {
		slog.Error("failed to unmarshal request", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	member, err := o.communityDb.UpsertMember(r.Context(), communitydb.UpsertMemberParams{
		ID:          uuid.New(),
		UserAddress: auth.UserAddress,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("failed to upsert member", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if req.JoinDefaultCommunity {
		err = o.joinDefaultCommunity(r.Context(), member.ID)
		if err != nil {
			slog.Error("failed to join default community", "error", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}

	o.writeProtoJson(w, &communityserverv1.JoinServerResponse{})
}

func (o *Routes) joinDefaultCommunity(ctx context.Context, memberId uuid.UUID) error {
	community, err := o.communityDb.GetDefaultCommunity(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get the default community: %w", err)
	}

	_, err = o.communityDb.UpsertCommunityMember(ctx, communitydb.UpsertCommunityMemberParams{
		ID:          uuid.New(),
		MemberID:    memberId,
		CommunityID: community.ID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to upsert community member: %w", err)
	}

	return nil
}
