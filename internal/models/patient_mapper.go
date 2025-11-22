package models

import (
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// PatientMapper handles conversion between domain Patient model and FHIR Patient resource
type PatientMapper struct{}

// NewPatientMapper creates a new instance of PatientMapper
func NewPatientMapper() *PatientMapper {
	return &PatientMapper{}
}

// ToFHIR converts a domain Patient model to a FHIR Patient resource
func (mapper *PatientMapper) ToFHIR(patient *Patient) *fhir.Patient {
	// Convert gender string to FHIR AdministrativeGender enum
	var fhirGender *fhir.AdministrativeGender
	if patient.Gender != "" {
		gender := mapGenderToFHIR(patient.Gender)
		fhirGender = &gender
	}

	// Convert birthdate to FHIR format (YYYY-MM-DD string)
	var fhirBirthDate *string
	if patient.BirthDate != nil {
		birthDateString := patient.BirthDate.Format("2006-01-02")
		fhirBirthDate = &birthDateString
	}

	// Build the FHIR Patient resource
	fhirPatient := &fhir.Patient{
		Id:     &patient.ID,
		Active: &patient.Active,
		Name: []fhir.HumanName{
			{
				Family: &patient.FamilyName,
				Given:  []string{patient.GivenName},
			},
		},
		Gender:    fhirGender,
		BirthDate: fhirBirthDate,
	}

	// Add identifier if present
	if patient.IdentifierSystem != "" && patient.IdentifierValue != "" {
		fhirPatient.Identifier = []fhir.Identifier{
			{
				System: &patient.IdentifierSystem,
				Value:  &patient.IdentifierValue,
			},
		}
	}

	return fhirPatient
}

// FromFHIR converts a FHIR Patient resource to a domain Patient model
func (mapper *PatientMapper) FromFHIR(fhirPatient *fhir.Patient) *Patient {
	patient := &Patient{}

	// Map ID
	if fhirPatient.Id != nil {
		patient.ID = *fhirPatient.Id
	}

	// Map Active status
	if fhirPatient.Active != nil {
		patient.Active = *fhirPatient.Active
	}

	// Map Name (take first name entry)
	if len(fhirPatient.Name) > 0 {
		if fhirPatient.Name[0].Family != nil {
			patient.FamilyName = *fhirPatient.Name[0].Family
		}
		if len(fhirPatient.Name[0].Given) > 0 {
			patient.GivenName = fhirPatient.Name[0].Given[0]
		}
	}

	// Map Gender
	if fhirPatient.Gender != nil {
		patient.Gender = mapGenderFromFHIR(*fhirPatient.Gender)
	}

	// Map BirthDate
	if fhirPatient.BirthDate != nil {
		parsedDate, parseError := time.Parse("2006-01-02", *fhirPatient.BirthDate)
		if parseError == nil {
			patient.BirthDate = &parsedDate
		}
	}

	// Map Identifier (take first identifier entry)
	if len(fhirPatient.Identifier) > 0 {
		if fhirPatient.Identifier[0].System != nil {
			patient.IdentifierSystem = *fhirPatient.Identifier[0].System
		}
		if fhirPatient.Identifier[0].Value != nil {
			patient.IdentifierValue = *fhirPatient.Identifier[0].Value
		}
	}

	return patient
}

// mapGenderToFHIR converts a string gender to FHIR AdministrativeGender enum
func mapGenderToFHIR(gender string) fhir.AdministrativeGender {
	switch gender {
	case "male":
		return fhir.AdministrativeGenderMale
	case "female":
		return fhir.AdministrativeGenderFemale
	case "other":
		return fhir.AdministrativeGenderOther
	default:
		return fhir.AdministrativeGenderUnknown
	}
}

// mapGenderFromFHIR converts FHIR AdministrativeGender enum to string
func mapGenderFromFHIR(gender fhir.AdministrativeGender) string {
	switch gender {
	case fhir.AdministrativeGenderMale:
		return "male"
	case fhir.AdministrativeGenderFemale:
		return "female"
	case fhir.AdministrativeGenderOther:
		return "other"
	default:
		return "unknown"
	}
}
