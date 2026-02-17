package communitydb

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

//go:embed migrations/*.sql
var fs embed.FS

func applyMigrations(connectionString string) error {
	driver, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("failed opening migrations dir: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", driver, connectionString)
	if err != nil {
		return fmt.Errorf("failed creating new migrations source: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed applying migrations: %w", err)
	}

	return nil
}

func createDatabaseIfNotExists(ctx context.Context, connectionStringWithoutDatabase, databaseName string) error {
	db, err := pgxpool.New(ctx, connectionStringWithoutDatabase)
	if err != nil {
		return fmt.Errorf("failed connecting to communitydb: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(databaseName)))
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			if pqErr.Code != "42P04" {
				return fmt.Errorf("failed creating database '%s' got code '%s': %w", databaseName, pqErr.Code, err)
			}
		} else {
			return fmt.Errorf("failed creating database '%s' got err: %w", databaseName, err)
		}
	}

	return nil
}
