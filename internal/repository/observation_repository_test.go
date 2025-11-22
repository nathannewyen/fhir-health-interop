package repository

import (
	"context"
	"testing"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/database"
	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// setupTestMongoDB creates a test MongoDB connection
func setupTestMongoDB(t *testing.T) *mongo.Database {
	mongoConfig := database.MongoConfig{
		Host:     "localhost",
		Port:     "27017",
		User:     "fhir_user",
		Password: "fhir_password",
		Database: "admin",
	}

	mongoDatabase, mongoError := database.NewMongoConnection(mongoConfig)
	if mongoError != nil {
		t.Fatalf("Failed to connect to test MongoDB: %v", mongoError)
	}

	return mongoDatabase
}

// cleanupMongoTestData removes test data from MongoDB
func cleanupMongoTestData(t *testing.T, collection *mongo.Collection) {
	_, deleteError := collection.DeleteMany(context.Background(), bson.M{})
	if deleteError != nil {
		t.Logf("Warning: Failed to cleanup test data: %v", deleteError)
	}
}

// TestMongoObservationRepository_Create verifies observation creation
func TestMongoObservationRepository_Create(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	observation := &models.Observation{
		PatientID:     "patient-123",
		Status:        "final",
		Category:      "vital-signs",
		Code:          "85354-9",
		CodeSystem:    "http://loinc.org",
		CodeDisplay:   "Blood pressure",
	}

	createdObservation, createError := repository.Create(context.Background(), observation)

	if createError != nil {
		t.Fatalf("Expected no error, got %v", createError)
	}
	if createdObservation.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if createdObservation.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if createdObservation.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// TestMongoObservationRepository_GetByID verifies retrieval by ID
func TestMongoObservationRepository_GetByID(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	// Create test observation
	observation := &models.Observation{
		PatientID: "patient-456",
		Status:    "final",
		Code:      "test-code",
	}
	createdObservation, _ := repository.Create(context.Background(), observation)

	// Retrieve observation
	retrievedObservation, getError := repository.GetByID(context.Background(), createdObservation.ID)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if retrievedObservation.ID != createdObservation.ID {
		t.Errorf("Expected ID %s, got %s", createdObservation.ID, retrievedObservation.ID)
	}
	if retrievedObservation.PatientID != "patient-456" {
		t.Errorf("Expected patient ID patient-456, got %s", retrievedObservation.PatientID)
	}
}

// TestMongoObservationRepository_GetByID_InvalidID verifies invalid ID handling
func TestMongoObservationRepository_GetByID_InvalidID(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	retrievedObservation, getError := repository.GetByID(context.Background(), "invalid-object-id")

	if getError == nil {
		t.Error("Expected error for invalid ObjectID")
	}
	if retrievedObservation != nil {
		t.Error("Expected nil observation for invalid ID")
	}
}

// TestMongoObservationRepository_GetByID_NotFound verifies not found scenario
func TestMongoObservationRepository_GetByID_NotFound(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	// Use valid ObjectID format but non-existent ID
	nonExistentID := primitive.NewObjectID().Hex()
	retrievedObservation, getError := repository.GetByID(context.Background(), nonExistentID)

	if getError == nil {
		t.Error("Expected error for non-existent observation")
	}
	if retrievedObservation != nil {
		t.Error("Expected nil observation when not found")
	}
}

// TestMongoObservationRepository_GetByPatientID verifies patient filtering
func TestMongoObservationRepository_GetByPatientID(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	// Create multiple observations for different patients
	repository.Create(context.Background(), &models.Observation{
		PatientID: "patient-123",
		Status:    "final",
		Code:      "obs-1",
	})
	repository.Create(context.Background(), &models.Observation{
		PatientID: "patient-123",
		Status:    "final",
		Code:      "obs-2",
	})
	repository.Create(context.Background(), &models.Observation{
		PatientID: "patient-456",
		Status:    "final",
		Code:      "obs-3",
	})

	// Retrieve observations for patient-123
	observations, getError := repository.GetByPatientID(context.Background(), "patient-123", 100, 0)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if len(observations) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(observations))
	}
}

// TestMongoObservationRepository_GetByPatientID_Empty verifies empty result
func TestMongoObservationRepository_GetByPatientID_Empty(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	observations, getError := repository.GetByPatientID(context.Background(), "non-existent-patient", 100, 0)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if len(observations) != 0 {
		t.Errorf("Expected 0 observations, got %d", len(observations))
	}
}

