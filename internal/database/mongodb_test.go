package database

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// TestNewMongoConnection_Success tests successful MongoDB connection
func TestNewMongoConnection_Success(t *testing.T) {
	config := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	mongoDatabase, connectionError := NewMongoConnection(config)

	if connectionError != nil {
		t.Skipf("Skipping test: MongoDB not available - %v", connectionError)
		return
	}

	if mongoDatabase == nil {
		t.Fatal("Expected database instance, got nil")
	}

	// Verify we can perform a simple operation
	collection := mongoDatabase.Collection("test_collection")
	if collection == nil {
		t.Error("Expected to get collection reference")
	}

	// Test a simple ping-like operation
	commandContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	commandResult := mongoDatabase.RunCommand(commandContext, bson.D{{Key: "ping", Value: 1}})
	if commandResult.Err() != nil {
		t.Errorf("Expected successful ping command, got error: %v", commandResult.Err())
	}
}

// TestNewMongoConnection_InvalidCredentials tests connection with wrong password
func TestNewMongoConnection_InvalidCredentials(t *testing.T) {
	config := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "wrong_password",
		Database: "admin",
	}

	mongoDatabase, connectionError := NewMongoConnection(config)

	if connectionError == nil {
		t.Fatal("Expected connection error with invalid credentials, got nil")
	}

	if mongoDatabase != nil {
		t.Error("Expected nil database with invalid credentials")
	}
}

