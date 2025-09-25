# Migrations

## Format
Migrations are executed using golang-migrate, and are named using timestamps. The reason for using timestamps over a numbering system is for `sqlc` compatibility, which requires migrations to be in their correct order when sorted.

## Backwards and forwards compatibility

When writing migrations, it's *crucial* to consider the following:

1. **Old code can still run even after the migration completes.** This allows canary releases to be feasible. This means you must avoid data migrations that modify existing columns unexpectedly.

2. **Strongly prefer to support multiple versions of the schema in code over writing a data migrations**, which can be time-consuming and bug-prone.