// TestMongoObservationRepository_GetByPatientID_Pagination verifies pagination
func TestMongoObservationRepository_GetByPatientID_Pagination(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	// Create 5 observations
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond) // Ensure different timestamps
		repository.Create(context.Background(), &models.Observation{
			PatientID: "patient-123",
			Status:    "final",
			Code:      "test",
		})
	}

	// Get first 2
	observations, _ := repository.GetByPatientID(context.Background(), "patient-123", 2, 0)
	if len(observations) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(observations))
	}

	// Get next 2
	observations, _ = repository.GetByPatientID(context.Background(), "patient-123", 2, 2)
	if len(observations) != 2 {
		t.Errorf("Expected 2 observations with offset, got %d", len(observations))
	}
}

// TestMongoObservationRepository_GetAll verifies retrieval of all observations
func TestMongoObservationRepository_GetAll(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	// Create test observations
	repository.Create(context.Background(), &models.Observation{PatientID: "p1", Status: "final", Code: "test"})
	repository.Create(context.Background(), &models.Observation{PatientID: "p2", Status: "final", Code: "test"})
	repository.Create(context.Background(), &models.Observation{PatientID: "p3", Status: "final", Code: "test"})

	observations, getError := repository.GetAll(context.Background(), 100, 0)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if len(observations) != 3 {
		t.Errorf("Expected 3 observations, got %d", len(observations))
	}
}

// TestMongoObservationRepository_GetAll_Empty verifies empty collection
func TestMongoObservationRepository_GetAll_Empty(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	observations, getError := repository.GetAll(context.Background(), 100, 0)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if len(observations) != 0 {
		t.Errorf("Expected 0 observations, got %d", len(observations))
	}
}

// TestMongoObservationRepository_Update verifies update operation
func TestMongoObservationRepository_Update(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	// Create observation
	observation := &models.Observation{
		PatientID: "patient-123",
		Status:    "preliminary",
		Code:      "old-code",
	}
	createdObservation, _ := repository.Create(context.Background(), observation)

	// Update observation
	createdObservation.Status = "final"
	createdObservation.Code = "new-code"
	updatedObservation, updateError := repository.Update(context.Background(), createdObservation)

	if updateError != nil {
		t.Fatalf("Expected no error, got %v", updateError)
	}
	if updatedObservation.Status != "final" {
		t.Errorf("Expected status final, got %s", updatedObservation.Status)
	}
	if updatedObservation.Code != "new-code" {
		t.Errorf("Expected code new-code, got %s", updatedObservation.Code)
	}
}

// TestMongoObservationRepository_Update_InvalidID verifies invalid ID handling
func TestMongoObservationRepository_Update_InvalidID(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	observation := &models.Observation{
		ID:     "invalid-id",
		Status: "final",
		Code:   "test",
	}

	updatedObservation, updateError := repository.Update(context.Background(), observation)

	if updateError == nil {
		t.Error("Expected error for invalid ObjectID")
	}
	if updatedObservation != nil {
		t.Error("Expected nil observation for invalid ID")
	}
}

// TestMongoObservationRepository_Update_NotFound verifies not found scenario
func TestMongoObservationRepository_Update_NotFound(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	observation := &models.Observation{
		ID:     primitive.NewObjectID().Hex(),
		Status: "final",
		Code:   "test",
	}

	updatedObservation, updateError := repository.Update(context.Background(), observation)

	if updateError == nil {
		t.Error("Expected error for non-existent observation")
	}
	if updatedObservation != nil {
		t.Error("Expected nil observation when not found")
	}
}

// TestMongoObservationRepository_Delete verifies deletion
func TestMongoObservationRepository_Delete(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	// Create observation
	observation := &models.Observation{
		PatientID: "patient-123",
		Status:    "final",
		Code:      "test",
	}
	createdObservation, _ := repository.Create(context.Background(), observation)

	// Delete observation
	deleteError := repository.Delete(context.Background(), createdObservation.ID)

	if deleteError != nil {
		t.Fatalf("Expected no error, got %v", deleteError)
	}

	// Verify deletion
	retrievedObservation, _ := repository.GetByID(context.Background(), createdObservation.ID)
	if retrievedObservation != nil {
		t.Error("Expected observation to be deleted")
	}
}

// TestMongoObservationRepository_Delete_InvalidID verifies invalid ID handling
func TestMongoObservationRepository_Delete_InvalidID(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	deleteError := repository.Delete(context.Background(), "invalid-id")

	if deleteError == nil {
		t.Error("Expected error for invalid ObjectID")
	}
}

// TestMongoObservationRepository_Delete_NotFound verifies not found scenario
func TestMongoObservationRepository_Delete_NotFound(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	nonExistentID := primitive.NewObjectID().Hex()
	deleteError := repository.Delete(context.Background(), nonExistentID)

	if deleteError == nil {
		t.Error("Expected error for non-existent observation")
	}
}

// TestNewMongoObservationRepository verifies constructor
func TestNewMongoObservationRepository(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)

	if repository == nil {
		t.Error("Expected non-nil repository")
	}
	if repository.collection == nil {
		t.Error("Expected collection to be set")
	}
}
