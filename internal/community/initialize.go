package community

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varsotech/prochat-server/internal/pkg/communitydb"
)

func Initialize(ctx context.Context, postgresClient *pgxpool.Pool, host string) error {
	queries := communitydb.New(postgresClient)

	_, err := queries.UpsertDefaultCommunity(ctx, communitydb.UpsertDefaultCommunityParams{
		ID:   uuid.New(),
		Name: host,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}
