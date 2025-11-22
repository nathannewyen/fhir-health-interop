package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	apperrors "github.com/nathannewyen/fhir-health-interop/internal/errors"
	"github.com/nathannewyen/fhir-health-interop/internal/middleware"
	"github.com/nathannewyen/fhir-health-interop/internal/service"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// PatientHandler handles Patient FHIR resource requests
type PatientHandler struct {
	patientService *service.PatientService
}

// NewPatientHandler creates a new instance of PatientHandler
func NewPatientHandler() *PatientHandler {
	return &PatientHandler{}
}

// NewPatientHandlerWithService creates a PatientHandler with a service layer
func NewPatientHandlerWithService(patientService *service.PatientService) *PatientHandler {
	return &PatientHandler{
		patientService: patientService,
	}
}

// Create handles POST /fhir/Patient - creates a new patient
func (handler *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Parse the FHIR Patient from request body
	var fhirPatient fhir.Patient
	decodeError := json.NewDecoder(r.Body).Decode(&fhirPatient)
	if decodeError != nil {
		middleware.WriteError(w, r, apperrors.InvalidInput("body", "Invalid FHIR Patient JSON"))
		return
	}

	// Create patient using service layer
	createdPatient, createError := handler.patientService.CreatePatient(r.Context(), &fhirPatient)
	if createError != nil {
		middleware.WriteError(w, r, apperrors.Internal("Failed to create patient", createError))
		return
	}

	// Return created patient with 201 status
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPatient)
}

// GetByID handles GET /fhir/Patient/{id} - retrieves a patient by ID
func (handler *PatientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract patient ID from URL path
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		middleware.WriteError(w, r, apperrors.ValidationError("Patient ID is required"))
		return
	}

	// Get patient using service layer
	fhirPatient, getError := handler.patientService.GetPatientByID(r.Context(), patientID)
	if getError != nil {
		middleware.WriteError(w, r, apperrors.NotFound("Patient", patientID))
		return
	}

	// Return patient
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fhirPatient)
}

// GetAll handles GET /fhir/Patient - retrieves all patients
func (handler *PatientHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Get all patients with default pagination
	fhirPatients, getError := handler.patientService.GetAllPatients(r.Context(), 100, 0)
	if getError != nil {
		middleware.WriteError(w, r, apperrors.Internal("Failed to retrieve patients", getError))
		return
	}

	// Return patients as FHIR Bundle (simplified - just array for now)
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fhirPatients)
}

// GetSamplePatient returns a hardcoded sample FHIR Patient resource
// This endpoint demonstrates the FHIR Patient structure before database integration
func (h *PatientHandler) GetSamplePatient(w http.ResponseWriter, r *http.Request) {
	// Create a sample FHIR Patient resource with hardcoded data
	// Using pointers for optional FHIR fields as per the library spec
	patientIdentifier := "12345"
	patientIdentifierSystem := "http://hospital.example.org/patients"
	patientFamilyName := "Smith"
	patientGivenName := "John"
	patientGender := fhir.AdministrativeGenderMale
	patientBirthDate := "1990-01-15"
	patientActive := true

	// Build the FHIR Patient resource following R4 specification
	// Note: ResourceType is automatically handled by the FHIR library marshaling
	samplePatient := fhir.Patient{
		Id:     &patientIdentifier,
		Active: &patientActive,
		Name: []fhir.HumanName{
			{
				Family: &patientFamilyName,
				Given:  []string{patientGivenName},
			},
		},
		Gender:    &patientGender,
		BirthDate: &patientBirthDate,
		Identifier: []fhir.Identifier{
			{
				System: &patientIdentifierSystem,
				Value:  &patientIdentifier,
			},
		},
	}

	// Set FHIR-compliant response headers
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)

	// Encode and return the Patient resource as JSON
	json.NewEncoder(w).Encode(samplePatient)
}
