package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// PatientHandler handles Patient FHIR resource requests
type PatientHandler struct{}

// NewPatientHandler creates a new instance of PatientHandler
func NewPatientHandler() *PatientHandler {
	return &PatientHandler{}
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
