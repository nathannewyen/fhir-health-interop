package models

import (
	"time"
)

// Patient represents a patient record in the database
// This model maps to the patients table and can be converted to FHIR format
type Patient struct {
	// Unique identifier for the patient (UUID)
	ID string `json:"id"`

	// FHIR identifier system (e.g., "http://hospital.example.org/patients")
	IdentifierSystem string `json:"identifier_system"`

	// FHIR identifier value (e.g., MRN number)
	IdentifierValue string `json:"identifier_value"`

	// Whether the patient record is active
	Active bool `json:"active"`

	// Patient's family (last) name
	FamilyName string `json:"family_name"`

	// Patient's given (first) name
	GivenName string `json:"given_name"`

	// Administrative gender (male, female, other, unknown)
	Gender string `json:"gender"`

	// Patient's birth date
	BirthDate *time.Time `json:"birth_date"`

	// Audit timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
