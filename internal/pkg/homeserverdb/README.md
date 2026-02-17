# Postgres

## Migrations

### Create migration
```bash
MIGRATION_NAME="" && docker run --mount "type=bind,source=./internal/database/postgres/migrations,target=/migrations" migrate/migrate create -ext sql -dir /migrations $MIGRATION_NAME
```