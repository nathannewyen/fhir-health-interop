package models

import "time"

// PatientSearchParams contains filter criteria for patient search
type PatientSearchParams struct {
	// Name searches both given_name and family_name (partial match, case-insensitive)
	Name string

	// FamilyName searches family_name only (partial match, case-insensitive)
	FamilyName string

	// GivenName searches given_name only (partial match, case-insensitive)
	GivenName string

	// Gender filters by exact gender match
	Gender string

	// BirthDate filters by exact birth date
	BirthDate *time.Time

	// BirthDateGreaterThan filters birth dates greater than or equal to this value
	BirthDateGreaterThan *time.Time

	// BirthDateLessThan filters birth dates less than or equal to this value
	BirthDateLessThan *time.Time

	// Active filters by active status (nil means no filter)
	Active *bool

	// SortBy specifies the field to sort by (name, birthdate, etc.)
	SortBy string

	// SortOrder specifies ascending (asc) or descending (desc)
	SortOrder string

	// Limit specifies maximum number of results to return
	Limit int

	// Offset specifies number of results to skip (for pagination)
	Offset int
}

// ObservationSearchParams contains filter criteria for observation search
type ObservationSearchParams struct {
	// PatientID filters observations for a specific patient
	PatientID string

	// Code filters by observation code (exact match)
	Code string

	// Category filters by observation category
	Category string

	// Status filters by observation status (final, preliminary, etc.)
	Status string

	// DateGreaterThan filters observations with effective date >= this value
	DateGreaterThan *time.Time

	// DateLessThan filters observations with effective date <= this value
	DateLessThan *time.Time

	// SortBy specifies the field to sort by (effective_date, code, etc.)
	SortBy string

	// SortOrder specifies ascending (asc) or descending (desc)
	SortOrder string

	// Limit specifies maximum number of results to return
	Limit int

	// Offset specifies number of results to skip (for pagination)
	Offset int
}
