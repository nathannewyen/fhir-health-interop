package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// MockPatientRepository implements PatientRepository interface for testing
type MockPatientRepository struct {
	patients       map[string]*models.Patient
	createError    error
	getByIDError   error
	getAllError    error
	updateError    error
	deleteError    error
	lastCreated    *models.Patient
	lastUpdated    *models.Patient
	lastDeletedID  string
}

// NewMockPatientRepository creates a new mock repository for testing
func NewMockPatientRepository() *MockPatientRepository {
	return &MockPatientRepository{
		patients: make(map[string]*models.Patient),
	}
}

// Create stores a patient and returns it with generated ID
func (mock *MockPatientRepository) Create(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if mock.createError != nil {
		return nil, mock.createError
	}
	// Simulate database behavior: generate UUID and set timestamps
	patient.ID = "generated-uuid-123"
	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	mock.patients[patient.ID] = patient
	mock.lastCreated = patient
	return patient, nil
}

// GetByID retrieves a patient by ID
func (mock *MockPatientRepository) GetByID(ctx context.Context, patientID string) (*models.Patient, error) {
	if mock.getByIDError != nil {
		return nil, mock.getByIDError
	}
	patient, exists := mock.patients[patientID]
	if !exists {
		return nil, errors.New("patient not found")
	}
	return patient, nil
}

// GetAll retrieves all patients with pagination
func (mock *MockPatientRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Patient, error) {
	if mock.getAllError != nil {
		return nil, mock.getAllError
	}
	result := make([]*models.Patient, 0, len(mock.patients))
	for _, patient := range mock.patients {
		result = append(result, patient)
	}
	return result, nil
}

// Update modifies an existing patient
func (mock *MockPatientRepository) Update(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if mock.updateError != nil {
		return nil, mock.updateError
	}
	patient.UpdatedAt = time.Now()
	mock.patients[patient.ID] = patient
	mock.lastUpdated = patient
	return patient, nil
}

// Delete removes a patient by ID
func (mock *MockPatientRepository) Delete(ctx context.Context, patientID string) error {
	if mock.deleteError != nil {
		return mock.deleteError
	}
	delete(mock.patients, patientID)
	mock.lastDeletedID = patientID
	return nil
}

// TestPatientService_CreatePatient verifies patient creation through service layer
func TestPatientService_CreatePatient(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Create FHIR Patient input
	familyName := "Smith"
	givenName := "John"
	gender := fhir.AdministrativeGenderMale
	birthDate := "1990-01-15"
	active := true

	fhirPatient := &fhir.Patient{
		Active: &active,
		Name: []fhir.HumanName{
			{
				Family: &familyName,
				Given:  []string{givenName},
			},
		},
		Gender:    &gender,
		BirthDate: &birthDate,
	}

	// Execute service method
	createdPatient, createError := patientService.CreatePatient(context.Background(), fhirPatient)

	// Verify no error
	if createError != nil {
		t.Fatalf("Expected no error, got %v", createError)
	}

	// Verify patient was created with ID
	if createdPatient.Id == nil || *createdPatient.Id == "" {
		t.Error("Expected created patient to have ID")
	}

	// Verify repository received the patient
	if mockRepo.lastCreated == nil {
		t.Error("Expected repository Create to be called")
	}

	// Verify name was preserved
	if *createdPatient.Name[0].Family != "Smith" {
		t.Errorf("Expected family name Smith, got %s", *createdPatient.Name[0].Family)
	}
}

// TestPatientService_CreatePatient_DefaultActive verifies default active status
func TestPatientService_CreatePatient_DefaultActive(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Create FHIR Patient without Active field
	familyName := "Test"
	fhirPatient := &fhir.Patient{
		Name: []fhir.HumanName{
			{Family: &familyName},
		},
	}

	// Execute service method
	_, createError := patientService.CreatePatient(context.Background(), fhirPatient)

	// Verify no error
	if createError != nil {
		t.Fatalf("Expected no error, got %v", createError)
	}

	// Verify default active was set to true in domain model
	if mockRepo.lastCreated.Active != true {
		t.Error("Expected Active to default to true")
	}
}

