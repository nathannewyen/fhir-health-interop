package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// MockObservationRepository implements ObservationRepository for testing
type MockObservationRepository struct {
	observations  map[string]*models.Observation
	createError   error
	getByIDError  error
	getAllError   error
	getByPatientError error
	updateError   error
	deleteError   error
	lastCreated   *models.Observation
}

func NewMockObservationRepository() *MockObservationRepository {
	return &MockObservationRepository{
		observations: make(map[string]*models.Observation),
	}
}

func (mock *MockObservationRepository) Create(ctx context.Context, observation *models.Observation) (*models.Observation, error) {
	if mock.createError != nil {
		return nil, mock.createError
	}
	observation.ID = "generated-mongo-id-123"
	observation.CreatedAt = time.Now()
	observation.UpdatedAt = time.Now()
	mock.observations[observation.ID] = observation
	mock.lastCreated = observation
	return observation, nil
}

func (mock *MockObservationRepository) GetByID(ctx context.Context, observationID string) (*models.Observation, error) {
	if mock.getByIDError != nil {
		return nil, mock.getByIDError
	}
	observation, exists := mock.observations[observationID]
	if !exists {
		return nil, errors.New("observation not found")
	}
	return observation, nil
}

func (mock *MockObservationRepository) GetByPatientID(ctx context.Context, patientID string, limit int, offset int) ([]*models.Observation, error) {
	if mock.getByPatientError != nil {
		return nil, mock.getByPatientError
	}
	result := make([]*models.Observation, 0)
	for _, observation := range mock.observations {
		if observation.PatientID == patientID {
			result = append(result, observation)
		}
	}
	return result, nil
}

func (mock *MockObservationRepository) GetAll(ctx context.Context, limit int, offset int) ([]*models.Observation, error) {
	if mock.getAllError != nil {
		return nil, mock.getAllError
	}
	result := make([]*models.Observation, 0, len(mock.observations))
	for _, observation := range mock.observations {
		result = append(result, observation)
	}
	return result, nil
}

func (mock *MockObservationRepository) Update(ctx context.Context, observation *models.Observation) (*models.Observation, error) {
	if mock.updateError != nil {
		return nil, mock.updateError
	}
	observation.UpdatedAt = time.Now()
	mock.observations[observation.ID] = observation
	return observation, nil
}

func (mock *MockObservationRepository) Delete(ctx context.Context, observationID string) error {
	if mock.deleteError != nil {
		return mock.deleteError
	}
	delete(mock.observations, observationID)
	return nil
}

// TestObservationService_CreateObservation verifies observation creation
func TestObservationService_CreateObservation(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	code := "85354-9"
	codeSystem := "http://loinc.org"
	codeDisplay := "Blood pressure"
	patientRef := "Patient/patient-123"

	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{
				{System: &codeSystem, Code: &code, Display: &codeDisplay},
			},
		},
		Subject: &fhir.Reference{Reference: &patientRef},
	}

	createdObservation, createError := observationService.CreateObservation(context.Background(), fhirObservation)

	if createError != nil {
		t.Fatalf("Expected no error, got %v", createError)
	}
	if createdObservation.Id == nil || *createdObservation.Id == "" {
		t.Error("Expected created observation to have ID")
	}
	if mockRepo.lastCreated == nil {
		t.Error("Expected repository Create to be called")
	}
	if mockRepo.lastCreated.PatientID != "patient-123" {
		t.Errorf("Expected patient ID patient-123, got %s", mockRepo.lastCreated.PatientID)
	}
}

// TestObservationService_CreateObservation_Error verifies error handling
func TestObservationService_CreateObservation_Error(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	mockRepo.createError = errors.New("database error")
	observationService := NewObservationService(mockRepo)

	code := "test"
	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code}},
		},
	}

	createdObservation, createError := observationService.CreateObservation(context.Background(), fhirObservation)

	if createError == nil {
		t.Error("Expected error, got nil")
	}
	if createdObservation != nil {
		t.Error("Expected nil observation on error")
	}
}

// TestObservationService_GetObservationByID verifies retrieval by ID
func TestObservationService_GetObservationByID(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	testObservation := &models.Observation{
		ID:         "test-id-123",
		PatientID:  "patient-456",
		Status:     "final",
		Code:       "85354-9",
		CodeSystem: "http://loinc.org",
	}
	mockRepo.observations["test-id-123"] = testObservation

	fhirObservation, getError := observationService.GetObservationByID(context.Background(), "test-id-123")

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if *fhirObservation.Id != "test-id-123" {
		t.Errorf("Expected ID test-id-123, got %s", *fhirObservation.Id)
	}
}

// TestObservationService_GetObservationByID_NotFound verifies not found handling
func TestObservationService_GetObservationByID_NotFound(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	fhirObservation, getError := observationService.GetObservationByID(context.Background(), "non-existent")

	if getError == nil {
		t.Error("Expected error for non-existent observation")
	}
	if fhirObservation != nil {
		t.Error("Expected nil observation when not found")
	}
}

