package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// TestFHIRValidator_ValidPatient verifies valid patient passes validation
func TestFHIRValidator_ValidPatient(t *testing.T) {
	validPatientJSON := `{
		"resourceType": "Patient",
		"name": [{"family": "Smith", "given": ["John"]}]
	}`

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	validatorMiddleware := FHIRValidator(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString(validPatientJSON))
	request.Header.Set("Content-Type", "application/fhir+json")
	recorder := httptest.NewRecorder()

	validatorMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", recorder.Code, recorder.Body.String())
	}
}

// TestFHIRValidator_InvalidPatient_NoName verifies validation fails for patient without name
func TestFHIRValidator_InvalidPatient_NoName(t *testing.T) {
	invalidPatientJSON := `{
		"resourceType": "Patient",
		"gender": "male"
	}`

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid patient")
	})

	validatorMiddleware := FHIRValidator(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString(invalidPatientJSON))
	recorder := httptest.NewRecorder()

	validatorMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "must have at least one name") {
		t.Errorf("Expected error message about missing name, got: %s", recorder.Body.String())
	}
}

// TestFHIRValidator_InvalidPatient_EmptyName verifies validation fails for empty name
func TestFHIRValidator_InvalidPatient_EmptyName(t *testing.T) {
	invalidPatientJSON := `{
		"resourceType": "Patient",
		"name": [{"family": "", "given": []}]
	}`

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid patient")
	})

	validatorMiddleware := FHIRValidator(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString(invalidPatientJSON))
	recorder := httptest.NewRecorder()

	validatorMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestFHIRValidator_InvalidJSON verifies validation fails for malformed JSON
func TestFHIRValidator_InvalidJSON(t *testing.T) {
	invalidJSON := `{invalid json`

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid JSON")
	})

	validatorMiddleware := FHIRValidator(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/fhir/Patient", bytes.NewBufferString(invalidJSON))
	recorder := httptest.NewRecorder()

	validatorMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
}

// TestFHIRValidator_GETRequest verifies GET requests skip validation
func TestFHIRValidator_GETRequest(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	validatorMiddleware := FHIRValidator(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient/123", nil)
	recorder := httptest.NewRecorder()

	validatorMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected GET request to pass through, got status %d", recorder.Code)
	}
}

// TestFHIRValidator_NonFHIREndpoint verifies non-FHIR endpoints skip validation
func TestFHIRValidator_NonFHIREndpoint(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	validatorMiddleware := FHIRValidator(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/health", bytes.NewBufferString("any body"))
	recorder := httptest.NewRecorder()

	validatorMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected non-FHIR endpoint to skip validation, got status %d", recorder.Code)
	}
}

// TestIsFHIREndpoint verifies FHIR endpoint detection
func TestIsFHIREndpoint(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"/fhir/Patient", true},
		{"/fhir/Observation", true},
		{"/fhir/Patient/123", true},
		{"/health", false},
		{"/api/data", false},
		{"/", false},
		{"/fhi", false},
	}

	for _, testCase := range testCases {
		result := isFHIREndpoint(testCase.path)
		if result != testCase.expected {
			t.Errorf("isFHIREndpoint(%s): expected %v, got %v", testCase.path, testCase.expected, result)
		}
	}
}

// TestExtractResourceType verifies resource type extraction from path
func TestExtractResourceType(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/fhir/Patient", "Patient"},
		{"/fhir/Patient/123", "Patient"},
		{"/fhir/Observation", "Observation"},
		{"/fhir/Observation/abc-def", "Observation"},
		{"/fhir/Patient?name=Smith", "Patient"},
		{"/fhir/", ""},
		{"/fhir", ""},
	}

	for _, testCase := range testCases {
		result := extractResourceType(testCase.path)
		if result != testCase.expected {
			t.Errorf("extractResourceType(%s): expected %s, got %s", testCase.path, testCase.expected, result)
		}
	}
}

// TestValidatePatient_ValidCases verifies valid patient scenarios
func TestValidatePatient_ValidCases(t *testing.T) {
	familyName := "Smith"
	givenName := "John"

	testCases := []struct {
		name    string
		patient *fhir.Patient
	}{
		{
			name: "Patient with family and given name",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{
					{Family: &familyName, Given: []string{givenName}},
				},
			},
		},
		{
			name: "Patient with only family name",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{
					{Family: &familyName},
				},
			},
		},
		{
			name: "Patient with only given name",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{
					{Given: []string{givenName}},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			validationError := validatePatient(testCase.patient)
			if validationError != nil {
				t.Errorf("Expected valid patient, got error: %v", validationError)
			}
		})
	}
}

// TestValidatePatient_InvalidCases verifies invalid patient scenarios
func TestValidatePatient_InvalidCases(t *testing.T) {
	emptyString := ""

	testCases := []struct {
		name          string
		patient       *fhir.Patient
		expectedError string
	}{
		{
			name:          "Patient with no name array",
			patient:       &fhir.Patient{},
			expectedError: "must have at least one name",
		},
		{
			name: "Patient with empty name array",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{},
			},
			expectedError: "must have at least one name",
		},
		{
			name: "Patient with empty family and no given name",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{
					{Family: &emptyString, Given: []string{}},
				},
			},
			expectedError: "must have family or given name",
		},
		{
			name: "Patient with empty family and empty given name",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{
					{Family: &emptyString, Given: []string{""}},
				},
			},
			expectedError: "must have family or given name",
		},
		{
			name: "Patient with only empty given names",
			patient: &fhir.Patient{
				Name: []fhir.HumanName{
					{Given: []string{"", "", ""}},
				},
			},
			expectedError: "must have family or given name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			validationError := validatePatient(testCase.patient)
			if validationError == nil {
				t.Error("Expected validation error, got nil")
			} else if !strings.Contains(validationError.Error(), testCase.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", testCase.expectedError, validationError)
			}
		})
	}
}

// TestValidationError_Error verifies ValidationError implements error interface
func TestValidationError_Error(t *testing.T) {
	validationError := &ValidationError{Message: "test error"}
	if validationError.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got: %s", validationError.Error())
	}
}
