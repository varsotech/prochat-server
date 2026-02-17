package homeserverdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

const (
	maxConnections = 100
)

// Connect opens a connection pool to a homeserverdb database
func Connect(ctx context.Context, user, password, host, port, databaseName, sslMode string) (*pgxpool.Pool, error) {
	connectionString := newConnectionString(user, password, host, port, "postgres", sslMode, maxConnections)
	err := createDatabaseIfNotExists(ctx, connectionString, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed creating database: %w", err)
	}

	connectionString = newConnectionString(user, password, host, port, databaseName, sslMode, 0)
	err = applyMigrations(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed applying migrations: %w", err)
	}

	connectionString = newConnectionString(user, password, host, port, databaseName, sslMode, maxConnections)
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to homeserverdb: %w", err)
	}

	return pool, nil
}

// newConnectionString creates a Postgres connection string
func newConnectionString(user, password, host, port, databaseName, sslMode string, maxConnections int) string {
	database := ""
	if databaseName != "" {
		database = "/" + databaseName
	}

	maxConn := ""
	if maxConnections > 0 {
		maxConn = fmt.Sprintf("&pool_max_conns=%d", maxConnections)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s%s?sslmode=%s%s", user, password, host, port, database, sslMode, maxConn)
}
