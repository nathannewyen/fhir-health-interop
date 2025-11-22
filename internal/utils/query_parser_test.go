package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestParsePatientSearchParams_Name tests parsing the name parameter
func TestParsePatientSearchParams_Name(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?name=Smith", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Name != "Smith" {
		t.Errorf("Expected name 'Smith', got '%s'", searchParams.Name)
	}
}

// TestParsePatientSearchParams_FamilyAndGiven tests parsing family and given name parameters
func TestParsePatientSearchParams_FamilyAndGiven(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?family=Doe&given=John", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.FamilyName != "Doe" {
		t.Errorf("Expected family name 'Doe', got '%s'", searchParams.FamilyName)
	}

	if searchParams.GivenName != "John" {
		t.Errorf("Expected given name 'John', got '%s'", searchParams.GivenName)
	}
}

// TestParsePatientSearchParams_Gender tests parsing gender parameter
func TestParsePatientSearchParams_Gender(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?gender=male", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Gender != "male" {
		t.Errorf("Expected gender 'male', got '%s'", searchParams.Gender)
	}
}

// TestParsePatientSearchParams_Active tests parsing active boolean parameter
func TestParsePatientSearchParams_Active(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?active=true", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Active == nil {
		t.Fatal("Expected active to be set")
	}

	if *searchParams.Active != true {
		t.Errorf("Expected active true, got %v", *searchParams.Active)
	}
}

// TestParsePatientSearchParams_BirthdateExact tests parsing exact birthdate
func TestParsePatientSearchParams_BirthdateExact(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?birthdate=1990-05-15", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.BirthDate == nil {
		t.Fatal("Expected birthdate to be set")
	}

	expectedDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	if !searchParams.BirthDate.Equal(expectedDate) {
		t.Errorf("Expected birthdate %v, got %v", expectedDate, *searchParams.BirthDate)
	}
}

// TestParsePatientSearchParams_BirthdateGreaterThan tests parsing birthdate with ge prefix
func TestParsePatientSearchParams_BirthdateGreaterThan(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?birthdate=ge1990-01-01", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.BirthDateGreaterThan == nil {
		t.Fatal("Expected birthdate greater than to be set")
	}

	expectedDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	if !searchParams.BirthDateGreaterThan.Equal(expectedDate) {
		t.Errorf("Expected birthdate >= %v, got %v", expectedDate, *searchParams.BirthDateGreaterThan)
	}
}

// TestParsePatientSearchParams_BirthdateLessThan tests parsing birthdate with le prefix
func TestParsePatientSearchParams_BirthdateLessThan(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?birthdate=le2000-12-31", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.BirthDateLessThan == nil {
		t.Fatal("Expected birthdate less than to be set")
	}

	expectedDate := time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)
	if !searchParams.BirthDateLessThan.Equal(expectedDate) {
		t.Errorf("Expected birthdate <= %v, got %v", expectedDate, *searchParams.BirthDateLessThan)
	}
}

// TestParsePatientSearchParams_Sorting tests parsing sort parameter
func TestParsePatientSearchParams_Sorting(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_sort=-name", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.SortBy != "name" {
		t.Errorf("Expected sort by 'name', got '%s'", searchParams.SortBy)
	}

	if searchParams.SortOrder != "desc" {
		t.Errorf("Expected sort order 'desc', got '%s'", searchParams.SortOrder)
	}
}

// TestParsePatientSearchParams_Pagination tests parsing limit and offset parameters
func TestParsePatientSearchParams_Pagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=20&_offset=40", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", searchParams.Limit)
	}

	if searchParams.Offset != 40 {
		t.Errorf("Expected offset 40, got %d", searchParams.Offset)
	}
}

// TestParsePatientSearchParams_MaxLimit tests that limit is capped at 100
func TestParsePatientSearchParams_MaxLimit(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=500", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Limit != 100 {
		t.Errorf("Expected limit capped at 100, got %d", searchParams.Limit)
	}
}

// TestParsePatientSearchParams_DefaultValues tests default values when no parameters provided
func TestParsePatientSearchParams_DefaultValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)

	searchParams, parseError := ParsePatientSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Limit != 10 {
		t.Errorf("Expected default limit 10, got %d", searchParams.Limit)
	}

	if searchParams.Offset != 0 {
		t.Errorf("Expected default offset 0, got %d", searchParams.Offset)
	}
}

// TestParseObservationSearchParams_Patient tests parsing patient parameter
func TestParseObservationSearchParams_Patient(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?patient=123", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.PatientID != "123" {
		t.Errorf("Expected patient ID '123', got '%s'", searchParams.PatientID)
	}
}

// TestParseObservationSearchParams_PatientWithPrefix tests parsing patient parameter with Patient/ prefix
func TestParseObservationSearchParams_PatientWithPrefix(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?patient=Patient/456", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.PatientID != "456" {
		t.Errorf("Expected patient ID '456', got '%s'", searchParams.PatientID)
	}
}

// TestParseObservationSearchParams_Code tests parsing code parameter
func TestParseObservationSearchParams_Code(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?code=8480-6", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Code != "8480-6" {
		t.Errorf("Expected code '8480-6', got '%s'", searchParams.Code)
	}
}

// TestParseObservationSearchParams_Status tests parsing status parameter
func TestParseObservationSearchParams_Status(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?status=final", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Status != "final" {
		t.Errorf("Expected status 'final', got '%s'", searchParams.Status)
	}
}

// TestParseObservationSearchParams_Category tests parsing category parameter
func TestParseObservationSearchParams_Category(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?category=vital-signs", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.Category != "vital-signs" {
		t.Errorf("Expected category 'vital-signs', got '%s'", searchParams.Category)
	}
}

// TestParseObservationSearchParams_DateRange tests parsing date range parameters
func TestParseObservationSearchParams_DateRange(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?date=ge2024-01-01", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.DateGreaterThan == nil {
		t.Fatal("Expected date greater than to be set")
	}

	expectedDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !searchParams.DateGreaterThan.Equal(expectedDate) {
		t.Errorf("Expected date >= %v, got %v", expectedDate, *searchParams.DateGreaterThan)
	}
}

// TestParseObservationSearchParams_MultipleFilters tests parsing multiple filters
func TestParseObservationSearchParams_MultipleFilters(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fhir/Observation?patient=123&code=8480-6&status=final&_count=25", nil)

	searchParams, parseError := ParseObservationSearchParams(request)

	if parseError != nil {
		t.Fatalf("Expected no error, got %v", parseError)
	}

	if searchParams.PatientID != "123" {
		t.Errorf("Expected patient ID '123', got '%s'", searchParams.PatientID)
	}

	if searchParams.Code != "8480-6" {
		t.Errorf("Expected code '8480-6', got '%s'", searchParams.Code)
	}

	if searchParams.Status != "final" {
		t.Errorf("Expected status 'final', got '%s'", searchParams.Status)
	}

	if searchParams.Limit != 25 {
		t.Errorf("Expected limit 25, got %d", searchParams.Limit)
	}
}