// TestObservationService_GetObservationsByPatientID verifies patient filtering
func TestObservationService_GetObservationsByPatientID(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	mockRepo.observations["obs-1"] = &models.Observation{
		ID: "obs-1", PatientID: "patient-123", Status: "final", Code: "test",
	}
	mockRepo.observations["obs-2"] = &models.Observation{
		ID: "obs-2", PatientID: "patient-123", Status: "final", Code: "test",
	}
	mockRepo.observations["obs-3"] = &models.Observation{
		ID: "obs-3", PatientID: "patient-456", Status: "final", Code: "test",
	}

	fhirObservations, getError := observationService.GetObservationsByPatientID(context.Background(), "patient-123", 100, 0)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if len(fhirObservations) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(fhirObservations))
	}
}

// TestObservationService_GetObservationsByPatientID_Error verifies error handling
func TestObservationService_GetObservationsByPatientID_Error(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	mockRepo.getByPatientError = errors.New("database error")
	observationService := NewObservationService(mockRepo)

	fhirObservations, getError := observationService.GetObservationsByPatientID(context.Background(), "patient-123", 100, 0)

	if getError == nil {
		t.Error("Expected error, got nil")
	}
	if fhirObservations != nil {
		t.Error("Expected nil on error")
	}
}

// TestObservationService_GetAllObservations verifies retrieval of all observations
func TestObservationService_GetAllObservations(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	mockRepo.observations["obs-1"] = &models.Observation{ID: "obs-1", Status: "final", Code: "test"}
	mockRepo.observations["obs-2"] = &models.Observation{ID: "obs-2", Status: "final", Code: "test"}

	fhirObservations, getError := observationService.GetAllObservations(context.Background(), 100, 0)

	if getError != nil {
		t.Fatalf("Expected no error, got %v", getError)
	}
	if len(fhirObservations) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(fhirObservations))
	}
}

// TestObservationService_GetAllObservations_Error verifies error handling
func TestObservationService_GetAllObservations_Error(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	mockRepo.getAllError = errors.New("database error")
	observationService := NewObservationService(mockRepo)

	fhirObservations, getError := observationService.GetAllObservations(context.Background(), 100, 0)

	if getError == nil {
		t.Error("Expected error, got nil")
	}
	if fhirObservations != nil {
		t.Error("Expected nil on error")
	}
}

// TestObservationService_UpdateObservation verifies update
func TestObservationService_UpdateObservation(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	mockRepo.observations["update-id"] = &models.Observation{
		ID: "update-id", Status: "preliminary", Code: "old-code",
	}

	code := "new-code"
	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code}},
		},
	}

	updatedObservation, updateError := observationService.UpdateObservation(context.Background(), "update-id", fhirObservation)

	if updateError != nil {
		t.Fatalf("Expected no error, got %v", updateError)
	}
	if updatedObservation.Status != fhir.ObservationStatusFinal {
		t.Error("Expected status to be updated")
	}
}

// TestObservationService_UpdateObservation_Error verifies error handling
func TestObservationService_UpdateObservation_Error(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	mockRepo.updateError = errors.New("database error")
	observationService := NewObservationService(mockRepo)

	code := "test"
	fhirObservation := &fhir.Observation{
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code}},
		},
	}

	updatedObservation, updateError := observationService.UpdateObservation(context.Background(), "test-id", fhirObservation)

	if updateError == nil {
		t.Error("Expected error, got nil")
	}
	if updatedObservation != nil {
		t.Error("Expected nil on error")
	}
}

// TestObservationService_DeleteObservation verifies deletion
func TestObservationService_DeleteObservation(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	mockRepo.observations["delete-id"] = &models.Observation{ID: "delete-id"}

	deleteError := observationService.DeleteObservation(context.Background(), "delete-id")

	if deleteError != nil {
		t.Fatalf("Expected no error, got %v", deleteError)
	}
	if _, exists := mockRepo.observations["delete-id"]; exists {
		t.Error("Expected observation to be deleted")
	}
}

// TestObservationService_DeleteObservation_Error verifies error handling
func TestObservationService_DeleteObservation_Error(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	mockRepo.deleteError = errors.New("database error")
	observationService := NewObservationService(mockRepo)

	deleteError := observationService.DeleteObservation(context.Background(), "test-id")

	if deleteError == nil {
		t.Error("Expected error, got nil")
	}
}

// TestNewObservationService verifies constructor
func TestNewObservationService(t *testing.T) {
	mockRepo := NewMockObservationRepository()
	observationService := NewObservationService(mockRepo)

	if observationService == nil {
		t.Error("Expected non-nil service")
	}
	if observationService.observationRepository == nil {
		t.Error("Expected repository to be set")
	}
	if observationService.observationMapper == nil {
		t.Error("Expected mapper to be set")
	}
}
