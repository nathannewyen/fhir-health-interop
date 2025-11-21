package models

import (
	"testing"
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// TestPatientMapper_ToFHIR verifies conversion from domain Patient to FHIR Patient
func TestPatientMapper_ToFHIR(t *testing.T) {
	mapper := NewPatientMapper()

	// Create test patient with all fields populated
	birthDate := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	domainPatient := &Patient{
		ID:               "test-uuid-123",
		IdentifierSystem: "http://hospital.example.org/patients",
		IdentifierValue:  "MRN-001",
		Active:           true,
		FamilyName:       "Smith",
		GivenName:        "John",
		Gender:           "male",
		BirthDate:        &birthDate,
	}

	// Convert to FHIR
	fhirPatient := mapper.ToFHIR(domainPatient)

	// Verify ID
	if fhirPatient.Id == nil {
		t.Error("Expected FHIR patient ID to be non-nil")
	} else if *fhirPatient.Id != domainPatient.ID {
		t.Errorf("Expected ID %s, got %s", domainPatient.ID, *fhirPatient.Id)
	}

	// Verify Active
	if fhirPatient.Active == nil {
		t.Error("Expected FHIR patient Active to be non-nil")
	} else if *fhirPatient.Active != true {
		t.Error("Expected Active to be true")
	}

	// Verify Name
	if len(fhirPatient.Name) == 0 {
		t.Error("Expected at least one name")
	} else {
		if *fhirPatient.Name[0].Family != "Smith" {
			t.Errorf("Expected family name Smith, got %s", *fhirPatient.Name[0].Family)
		}
		if fhirPatient.Name[0].Given[0] != "John" {
			t.Errorf("Expected given name John, got %s", fhirPatient.Name[0].Given[0])
		}
	}

	// Verify Gender
	if fhirPatient.Gender == nil {
		t.Error("Expected FHIR patient Gender to be non-nil")
	} else if *fhirPatient.Gender != fhir.AdministrativeGenderMale {
		t.Errorf("Expected gender male, got %s", *fhirPatient.Gender)
	}

	// Verify BirthDate
	if fhirPatient.BirthDate == nil {
		t.Error("Expected FHIR patient BirthDate to be non-nil")
	} else if *fhirPatient.BirthDate != "1990-06-15" {
		t.Errorf("Expected birth date 1990-06-15, got %s", *fhirPatient.BirthDate)
	}

	// Verify Identifier
	if len(fhirPatient.Identifier) == 0 {
		t.Error("Expected at least one identifier")
	} else {
		if *fhirPatient.Identifier[0].System != domainPatient.IdentifierSystem {
			t.Errorf("Expected identifier system %s, got %s", domainPatient.IdentifierSystem, *fhirPatient.Identifier[0].System)
		}
		if *fhirPatient.Identifier[0].Value != domainPatient.IdentifierValue {
			t.Errorf("Expected identifier value %s, got %s", domainPatient.IdentifierValue, *fhirPatient.Identifier[0].Value)
		}
	}
}

// TestPatientMapper_FromFHIR verifies conversion from FHIR Patient to domain Patient
func TestPatientMapper_FromFHIR(t *testing.T) {
	mapper := NewPatientMapper()

	// Create FHIR patient with all fields populated
	patientID := "fhir-uuid-456"
	active := true
	familyName := "Johnson"
	givenName := "Jane"
	gender := fhir.AdministrativeGenderFemale
	birthDate := "1985-03-20"
	identifierSystem := "http://hospital.example.org/patients"
	identifierValue := "MRN-002"

	fhirPatient := &fhir.Patient{
		Id:     &patientID,
		Active: &active,
		Name: []fhir.HumanName{
			{
				Family: &familyName,
				Given:  []string{givenName},
			},
		},
		Gender:    &gender,
		BirthDate: &birthDate,
		Identifier: []fhir.Identifier{
			{
				System: &identifierSystem,
				Value:  &identifierValue,
			},
		},
	}

	// Convert to domain model
	domainPatient := mapper.FromFHIR(fhirPatient)

	// Verify ID
	if domainPatient.ID != patientID {
		t.Errorf("Expected ID %s, got %s", patientID, domainPatient.ID)
	}

	// Verify Active
	if domainPatient.Active != true {
		t.Error("Expected Active to be true")
	}

	// Verify Name
	if domainPatient.FamilyName != "Johnson" {
		t.Errorf("Expected family name Johnson, got %s", domainPatient.FamilyName)
	}
	if domainPatient.GivenName != "Jane" {
		t.Errorf("Expected given name Jane, got %s", domainPatient.GivenName)
	}

	// Verify Gender
	if domainPatient.Gender != "female" {
		t.Errorf("Expected gender female, got %s", domainPatient.Gender)
	}

	// Verify BirthDate
	if domainPatient.BirthDate == nil {
		t.Error("Expected BirthDate to be non-nil")
	} else {
		expectedDate := time.Date(1985, 3, 20, 0, 0, 0, 0, time.UTC)
		if !domainPatient.BirthDate.Equal(expectedDate) {
			t.Errorf("Expected birth date %v, got %v", expectedDate, domainPatient.BirthDate)
		}
	}

	// Verify Identifier
	if domainPatient.IdentifierSystem != identifierSystem {
		t.Errorf("Expected identifier system %s, got %s", identifierSystem, domainPatient.IdentifierSystem)
	}
	if domainPatient.IdentifierValue != identifierValue {
		t.Errorf("Expected identifier value %s, got %s", identifierValue, domainPatient.IdentifierValue)
	}
}

// TestPatientMapper_ToFHIR_NilBirthDate verifies handling of nil birth date
func TestPatientMapper_ToFHIR_NilBirthDate(t *testing.T) {
	mapper := NewPatientMapper()

	domainPatient := &Patient{
		ID:         "test-uuid",
		FamilyName: "Test",
		GivenName:  "Patient",
		BirthDate:  nil,
	}

	fhirPatient := mapper.ToFHIR(domainPatient)

	if fhirPatient.BirthDate != nil {
		t.Error("Expected BirthDate to be nil when domain patient has nil birth date")
	}
}

// TestPatientMapper_ToFHIR_EmptyIdentifier verifies handling of empty identifier
func TestPatientMapper_ToFHIR_EmptyIdentifier(t *testing.T) {
	mapper := NewPatientMapper()

	domainPatient := &Patient{
		ID:               "test-uuid",
		FamilyName:       "Test",
		GivenName:        "Patient",
		IdentifierSystem: "",
		IdentifierValue:  "",
	}

	fhirPatient := mapper.ToFHIR(domainPatient)

	if len(fhirPatient.Identifier) != 0 {
		t.Error("Expected no identifiers when domain patient has empty identifier")
	}
}

// TestGenderMapping verifies all gender mappings work correctly
func TestGenderMapping(t *testing.T) {
	testCases := []struct {
		domainGender string
		fhirGender   fhir.AdministrativeGender
	}{
		{"male", fhir.AdministrativeGenderMale},
		{"female", fhir.AdministrativeGenderFemale},
		{"other", fhir.AdministrativeGenderOther},
		{"unknown", fhir.AdministrativeGenderUnknown},
		{"invalid", fhir.AdministrativeGenderUnknown},
	}

	for _, testCase := range testCases {
		result := mapGenderToFHIR(testCase.domainGender)
		if result != testCase.fhirGender {
			t.Errorf("mapGenderToFHIR(%s): expected %v, got %v", testCase.domainGender, testCase.fhirGender, result)
		}
	}
}

// TestNewPatientMapper verifies constructor creates valid instance
func TestNewPatientMapper(t *testing.T) {
	mapper := NewPatientMapper()

	if mapper == nil {
		t.Error("Expected NewPatientMapper to return non-nil instance")
	}
}
