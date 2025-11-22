package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConfig holds MongoDB connection configuration
type MongoConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewMongoConnection creates a new MongoDB connection
func NewMongoConnection(config MongoConfig) (*mongo.Database, error) {
	// Build connection URI
	connectionURI := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
	)

	// Set client options
	clientOptions := options.Client().ApplyURI(connectionURI)

	// Create context with timeout for connection
	connectionContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, connectionError := mongo.Connect(connectionContext, clientOptions)
	if connectionError != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", connectionError)
	}

	// Ping the database to verify connection
	pingContext, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	pingError := client.Ping(pingContext, nil)
	if pingError != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", pingError)
	}

	// Return database instance
	database := client.Database(config.Database)
	return database, nil
}
