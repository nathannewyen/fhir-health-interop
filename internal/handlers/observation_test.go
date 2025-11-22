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
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// MockObservationService mocks the observation service for testing
type MockObservationService struct {
	observations       map[string]*fhir.Observation
	createError        error
	getByIDError       error
	getByPatientError  error
	getAllError        error
}

func NewMockObservationService() *MockObservationService {
	return &MockObservationService{
		observations: make(map[string]*fhir.Observation),
	}
}

func (mock *MockObservationService) CreateObservation(ctx context.Context, fhirObservation *fhir.Observation) (*fhir.Observation, error) {
	if mock.createError != nil {
		return nil, mock.createError
	}
	id := "created-id-123"
	fhirObservation.Id = &id
	mock.observations[id] = fhirObservation
	return fhirObservation, nil
}

func (mock *MockObservationService) GetObservationByID(ctx context.Context, observationID string) (*fhir.Observation, error) {
	if mock.getByIDError != nil {
		return nil, mock.getByIDError
	}
	observation, exists := mock.observations[observationID]
	if !exists {
		return nil, errors.New("observation not found")
	}
	return observation, nil
}

func (mock *MockObservationService) GetObservationsByPatientID(ctx context.Context, patientID string, limit int, offset int) ([]*fhir.Observation, error) {
	if mock.getByPatientError != nil {
		return nil, mock.getByPatientError
	}
	result := make([]*fhir.Observation, 0)
	for _, observation := range mock.observations {
		if observation.Subject != nil && observation.Subject.Reference != nil {
			if *observation.Subject.Reference == "Patient/"+patientID {
				result = append(result, observation)
			}
		}
	}
	return result, nil
}

func (mock *MockObservationService) GetAllObservations(ctx context.Context, limit int, offset int) ([]*fhir.Observation, error) {
	if mock.getAllError != nil {
		return nil, mock.getAllError
	}
	result := make([]*fhir.Observation, 0, len(mock.observations))
	for _, observation := range mock.observations {
		result = append(result, observation)
	}
	return result, nil
}

func (mock *MockObservationService) UpdateObservation(ctx context.Context, observationID string, fhirObservation *fhir.Observation) (*fhir.Observation, error) {
	return fhirObservation, nil
}

func (mock *MockObservationService) SearchObservations(ctx context.Context, searchParams *models.ObservationSearchParams) ([]*fhir.Observation, error) {
	if mock.getAllError != nil {
		return nil, mock.getAllError
	}
	result := make([]*fhir.Observation, 0, len(mock.observations))
	for _, observation := range mock.observations {
		result = append(result, observation)
	}
	return result, nil
}

func (mock *MockObservationService) DeleteObservation(ctx context.Context, observationID string) error {
	return nil
}

// TestObservationHandler_Create_Success verifies observation creation
func TestObservationHandler_Create_Success(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	requestBody := `{
		"resourceType": "Observation",
		"status": "final",
		"code": {"coding": [{"code": "85354-9"}]}
	}`

	request := httptest.NewRequest(http.MethodPost, "/fhir/Observation", bytes.NewBufferString(requestBody))
	request.Header.Set("Content-Type", "application/fhir+json")
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", recorder.Code, recorder.Body.String())
	}

	var responseObservation fhir.Observation
	json.NewDecoder(recorder.Body).Decode(&responseObservation)
	if responseObservation.Id == nil || *responseObservation.Id == "" {
		t.Error("Expected response to contain observation ID")
	}
}

