package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/nathannewyen/fhir-health-interop/internal/models"
	"github.com/nathannewyen/fhir-health-interop/internal/service"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// MockPatientRepository implements repository.PatientRepository for handler tests
type MockPatientRepository struct {
	patients     map[string]*models.Patient
	createError  error
	getByIDError error
	getAllError  error
}

func NewMockPatientRepository() *MockPatientRepository {
	return &MockPatientRepository{
		patients: make(map[string]*models.Patient),
	}
}

func (mock *MockPatientRepository) Create(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if mock.createError != nil {
		return nil, mock.createError
	}
	patient.ID = "created-uuid-123"
	mock.patients[patient.ID] = patient
	return patient, nil
}

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

func (mock *MockPatientRepository) Update(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	return patient, nil
}

func (mock *MockPatientRepository) Search(ctx context.Context, searchParams *models.PatientSearchParams) ([]*models.Patient, error) {
	if mock.getAllError != nil {
		return nil, mock.getAllError
	}
	result := make([]*models.Patient, 0, len(mock.patients))
	for _, patient := range mock.patients {
		result = append(result, patient)
	}
	return result, nil
}

func (mock *MockPatientRepository) Delete(ctx context.Context, patientID string) error {
	return nil
}

// TestPatientHandler_Create_Success verifies POST /fhir/Patient creates a patient
func TestPatientHandler_Create_Success(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	// Create FHIR Patient JSON body
	requestBody := `{
		"resourceType": "Patient",
		"active": true,
		"name": [{"family": "Smith", "given": ["John"]}],
		"gender": "male",
		"birthDate": "1990-01-15"
	}`

	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString(requestBody))
	request.Header.Set("Content-Type", "application/fhir+json")
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	// Verify status code
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", recorder.Code)
	}

	// Verify content type
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/fhir+json" {
		t.Errorf("Expected Content-Type application/fhir+json, got %s", contentType)
	}

	// Verify response body contains patient
	var responsePatient fhir.Patient
	decodeError := json.NewDecoder(recorder.Body).Decode(&responsePatient)
	if decodeError != nil {
		t.Fatalf("Failed to decode response: %v", decodeError)
	}

	if responsePatient.Id == nil || *responsePatient.Id == "" {
		t.Error("Expected response to contain patient ID")
	}
}

// TestPatientHandler_Create_InvalidJSON verifies error on invalid JSON
func TestPatientHandler_Create_InvalidJSON(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString("invalid json"))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestPatientHandler_Create_DatabaseError verifies error handling on database failure
func TestPatientHandler_Create_DatabaseError(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	mockRepo.createError = errors.New("database connection failed")
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	requestBody := `{"name": [{"family": "Test"}]}`
	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestPatientHandler_GetByID_Success verifies GET /fhir/Patient/{id}
func TestPatientHandler_GetByID_Success(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	mockRepo.patients["test-uuid-123"] = &models.Patient{
		ID:         "test-uuid-123",
		FamilyName: "Johnson",
		GivenName:  "Jane",
		Gender:     "female",
		Active:     true,
	}
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	// Create request with Chi URL parameter
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient/test-uuid-123", nil)
	recorder := httptest.NewRecorder()

	// Set up Chi context with URL parameter
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "test-uuid-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	handler.GetByID(recorder, request)

	// Verify status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	// Verify response body
	var responsePatient fhir.Patient
	decodeError := json.NewDecoder(recorder.Body).Decode(&responsePatient)
	if decodeError != nil {
		t.Fatalf("Failed to decode response: %v", decodeError)
	}

	if *responsePatient.Id != "test-uuid-123" {
		t.Errorf("Expected ID test-uuid-123, got %s", *responsePatient.Id)
	}
}

// TestPatientHandler_GetByID_NotFound verifies 404 for non-existent patient
func TestPatientHandler_GetByID_NotFound(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient/non-existent", nil)
	recorder := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "non-existent")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", recorder.Code)
	}
}

// TestPatientHandler_GetByID_MissingID verifies 400 for missing ID parameter
func TestPatientHandler_GetByID_MissingID(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient/", nil)
	recorder := httptest.NewRecorder()

	// Empty Chi context (no URL param)
	routeContext := chi.NewRouteContext()
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestPatientHandler_GetAll_Success verifies GET /fhir/Patient returns all patients
func TestPatientHandler_GetAll_Success(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	mockRepo.patients["uuid-1"] = &models.Patient{ID: "uuid-1", FamilyName: "Smith"}
	mockRepo.patients["uuid-2"] = &models.Patient{ID: "uuid-2", FamilyName: "Johnson"}
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	recorder := httptest.NewRecorder()

	handler.GetAll(recorder, request)

	// Verify status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	// Verify response is array
	var responsePatients []*fhir.Patient
	decodeError := json.NewDecoder(recorder.Body).Decode(&responsePatients)
	if decodeError != nil {
		t.Fatalf("Failed to decode response: %v", decodeError)
	}

	if len(responsePatients) != 2 {
		t.Errorf("Expected 2 patients, got %d", len(responsePatients))
	}
}

// TestPatientHandler_GetAll_Empty verifies GET /fhir/Patient with no patients
func TestPatientHandler_GetAll_Empty(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	recorder := httptest.NewRecorder()

	handler.GetAll(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var responsePatients []*fhir.Patient
	json.NewDecoder(recorder.Body).Decode(&responsePatients)

	if len(responsePatients) != 0 {
		t.Errorf("Expected 0 patients, got %d", len(responsePatients))
	}
}

// TestPatientHandler_GetAll_DatabaseError verifies error handling
func TestPatientHandler_GetAll_DatabaseError(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	mockRepo.getAllError = errors.New("database error")
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	recorder := httptest.NewRecorder()

	handler.GetAll(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestNewPatientHandlerWithService verifies constructor
func TestNewPatientHandlerWithService(t *testing.T) {
	mockRepo := NewMockPatientRepository()
	patientService := service.NewPatientService(mockRepo)
	handler := NewPatientHandlerWithService(patientService)

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}
	if handler.patientService == nil {
		t.Error("Expected handler.patientService to be non-nil")
	}
}
