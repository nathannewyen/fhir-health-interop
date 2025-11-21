package database

import (
	"database/sql"
	"fmt"

	// PostgreSQL driver import for side effects (registers the driver)
	_ "github.com/lib/pq"
)

// PostgresConfig holds the configuration for PostgreSQL connection
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewPostgresConnection creates and returns a new PostgreSQL database connection
func NewPostgresConnection(config PostgresConfig) (*sql.DB, error) {
	// Build the connection string using the provided configuration
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
	)

	// Open a connection to the database
	databaseConnection, openError := sql.Open("postgres", connectionString)
	if openError != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", openError)
	}

	// Verify the connection is working by pinging the database
	pingError := databaseConnection.Ping()
	if pingError != nil {
		return nil, fmt.Errorf("failed to ping database: %w", pingError)
	}

	return databaseConnection, nil
}