// TestPatientService_CreatePatient_Error verifies error handling
func TestPatientService_CreatePatient_Error(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	mockRepo.createError = errors.New("database connection failed")
	patientService := NewPatientService(mockRepo)

	familyName := "Test"
	fhirPatient := &fhir.Patient{
		Name: []fhir.HumanName{
			{Family: &familyName},
		},
	}

	// Execute service method
	createdPatient, createError := patientService.CreatePatient(context.Background(), fhirPatient)

	// Verify error returned
	if createError == nil {
		t.Error("Expected error, got nil")
	}

	// Verify no patient returned
	if createdPatient != nil {
		t.Error("Expected nil patient on error")
	}
}

// TestPatientService_GetPatientByID verifies patient retrieval
func TestPatientService_GetPatientByID(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Add test patient to mock repository
	testPatient := &models.Patient{
		ID:         "test-uuid-123",
		FamilyName: "Johnson",
		GivenName:  "Jane",
		Gender:     "female",
		Active:     true,
	}
	mockRepo.patients["test-uuid-123"] = testPatient

	// Execute service method
	fhirPatient, getError := patientService.GetPatientByID(context.Background(), "test-uuid-123")

	// Verify no error
	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}

	// Verify patient data
	if *fhirPatient.Id != "test-uuid-123" {
		t.Errorf("Expected ID test-uuid-123, got %s", *fhirPatient.Id)
	}
	if *fhirPatient.Name[0].Family != "Johnson" {
		t.Errorf("Expected family name Johnson, got %s", *fhirPatient.Name[0].Family)
	}
}

// TestPatientService_GetPatientByID_NotFound verifies not found handling
func TestPatientService_GetPatientByID_NotFound(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Execute service method for non-existent patient
	fhirPatient, getError := patientService.GetPatientByID(context.Background(), "non-existent-id")

	// Verify error returned
	if getError == nil {
		t.Error("Expected error for non-existent patient")
	}

	// Verify no patient returned
	if fhirPatient != nil {
		t.Error("Expected nil patient when not found")
	}
}

// TestPatientService_GetAllPatients verifies list retrieval
func TestPatientService_GetAllPatients(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Add test patients to mock repository
	mockRepo.patients["uuid-1"] = &models.Patient{ID: "uuid-1", FamilyName: "Smith", GivenName: "John"}
	mockRepo.patients["uuid-2"] = &models.Patient{ID: "uuid-2", FamilyName: "Johnson", GivenName: "Jane"}

	// Execute service method
	fhirPatients, getError := patientService.GetAllPatients(context.Background(), 100, 0)

	// Verify no error
	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}

	// Verify count
	if len(fhirPatients) != 2 {
		t.Errorf("Expected 2 patients, got %d", len(fhirPatients))
	}
}

// TestPatientService_UpdatePatient verifies patient update
func TestPatientService_UpdatePatient(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Add existing patient
	mockRepo.patients["update-uuid"] = &models.Patient{
		ID:         "update-uuid",
		FamilyName: "OldName",
	}

	// Create updated FHIR Patient
	newFamilyName := "NewName"
	fhirPatient := &fhir.Patient{
		Name: []fhir.HumanName{
			{Family: &newFamilyName},
		},
	}

	// Execute service method
	updatedPatient, updateError := patientService.UpdatePatient(context.Background(), "update-uuid", fhirPatient)

	// Verify no error
	if updateError != nil {
		t.Fatalf("Expected no error, got %v", updateError)
	}

	// Verify updated name
	if *updatedPatient.Name[0].Family != "NewName" {
		t.Errorf("Expected family name NewName, got %s", *updatedPatient.Name[0].Family)
	}
}

// TestPatientService_DeletePatient verifies patient deletion
func TestPatientService_DeletePatient(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	// Add patient to delete
	mockRepo.patients["delete-uuid"] = &models.Patient{ID: "delete-uuid"}

	// Execute service method
	deleteError := patientService.DeletePatient(context.Background(), "delete-uuid")

	// Verify no error
	if deleteError != nil {
		t.Fatalf("Expected no error, got %v", deleteError)
	}

	// Verify delete was called
	if mockRepo.lastDeletedID != "delete-uuid" {
		t.Errorf("Expected delete called with delete-uuid, got %s", mockRepo.lastDeletedID)
	}
}

// TestNewPatientService verifies constructor
func TestNewPatientService(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := NewPatientService(mockRepo)

	if patientService == nil {
		t.Error("Expected NewPatientService to return non-nil instance")
	}
}
