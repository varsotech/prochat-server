package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
