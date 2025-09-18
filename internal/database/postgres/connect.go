package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func Connect(ctx context.Context, user, password, host, port, databaseName string) (*pgxpool.Pool, error) {
	err := applyMigrations(ctx, user, password, host, port, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed applying migrations: %w", err)
	}

	connectionString := newPooledConnectionString(user, password, host, port, databaseName, 100)

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to postgres: %w", err)
	}

	return pool, nil
}