// TestObservationHandler_Create_InvalidJSON verifies error on invalid JSON
func TestObservationHandler_Create_InvalidJSON(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	request := httptest.NewRequest(http.MethodPost, "/fhir/Observation", bytes.NewBufferString("invalid json"))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestObservationHandler_Create_ServiceError verifies error handling
func TestObservationHandler_Create_ServiceError(t *testing.T) {
	mockService := NewMockObservationService()
	mockService.createError = errors.New("database error")
	handler := NewObservationHandler(mockService)

	requestBody := `{"status": "final", "code": {"coding": [{"code": "test"}]}}`
	request := httptest.NewRequest(http.MethodPost, "/fhir/Observation", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestObservationHandler_GetByID_Success verifies retrieval by ID
func TestObservationHandler_GetByID_Success(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	id := "test-id-123"
	code := "85354-9"
	mockService.observations["test-id-123"] = &fhir.Observation{
		Id:     &id,
		Status: fhir.ObservationStatusFinal,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{Code: &code}},
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation/test-id-123", nil)
	recorder := httptest.NewRecorder()

	// Set up Chi context
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "test-id-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var responseObservation fhir.Observation
	json.NewDecoder(recorder.Body).Decode(&responseObservation)
	if *responseObservation.Id != "test-id-123" {
		t.Errorf("Expected ID test-id-123, got %s", *responseObservation.Id)
	}
}

// TestObservationHandler_GetByID_MissingID verifies error on missing ID
func TestObservationHandler_GetByID_MissingID(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation/", nil)
	recorder := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestObservationHandler_GetByID_NotFound verifies 404 handling
func TestObservationHandler_GetByID_NotFound(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation/non-existent", nil)
	recorder := httptest.NewRecorder()

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", "non-existent")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", recorder.Code)
	}
}

// TestObservationHandler_GetByPatientID_Success verifies patient filtering
func TestObservationHandler_GetByPatientID_Success(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	id1 := "obs-1"
	id2 := "obs-2"
	patientRef := "Patient/patient-123"
	code := "test"
	mockService.observations["obs-1"] = &fhir.Observation{
		Id:      &id1,
		Status:  fhir.ObservationStatusFinal,
		Code:    fhir.CodeableConcept{Coding: []fhir.Coding{{Code: &code}}},
		Subject: &fhir.Reference{Reference: &patientRef},
	}
	mockService.observations["obs-2"] = &fhir.Observation{
		Id:      &id2,
		Status:  fhir.ObservationStatusFinal,
		Code:    fhir.CodeableConcept{Coding: []fhir.Coding{{Code: &code}}},
		Subject: &fhir.Reference{Reference: &patientRef},
	}

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?patient=patient-123", nil)
	recorder := httptest.NewRecorder()

	handler.GetByPatientID(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var responseObservations []*fhir.Observation
	json.NewDecoder(recorder.Body).Decode(&responseObservations)
	if len(responseObservations) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(responseObservations))
	}
}

// TestObservationHandler_GetByPatientID_MissingParam verifies error on missing patient param
func TestObservationHandler_GetByPatientID_MissingParam(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	recorder := httptest.NewRecorder()

	handler.GetByPatientID(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestObservationHandler_GetByPatientID_ServiceError verifies error handling
func TestObservationHandler_GetByPatientID_ServiceError(t *testing.T) {
	mockService := NewMockObservationService()
	mockService.getByPatientError = errors.New("database error")
	handler := NewObservationHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?patient=patient-123", nil)
	recorder := httptest.NewRecorder()

	handler.GetByPatientID(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestObservationHandler_GetAll_Success verifies retrieval of all observations
func TestObservationHandler_GetAll_Success(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	id1 := "obs-1"
	id2 := "obs-2"
	code := "test"
	mockService.observations["obs-1"] = &fhir.Observation{
		Id:     &id1,
		Status: fhir.ObservationStatusFinal,
		Code:   fhir.CodeableConcept{Coding: []fhir.Coding{{Code: &code}}},
	}
	mockService.observations["obs-2"] = &fhir.Observation{
		Id:     &id2,
		Status: fhir.ObservationStatusFinal,
		Code:   fhir.CodeableConcept{Coding: []fhir.Coding{{Code: &code}}},
	}

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	recorder := httptest.NewRecorder()

	handler.GetAll(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	var responseObservations []*fhir.Observation
	json.NewDecoder(recorder.Body).Decode(&responseObservations)
	if len(responseObservations) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(responseObservations))
	}
}

// TestObservationHandler_GetAll_WithPatientParam verifies patient filtering via GetAll
func TestObservationHandler_GetAll_WithPatientParam(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	id1 := "obs-1"
	patientRef := "Patient/patient-123"
	code := "test"
	mockService.observations["obs-1"] = &fhir.Observation{
		Id:      &id1,
		Status:  fhir.ObservationStatusFinal,
		Code:    fhir.CodeableConcept{Coding: []fhir.Coding{{Code: &code}}},
		Subject: &fhir.Reference{Reference: &patientRef},
	}

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?patient=patient-123", nil)
	recorder := httptest.NewRecorder()

	handler.GetAll(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

// TestObservationHandler_GetAll_ServiceError verifies error handling
func TestObservationHandler_GetAll_ServiceError(t *testing.T) {
	mockService := NewMockObservationService()
	mockService.getAllError = errors.New("database error")
	handler := NewObservationHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation", nil)
	recorder := httptest.NewRecorder()

	handler.GetAll(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}
}

// TestNewObservationHandler verifies constructor
func TestNewObservationHandler(t *testing.T) {
	mockService := NewMockObservationService()
	handler := NewObservationHandler(mockService)

	if handler == nil {
		t.Error("Expected non-nil handler")
	}
	if handler.observationService == nil {
		t.Error("Expected service to be set")
	}
}
