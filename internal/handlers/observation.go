package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	apperrors "github.com/nathannewyen/fhir-health-interop/internal/errors"
	"github.com/nathannewyen/fhir-health-interop/internal/middleware"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// ObservationServiceInterface defines the observation service contract
type ObservationServiceInterface interface {
	CreateObservation(ctx context.Context, fhirObservation *fhir.Observation) (*fhir.Observation, error)
	GetObservationByID(ctx context.Context, observationID string) (*fhir.Observation, error)
	GetObservationsByPatientID(ctx context.Context, patientID string, limit int, offset int) ([]*fhir.Observation, error)
	GetAllObservations(ctx context.Context, limit int, offset int) ([]*fhir.Observation, error)
	UpdateObservation(ctx context.Context, observationID string, fhirObservation *fhir.Observation) (*fhir.Observation, error)
	DeleteObservation(ctx context.Context, observationID string) error
}

// ObservationHandler handles Observation FHIR resource requests
type ObservationHandler struct {
	observationService ObservationServiceInterface
}

// NewObservationHandler creates a new observation handler instance
func NewObservationHandler(observationService ObservationServiceInterface) *ObservationHandler {
	return &ObservationHandler{
		observationService: observationService,
	}
}

// Create handles POST /fhir/Observation - creates a new observation
func (handler *ObservationHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Parse the FHIR Observation from request body
	var fhirObservation fhir.Observation
	decodeError := json.NewDecoder(r.Body).Decode(&fhirObservation)
	if decodeError != nil {
		middleware.WriteError(w, r, apperrors.InvalidInput("body", "Invalid FHIR Observation JSON"))
		return
	}

	// Create observation using service layer
	createdObservation, createError := handler.observationService.CreateObservation(r.Context(), &fhirObservation)
	if createError != nil {
		middleware.WriteError(w, r, apperrors.Internal("Failed to create observation", createError))
		return
	}

	// Return created observation with 201 status
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdObservation)
}

// GetByID handles GET /fhir/Observation/{id} - retrieves an observation by ID
func (handler *ObservationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract observation ID from URL path
	observationID := chi.URLParam(r, "id")
	if observationID == "" {
		middleware.WriteError(w, r, apperrors.ValidationError("Observation ID is required"))
		return
	}

	// Get observation using service layer
	fhirObservation, getError := handler.observationService.GetObservationByID(r.Context(), observationID)
	if getError != nil {
		middleware.WriteError(w, r, apperrors.NotFound("Observation", observationID))
		return
	}

	// Return observation
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fhirObservation)
}

// GetByPatientID handles GET /fhir/Observation?patient={id} - retrieves observations for a patient
func (handler *ObservationHandler) GetByPatientID(w http.ResponseWriter, r *http.Request) {
	// Extract patient ID from query parameter
	patientID := r.URL.Query().Get("patient")
	if patientID == "" {
		middleware.WriteError(w, r, apperrors.ValidationError("Patient ID query parameter is required"))
		return
	}

	// Get observations using service layer with default pagination
	fhirObservations, getError := handler.observationService.GetObservationsByPatientID(r.Context(), patientID, 100, 0)
	if getError != nil {
		middleware.WriteError(w, r, apperrors.Internal("Failed to retrieve observations", getError))
		return
	}

	// Return observations
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fhirObservations)
}

// GetAll handles GET /fhir/Observation - retrieves all observations
func (handler *ObservationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Check if patient query parameter exists
	if r.URL.Query().Has("patient") {
		handler.GetByPatientID(w, r)
		return
	}

	// Get all observations with default pagination
	fhirObservations, getError := handler.observationService.GetAllObservations(r.Context(), 100, 0)
	if getError != nil {
		middleware.WriteError(w, r, apperrors.Internal("Failed to retrieve observations", getError))
		return
	}

	// Return observations
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fhirObservations)
}
