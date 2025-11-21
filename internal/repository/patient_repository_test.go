package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/nathannewyen/fhir-health-interop/internal/models"
)

// setupTestDatabase creates a connection to the test database
func setupTestDatabase(t *testing.T) *sql.DB {
	// Connection string for test database (using Docker PostgreSQL)
	connectionString := "host=localhost port=5432 user=fhir_user password=fhir_password dbname=fhir_health_db sslmode=disable"

	databaseConnection, connectionError := sql.Open("postgres", connectionString)
	if connectionError != nil {
		t.Fatalf("Failed to connect to test database: %v", connectionError)
	}

	// Verify connection is working
	pingError := databaseConnection.Ping()
	if pingError != nil {
		t.Fatalf("Failed to ping test database: %v", pingError)
	}

	return databaseConnection
}

// cleanupTestData removes all test data from the patients table
func cleanupTestData(t *testing.T, databaseConnection *sql.DB) {
	_, deleteError := databaseConnection.Exec("DELETE FROM patients")
	if deleteError != nil {
		t.Fatalf("Failed to cleanup test data: %v", deleteError)
	}
}

// TestPostgresPatientRepository_Create verifies patient creation works correctly
func TestPostgresPatientRepository_Create(t *testing.T) {
	// Setup test database connection
	databaseConnection := setupTestDatabase(t)
	defer databaseConnection.Close()

	// Cleanup before and after test
	cleanupTestData(t, databaseConnection)
	defer cleanupTestData(t, databaseConnection)

	// Create a repository instance
	patientRepository := NewPostgresPatientRepository(databaseConnection)

	// Create test patient data
	birthDate := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	testPatient := &models.Patient{
		IdentifierSystem: "http://hospital.example.org/patients",
		IdentifierValue:  "TEST-001",
		Active:           true,
		FamilyName:       "TestFamily",
		GivenName:        "TestGiven",
		Gender:           "male",
		BirthDate:        &birthDate,
	}

	// Execute create operation
	createdPatient, createError := patientRepository.Create(context.Background(), testPatient)

	// Verify no error occurred
	if createError != nil {
		t.Fatalf("Failed to create patient: %v", createError)
	}

	// Verify patient ID was generated
	if createdPatient.ID == "" {
		t.Error("Expected patient ID to be generated")
	}

	// Verify timestamps were set
	if createdPatient.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Verify data matches
	if createdPatient.FamilyName != testPatient.FamilyName {
		t.Errorf("Expected family name %s, got %s", testPatient.FamilyName, createdPatient.FamilyName)
	}
}

// TestPostgresPatientRepository_GetByID verifies patient retrieval by ID works correctly
func TestPostgresPatientRepository_GetByID(t *testing.T) {
	// Setup test database connection
	databaseConnection := setupTestDatabase(t)
	defer databaseConnection.Close()

	// Cleanup before and after test
	cleanupTestData(t, databaseConnection)
	defer cleanupTestData(t, databaseConnection)

	// Create repository instance
	patientRepository := NewPostgresPatientRepository(databaseConnection)

	// Create test patient first
	birthDate := time.Date(1985, 6, 20, 0, 0, 0, 0, time.UTC)
	testPatient := &models.Patient{
		IdentifierSystem: "http://hospital.example.org/patients",
		IdentifierValue:  "TEST-002",
		Active:           true,
		FamilyName:       "Smith",
		GivenName:        "John",
		Gender:           "male",
		BirthDate:        &birthDate,
	}

	createdPatient, _ := patientRepository.Create(context.Background(), testPatient)

	// Execute GetByID operation
	retrievedPatient, getError := patientRepository.GetByID(context.Background(), createdPatient.ID)

	// Verify no error occurred
	if getError != nil {
		t.Fatalf("Failed to get patient by ID: %v", getError)
	}

	// Verify correct patient was retrieved
	if retrievedPatient.ID != createdPatient.ID {
		t.Errorf("Expected patient ID %s, got %s", createdPatient.ID, retrievedPatient.ID)
	}

	if retrievedPatient.FamilyName != testPatient.FamilyName {
		t.Errorf("Expected family name %s, got %s", testPatient.FamilyName, retrievedPatient.FamilyName)
	}
}

