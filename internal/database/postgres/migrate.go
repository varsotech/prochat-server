package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

//go:embed migrations/*.sql
var fs embed.FS

func applyMigrations(ctx context.Context, user, password, host, port, databaseName string) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("failed opening migrations dir: %w", err)
	}

	err = createDatabaseIfNotExists(ctx, user, password, host, port, databaseName)
	if err != nil {
		return fmt.Errorf("failed creating database: %w", err)
	}

	connectionString := newConnectionString(user, password, host, port, databaseName)

	m, err := migrate.NewWithSourceInstance("iofs", d, connectionString)
	if err != nil {
		return fmt.Errorf("failed creating new migrations source: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed applying migrations: %w", err)
	}

	return nil
}

func createDatabaseIfNotExists(ctx context.Context, user, password, host, port, databaseName string) error {
	connectionStringWithoutDatabase := newConnectionStringWithoutDatabase(user, password, host, port)

	db, err := pgxpool.New(ctx, connectionStringWithoutDatabase)
	if err != nil {
		return fmt.Errorf("failed connecting to postgres: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(databaseName)))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() != "duplicate_database" {
			return fmt.Errorf("failed creating database '%s': %w", databaseName, err)
		}
	}

	return nil
}
