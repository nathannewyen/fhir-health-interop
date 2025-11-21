package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// TestPatientHandler_GetSamplePatient verifies the sample patient endpoint returns correct FHIR data
func TestPatientHandler_GetSamplePatient(t *testing.T) {
	// Create a new patient handler instance
	patientHandler := NewPatientHandler()

	// Create a mock HTTP request for GET /fhir/Patient/sample
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient/sample", nil)

	// Create a response recorder to capture the response
	responseRecorder := httptest.NewRecorder()

	// Call the sample patient handler
	patientHandler.GetSamplePatient(responseRecorder, request)

	// Verify HTTP status code is 200 OK
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	// Verify Content-Type header is application/fhir+json
	contentType := responseRecorder.Header().Get("Content-Type")
	expectedContentType := "application/fhir+json"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Parse the JSON response body into FHIR Patient struct
	var patientResponse fhir.Patient
	decodeError := json.NewDecoder(responseRecorder.Body).Decode(&patientResponse)
	if decodeError != nil {
		t.Fatalf("Failed to decode response body: %v", decodeError)
	}

	// Verify the patient ID field
	expectedPatientId := "12345"
	if patientResponse.Id == nil {
		t.Error("Expected patient ID to be non-nil")
	} else if *patientResponse.Id != expectedPatientId {
		t.Errorf("Expected patient ID %s, got %s", expectedPatientId, *patientResponse.Id)
	}

	// Verify the patient active status
	if patientResponse.Active == nil {
		t.Error("Expected patient Active to be non-nil")
	} else if *patientResponse.Active != true {
		t.Errorf("Expected patient Active to be true, got %v", *patientResponse.Active)
	}

	// Verify the patient gender
	expectedGender := fhir.AdministrativeGenderMale
	if patientResponse.Gender == nil {
		t.Error("Expected patient Gender to be non-nil")
	} else if *patientResponse.Gender != expectedGender {
		t.Errorf("Expected patient Gender %s, got %s", expectedGender, *patientResponse.Gender)
	}

	// Verify the patient birth date
	expectedBirthDate := "1990-01-15"
	if patientResponse.BirthDate == nil {
		t.Error("Expected patient BirthDate to be non-nil")
	} else if *patientResponse.BirthDate != expectedBirthDate {
		t.Errorf("Expected patient BirthDate %s, got %s", expectedBirthDate, *patientResponse.BirthDate)
	}

	// Verify the patient has at least one name
	if len(patientResponse.Name) == 0 {
		t.Error("Expected patient to have at least one name")
	} else {
		// Verify family name
		expectedFamilyName := "Smith"
		if patientResponse.Name[0].Family == nil {
			t.Error("Expected patient family name to be non-nil")
		} else if *patientResponse.Name[0].Family != expectedFamilyName {
			t.Errorf("Expected family name %s, got %s", expectedFamilyName, *patientResponse.Name[0].Family)
		}

		// Verify given name
		if len(patientResponse.Name[0].Given) == 0 {
			t.Error("Expected patient to have at least one given name")
		} else {
			expectedGivenName := "John"
			if patientResponse.Name[0].Given[0] != expectedGivenName {
				t.Errorf("Expected given name %s, got %s", expectedGivenName, patientResponse.Name[0].Given[0])
			}
		}
	}

	// Verify the patient has at least one identifier
	if len(patientResponse.Identifier) == 0 {
		t.Error("Expected patient to have at least one identifier")
	}
}

// TestNewPatientHandler verifies the constructor creates a valid instance
func TestNewPatientHandler(t *testing.T) {
	patientHandler := NewPatientHandler()

	if patientHandler == nil {
		t.Error("Expected NewPatientHandler to return non-nil instance")
	}
}