// TestPostgresPatientRepository_GetByID_NotFound verifies error handling for non-existent patient
func TestPostgresPatientRepository_GetByID_NotFound(t *testing.T) {
	// Setup test database connection
	databaseConnection := setupTestDatabase(t)
	defer databaseConnection.Close()

	// Create repository instance
	patientRepository := NewPostgresPatientRepository(databaseConnection)

	// Try to get non-existent patient
	nonExistentID := "00000000-0000-0000-0000-000000000000"
	_, getError := patientRepository.GetByID(context.Background(), nonExistentID)

	// Verify error occurred (sql.ErrNoRows expected)
	if getError == nil {
		t.Error("Expected error when getting non-existent patient")
	}
}

// TestPostgresPatientRepository_GetAll verifies listing all patients works correctly
func TestPostgresPatientRepository_GetAll(t *testing.T) {
	// Setup test database connection
	databaseConnection := setupTestDatabase(t)
	defer databaseConnection.Close()

	// Cleanup before and after test
	cleanupTestData(t, databaseConnection)
	defer cleanupTestData(t, databaseConnection)

	// Create repository instance
	patientRepository := NewPostgresPatientRepository(databaseConnection)

	// Create multiple test patients
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		testPatient := &models.Patient{
			Active:     true,
			FamilyName: "TestFamily",
			GivenName:  "TestGiven",
			Gender:     "male",
			BirthDate:  &birthDate,
		}
		patientRepository.Create(context.Background(), testPatient)
	}

	// Execute GetAll operation
	patients, getAllError := patientRepository.GetAll(context.Background(), 10, 0)

	// Verify no error occurred
	if getAllError != nil {
		t.Fatalf("Failed to get all patients: %v", getAllError)
	}

	// Verify correct number of patients returned
	if len(patients) != 3 {
		t.Errorf("Expected 3 patients, got %d", len(patients))
	}
}

// TestPostgresPatientRepository_Update verifies patient update works correctly
func TestPostgresPatientRepository_Update(t *testing.T) {
	// Setup test database connection
	databaseConnection := setupTestDatabase(t)
	defer databaseConnection.Close()

	// Cleanup before and after test
	cleanupTestData(t, databaseConnection)
	defer cleanupTestData(t, databaseConnection)

	// Create repository instance
	patientRepository := NewPostgresPatientRepository(databaseConnection)

	// Create test patient first
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	testPatient := &models.Patient{
		Active:     true,
		FamilyName: "OriginalName",
		GivenName:  "Original",
		Gender:     "male",
		BirthDate:  &birthDate,
	}

	createdPatient, _ := patientRepository.Create(context.Background(), testPatient)

	// Update patient data
	createdPatient.FamilyName = "UpdatedName"
	createdPatient.Active = false

	// Execute Update operation
	updatedPatient, updateError := patientRepository.Update(context.Background(), createdPatient)

	// Verify no error occurred
	if updateError != nil {
		t.Fatalf("Failed to update patient: %v", updateError)
	}

	// Verify data was updated
	if updatedPatient.FamilyName != "UpdatedName" {
		t.Errorf("Expected family name UpdatedName, got %s", updatedPatient.FamilyName)
	}

	if updatedPatient.Active != false {
		t.Error("Expected Active to be false")
	}
}

// TestPostgresPatientRepository_Delete verifies patient deletion works correctly
func TestPostgresPatientRepository_Delete(t *testing.T) {
	// Setup test database connection
	databaseConnection := setupTestDatabase(t)
	defer databaseConnection.Close()

	// Cleanup before and after test
	cleanupTestData(t, databaseConnection)
	defer cleanupTestData(t, databaseConnection)

	// Create repository instance
	patientRepository := NewPostgresPatientRepository(databaseConnection)

	// Create test patient first
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	testPatient := &models.Patient{
		Active:     true,
		FamilyName: "ToDelete",
		GivenName:  "Patient",
		Gender:     "female",
		BirthDate:  &birthDate,
	}

	createdPatient, _ := patientRepository.Create(context.Background(), testPatient)

	// Execute Delete operation
	deleteError := patientRepository.Delete(context.Background(), createdPatient.ID)

	// Verify no error occurred
	if deleteError != nil {
		t.Fatalf("Failed to delete patient: %v", deleteError)
	}

	// Verify patient no longer exists
	_, getError := patientRepository.GetByID(context.Background(), createdPatient.ID)
	if getError == nil {
		t.Error("Expected error when getting deleted patient")
	}
}