// TestNewMongoConnection_InvalidHost tests connection with wrong host
func TestNewMongoConnection_InvalidHost(t *testing.T) {
	config := MongoConfig{
		Host:     "invalid-host-that-does-not-exist",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	mongoDatabase, connectionError := NewMongoConnection(config)

	if connectionError == nil {
		t.Fatal("Expected connection error with invalid host, got nil")
	}

	if mongoDatabase != nil {
		t.Error("Expected nil database with invalid host")
	}
}

// TestNewMongoConnection_InvalidPort tests connection with wrong port
func TestNewMongoConnection_InvalidPort(t *testing.T) {
	config := MongoConfig{
		Host:     "localhost",
		Port:     "9999",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	mongoDatabase, connectionError := NewMongoConnection(config)

	if connectionError == nil {
		t.Fatal("Expected connection error with invalid port, got nil")
	}

	if mongoDatabase != nil {
		t.Error("Expected nil database with invalid port")
	}
}

// TestNewMongoConnection_ConnectionTimeout tests that connection respects timeout
func TestNewMongoConnection_ConnectionTimeout(t *testing.T) {
	config := MongoConfig{
		Host:     "192.0.2.1", // Non-routable IP to trigger timeout
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	startTime := time.Now()
	mongoDatabase, connectionError := NewMongoConnection(config)
	elapsed := time.Since(startTime)

	if connectionError == nil {
		t.Fatal("Expected connection error with unreachable host, got nil")
	}

	if mongoDatabase != nil {
		t.Error("Expected nil database with unreachable host")
	}

	// Should timeout within reasonable time (10s connection + 5s ping + some buffer)
	if elapsed > 20*time.Second {
		t.Errorf("Connection took too long: %v (expected < 20s)", elapsed)
	}
}

// TestNewMongoConnection_EmptyConfig tests connection with empty configuration
func TestNewMongoConnection_EmptyConfig(t *testing.T) {
	config := MongoConfig{}

	mongoDatabase, connectionError := NewMongoConnection(config)

	if connectionError == nil {
		t.Fatal("Expected connection error with empty config, got nil")
	}

	if mongoDatabase != nil {
		t.Error("Expected nil database with empty config")
	}
}

// TestNewMongoConnection_DifferentDatabases tests connecting to different databases
func TestNewMongoConnection_DifferentDatabases(t *testing.T) {
	config1 := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	config2 := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "fhir_health_db",
	}

	database1, error1 := NewMongoConnection(config1)
	if error1 != nil {
		t.Skipf("Skipping test: MongoDB not available - %v", error1)
		return
	}

	database2, error2 := NewMongoConnection(config2)
	if error2 != nil {
		t.Skipf("Skipping test: MongoDB not available - %v", error2)
		return
	}

	// Verify they reference different databases
	if database1.Name() != "admin" {
		t.Errorf("Expected database1 name 'admin', got '%s'", database1.Name())
	}

	if database2.Name() != "fhir_health_db" {
		t.Errorf("Expected database2 name 'fhir_health_db', got '%s'", database2.Name())
	}

	if database1.Name() == database2.Name() {
		t.Error("Expected different database names")
	}
}

// TestNewMongoConnection_CollectionOperations tests basic collection operations
func TestNewMongoConnection_CollectionOperations(t *testing.T) {
	config := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	mongoDatabase, connectionError := NewMongoConnection(config)
	if connectionError != nil {
		t.Skipf("Skipping test: MongoDB not available - %v", connectionError)
		return
	}

	// Get collection
	collection := mongoDatabase.Collection("test_collection")
	if collection == nil {
		t.Fatal("Expected collection reference, got nil")
	}

	// Verify collection name
	if collection.Name() != "test_collection" {
		t.Errorf("Expected collection name 'test_collection', got '%s'", collection.Name())
	}

	// Verify database reference
	if collection.Database().Name() != "admin" {
		t.Errorf("Expected database name 'admin', got '%s'", collection.Database().Name())
	}
}

// TestNewMongoConnection_URIFormat tests that connection URI is properly formatted
func TestNewMongoConnection_URIFormat(t *testing.T) {
	config := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "test_user",
		Password: "test_pass",
		Database: "test_db",
	}

	// This should create URI: mongodb://test_user:test_pass@localhost:27017
	// We can't directly test the URI, but we can test that it attempts connection
	_, connectionError := NewMongoConnection(config)

	// We expect an error (likely authentication failure with fake credentials)
	// but the error should be about authentication, not URI parsing
	if connectionError != nil {
		errorMessage := connectionError.Error()

		// Should contain "failed to connect" or "failed to ping"
		if errorMessage == "" {
			t.Error("Expected non-empty error message")
		}

		// Should NOT contain URI parsing errors
		if containsURIError(errorMessage) {
			t.Errorf("Expected authentication error, got URI parsing error: %v", connectionError)
		}
	}
}

// TestMongoConfig_AllFieldsSet verifies config struct has all required fields
func TestMongoConfig_AllFieldsSet(t *testing.T) {
	config := MongoConfig{
		Host:     "testhost",
		Port:     "27017",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
	}

	if config.Host != "testhost" {
		t.Errorf("Expected Host 'testhost', got '%s'", config.Host)
	}

	if config.Port != "27017" {
		t.Errorf("Expected Port '27017', got '%s'", config.Port)
	}

	if config.User != "testuser" {
		t.Errorf("Expected User 'testuser', got '%s'", config.User)
	}

	if config.Password != "testpass" {
		t.Errorf("Expected Password 'testpass', got '%s'", config.Password)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected Database 'testdb', got '%s'", config.Database)
	}
}

// TestNewMongoConnection_MultipleConnections tests creating multiple connections
func TestNewMongoConnection_MultipleConnections(t *testing.T) {
	config := MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	// Create first connection
	database1, error1 := NewMongoConnection(config)
	if error1 != nil {
		t.Skipf("Skipping test: MongoDB not available - %v", error1)
		return
	}

	// Create second connection
	database2, error2 := NewMongoConnection(config)
	if error2 != nil {
		t.Fatalf("Expected second connection to succeed, got error: %v", error2)
	}

	// Both should be able to run commands
	commandContext1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel1()

	result1 := database1.RunCommand(commandContext1, bson.D{{Key: "ping", Value: 1}})
	if result1.Err() != nil {
		t.Errorf("First connection command failed: %v", result1.Err())
	}

	commandContext2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	result2 := database2.RunCommand(commandContext2, bson.D{{Key: "ping", Value: 1}})
	if result2.Err() != nil {
		t.Errorf("Second connection command failed: %v", result2.Err())
	}
}

// Helper function to check if error message contains URI parsing errors
func containsURIError(errorMessage string) bool {
	uriErrorIndicators := []string{
		"error parsing uri",
		"invalid uri",
		"malformed uri",
		"uri must",
	}

	for _, indicator := range uriErrorIndicators {
		if len(errorMessage) >= len(indicator) {
			for i := 0; i <= len(errorMessage)-len(indicator); i++ {
				if errorMessage[i:i+len(indicator)] == indicator {
					return true
				}
			}
		}
	}

	return false
}
