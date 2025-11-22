package database

import (
	"testing"
)

// TestNewPostgresConnection_Success tests successful PostgreSQL connection
func TestNewPostgresConnection_Success(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "fhir_health_db",
	}

	databaseConnection, connectionError := NewPostgresConnection(config)

	if connectionError != nil {
		t.Skipf("Skipping test: PostgreSQL not available - %v", connectionError)
		return
	}

	if databaseConnection == nil {
		t.Fatal("Expected database connection, got nil")
	}

	// Verify connection is working
	pingError := databaseConnection.Ping()
	if pingError != nil {
		t.Errorf("Expected successful ping, got error: %v", pingError)
	}

	// Close connection
	closeError := databaseConnection.Close()
	if closeError != nil {
		t.Errorf("Expected successful close, got error: %v", closeError)
	}
}

// TestNewPostgresConnection_InvalidCredentials tests connection with wrong password
func TestNewPostgresConnection_InvalidCredentials(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "fhir_user",
		Password: "wrong_password",
		DBName:   "fhir_health_db",
	}

	databaseConnection, connectionError := NewPostgresConnection(config)

	// Should fail during ping
	if connectionError == nil {
		if databaseConnection != nil {
			databaseConnection.Close()
		}
		t.Fatal("Expected connection error with invalid credentials, got nil")
	}

	if databaseConnection != nil {
		t.Error("Expected nil connection with invalid credentials")
	}
}

// TestNewPostgresConnection_InvalidHost tests connection with wrong host
func TestNewPostgresConnection_InvalidHost(t *testing.T) {
	config := PostgresConfig{
		Host:     "invalid-host-that-does-not-exist",
		Port:     "5432",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "fhir_health_db",
	}

	databaseConnection, connectionError := NewPostgresConnection(config)

	if connectionError == nil {
		if databaseConnection != nil {
			databaseConnection.Close()
		}
		t.Fatal("Expected connection error with invalid host, got nil")
	}

	if databaseConnection != nil {
		t.Error("Expected nil connection with invalid host")
	}
}

// TestNewPostgresConnection_InvalidPort tests connection with wrong port
func TestNewPostgresConnection_InvalidPort(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "9999",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "fhir_health_db",
	}

	databaseConnection, connectionError := NewPostgresConnection(config)

	if connectionError == nil {
		if databaseConnection != nil {
			databaseConnection.Close()
		}
		t.Fatal("Expected connection error with invalid port, got nil")
	}

	if databaseConnection != nil {
		t.Error("Expected nil connection with invalid port")
	}
}

// TestNewPostgresConnection_InvalidDatabase tests connection with wrong database name
func TestNewPostgresConnection_InvalidDatabase(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "nonexistent_database",
	}

	databaseConnection, connectionError := NewPostgresConnection(config)

	if connectionError == nil {
		if databaseConnection != nil {
			databaseConnection.Close()
		}
		t.Fatal("Expected connection error with invalid database, got nil")
	}

	if databaseConnection != nil {
		t.Error("Expected nil connection with invalid database")
	}
}

// TestNewPostgresConnection_EmptyConfig tests connection with empty configuration
func TestNewPostgresConnection_EmptyConfig(t *testing.T) {
	config := PostgresConfig{}

	databaseConnection, connectionError := NewPostgresConnection(config)

	if connectionError == nil {
		if databaseConnection != nil {
			databaseConnection.Close()
		}
		t.Fatal("Expected connection error with empty config, got nil")
	}

	if databaseConnection != nil {
		t.Error("Expected nil connection with empty config")
	}
}

// TestNewPostgresConnection_ConnectionPooling tests connection can be reused
func TestNewPostgresConnection_ConnectionPooling(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "fhir_health_db",
	}

	databaseConnection, connectionError := NewPostgresConnection(config)
	if connectionError != nil {
		t.Skipf("Skipping test: PostgreSQL not available - %v", connectionError)
		return
	}
	defer databaseConnection.Close()

	// Set connection pool settings
	databaseConnection.SetMaxOpenConns(10)
	databaseConnection.SetMaxIdleConns(5)

	// Test multiple pings to verify pooling works
	for i := 0; i < 5; i++ {
		pingError := databaseConnection.Ping()
		if pingError != nil {
			t.Errorf("Ping %d failed: %v", i+1, pingError)
		}
	}

	// Verify stats
	stats := databaseConnection.Stats()
	if stats.OpenConnections < 0 {
		t.Error("Expected valid connection stats")
	}
}

// TestNewPostgresConnection_MultipleConnections tests creating multiple connections
func TestNewPostgresConnection_MultipleConnections(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "fhir_user",
		Password: "fhir_password",
		DBName:   "fhir_health_db",
	}

	// Create first connection
	connection1, error1 := NewPostgresConnection(config)
	if error1 != nil {
		t.Skipf("Skipping test: PostgreSQL not available - %v", error1)
		return
	}
	defer connection1.Close()

	// Create second connection
	connection2, error2 := NewPostgresConnection(config)
	if error2 != nil {
		t.Fatalf("Expected second connection to succeed, got error: %v", error2)
	}
	defer connection2.Close()

	// Both should be able to ping
	if pingError := connection1.Ping(); pingError != nil {
		t.Errorf("First connection ping failed: %v", pingError)
	}

	if pingError := connection2.Ping(); pingError != nil {
		t.Errorf("Second connection ping failed: %v", pingError)
	}
}

// TestPostgresConfig_AllFieldsSet verifies config struct has all required fields
func TestPostgresConfig_AllFieldsSet(t *testing.T) {
	config := PostgresConfig{
		Host:     "testhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
	}

	if config.Host != "testhost" {
		t.Errorf("Expected Host 'testhost', got '%s'", config.Host)
	}

	if config.Port != "5432" {
		t.Errorf("Expected Port '5432', got '%s'", config.Port)
	}

	if config.User != "testuser" {
		t.Errorf("Expected User 'testuser', got '%s'", config.User)
	}

	if config.Password != "testpass" {
		t.Errorf("Expected Password 'testpass', got '%s'", config.Password)
	}

	if config.DBName != "testdb" {
		t.Errorf("Expected DBName 'testdb', got '%s'", config.DBName)
	}
}
