package postgres

import "fmt"

// newConnectionString creates a Postgres connection string
func newConnectionString(user, password, host, port, databaseName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, databaseName)
}

// newPooledConnectionString creates a Postgres connection string with connection pooling
func newPooledConnectionString(user, password, host, port, databaseName string, maxConnections int32) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=%d", user, password, host, port, databaseName, maxConnections)
}

// newConnectionStringWithoutDatabase creates a Postgres connection string without a database
func newConnectionStringWithoutDatabase(user, password, host, port string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s?sslmode=disable", user, password, host, port)
}
