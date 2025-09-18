package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

const (
	maxConnections = 100
)

// Connect opens a connection pool to a postgres database
func Connect(ctx context.Context, user, password, host, port, databaseName, sslMode string) (*pgxpool.Pool, error) {
	connectionStringWithoutDatabase := newConnectionString(user, password, host, port, "", sslMode, maxConnections)

	err := createDatabaseIfNotExists(ctx, connectionStringWithoutDatabase, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed creating database: %w", err)
	}

	connectionString := newConnectionString(user, password, host, port, databaseName, sslMode, maxConnections)

	err = applyMigrations(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed applying migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to postgres: %w", err)
	}

	return pool, nil
}

// newConnectionString creates a Postgres connection string
func newConnectionString(user, password, host, port, databaseName, sslMode string, maxConnections int) string {
	database := ""
	if databaseName == "" {
		database = "/" + databaseName
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s%s?sslmode=%s&pool_max_conns=%d", user, password, host, port, database, sslMode, maxConnections)
}
