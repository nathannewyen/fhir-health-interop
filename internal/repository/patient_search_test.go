package repository

import (
	"context"
	"testing"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
)

// setupPatientSearchTestData creates test patients for search testing
func setupPatientSearchTestData(t *testing.T, repository *PostgresPatientRepository) {
	testPatients := []*models.Patient{
		{
			IdentifierSystem: "http://hospital.com",
			IdentifierValue:  "P001",
			Active:           true,
			FamilyName:       "Smith",
			GivenName:        "John",
			Gender:           "male",
			BirthDate:        parseTestDate("1985-03-15"),
		},
		{
			IdentifierSystem: "http://hospital.com",
			IdentifierValue:  "P002",
			Active:           true,
			FamilyName:       "Smith",
			GivenName:        "Jane",
			Gender:           "female",
			BirthDate:        parseTestDate("1990-07-22"),
		},
		{
			IdentifierSystem: "http://hospital.com",
			IdentifierValue:  "P003",
			Active:           false,
			FamilyName:       "Johnson",
			GivenName:        "Robert",
			Gender:           "male",
			BirthDate:        parseTestDate("1978-11-30"),
		},
		{
			IdentifierSystem: "http://hospital.com",
			IdentifierValue:  "P004",
			Active:           true,
			FamilyName:       "Williams",
			GivenName:        "Emily",
			Gender:           "female",
			BirthDate:        parseTestDate("1995-05-18"),
		},
		{
			IdentifierSystem: "http://hospital.com",
			IdentifierValue:  "P005",
			Active:           true,
			FamilyName:       "Brown",
			GivenName:        "Michael",
			Gender:           "male",
			BirthDate:        parseTestDate("2000-01-10"),
		},
	}

	for _, patient := range testPatients {
		_, createError := repository.Create(context.Background(), patient)
		if createError != nil {
			t.Fatalf("Failed to create test patient: %v", createError)
		}
	}
}

// parseTestDate helper for test data - returns pointer
func parseTestDate(dateStr string) *time.Time {
	date, _ := time.Parse("2006-01-02", dateStr)
	return &date
}

// parseDate helper for date filtering (returns value)
func parseTestDateValue(dateStr string) time.Time {
	date, _ := time.Parse("2006-01-02", dateStr)
	return date
}

// TestPatientRepository_Search_ByName tests searching by name
func TestPatientRepository_Search_ByName(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		Name:   "Smith",
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 patients named Smith, got %d", len(results))
	}

	// Verify results contain Smith patients
	for _, patient := range results {
		if patient.FamilyName != "Smith" {
			t.Errorf("Expected family name Smith, got %s", patient.FamilyName)
		}
	}
}

// TestPatientRepository_Search_ByFamilyName tests searching by family name only
func TestPatientRepository_Search_ByFamilyName(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		FamilyName: "Johnson",
		Limit:      10,
		Offset:     0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 patient named Johnson, got %d", len(results))
	}

	if len(results) > 0 && results[0].FamilyName != "Johnson" {
		t.Errorf("Expected family name Johnson, got %s", results[0].FamilyName)
	}
}

// TestPatientRepository_Search_ByGivenName tests searching by given name only
func TestPatientRepository_Search_ByGivenName(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		GivenName: "John",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 patient with given name John, got %d", len(results))
	}

	if len(results) > 0 && results[0].GivenName != "John" {
		t.Errorf("Expected given name John, got %s", results[0].GivenName)
	}
}

// TestPatientRepository_Search_ByGender tests filtering by gender
func TestPatientRepository_Search_ByGender(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		Gender: "female",
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 female patients, got %d", len(results))
	}

	for _, patient := range results {
		if patient.Gender != "female" {
			t.Errorf("Expected gender female, got %s", patient.Gender)
		}
	}
}

// TestPatientRepository_Search_ByActiveStatus tests filtering by active status
func TestPatientRepository_Search_ByActiveStatus(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	activeTrue := true
	searchParams := &models.PatientSearchParams{
		Active: &activeTrue,
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 active patients, got %d", len(results))
	}

	for _, patient := range results {
		if !patient.Active {
			t.Error("Expected all patients to be active")
		}
	}
}

// TestPatientRepository_Search_ByInactiveStatus tests filtering by inactive status
func TestPatientRepository_Search_ByInactiveStatus(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	activeFalse := false
	searchParams := &models.PatientSearchParams{
		Active: &activeFalse,
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 inactive patient, got %d", len(results))
	}

	if len(results) > 0 && results[0].Active {
		t.Error("Expected patient to be inactive")
	}
}

// TestPatientRepository_Search_ByBirthDateExact tests exact birth date match
func TestPatientRepository_Search_ByBirthDateExact(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	exactDate := parseTestDateValue("1990-07-22")
	searchParams := &models.PatientSearchParams{
		BirthDate: &exactDate,
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get exactly one patient (Jane Smith born 1990-07-22)
	if len(results) != 1 {
		t.Errorf("Expected 1 patient with exact birth date 1990-07-22, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].FamilyName != "Smith" || results[0].GivenName != "Jane" {
			t.Errorf("Expected Jane Smith, got %s %s", results[0].GivenName, results[0].FamilyName)
		}
	}
}

// TestPatientRepository_Search_ByBirthDateGreaterThan tests birth date >= filter
func TestPatientRepository_Search_ByBirthDateGreaterThan(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	cutoffDate := parseTestDateValue("1990-01-01")
	searchParams := &models.PatientSearchParams{
		BirthDateGreaterThan: &cutoffDate,
		Limit:                10,
		Offset:               0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get patients born in 1990, 1995, and 2000
	if len(results) != 3 {
		t.Errorf("Expected 3 patients born after 1990, got %d", len(results))
	}

	for _, patient := range results {
		if patient.BirthDate.Before(cutoffDate) {
			t.Errorf("Expected birth date >= 1990-01-01, got %v", patient.BirthDate)
		}
	}
}

// TestPatientRepository_Search_ByBirthDateLessThan tests birth date <= filter
func TestPatientRepository_Search_ByBirthDateLessThan(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	cutoffDate := parseTestDateValue("1990-01-01")
	searchParams := &models.PatientSearchParams{
		BirthDateLessThan: &cutoffDate,
		Limit:             10,
		Offset:            0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get patients born in 1978 and 1985
	if len(results) != 2 {
		t.Errorf("Expected 2 patients born before 1990, got %d", len(results))
	}

	for _, patient := range results {
		if patient.BirthDate.After(cutoffDate) {
			t.Errorf("Expected birth date <= 1990-01-01, got %v", patient.BirthDate)
		}
	}
}

// TestPatientRepository_Search_CombinedFilters tests multiple filters together
func TestPatientRepository_Search_CombinedFilters(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	activeTrue := true
	cutoffDate := parseTestDateValue("1990-01-01")
	searchParams := &models.PatientSearchParams{
		Gender:               "female",
		Active:               &activeTrue,
		BirthDateGreaterThan: &cutoffDate,
		Limit:                10,
		Offset:               0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get active females born after 1990 (Jane Smith and Emily Williams)
	if len(results) != 2 {
		t.Errorf("Expected 2 patients, got %d", len(results))
	}

	for _, patient := range results {
		if patient.Gender != "female" {
			t.Errorf("Expected gender female, got %s", patient.Gender)
		}
		if !patient.Active {
			t.Error("Expected patient to be active")
		}
		if patient.BirthDate.Before(cutoffDate) {
			t.Errorf("Expected birth date >= 1990-01-01, got %v", patient.BirthDate)
		}
	}
}

// TestPatientRepository_Search_SortByName tests sorting by name
func TestPatientRepository_Search_SortByName(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		SortBy:    "name",
		SortOrder: "asc",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) < 2 {
		t.Fatal("Expected at least 2 results for sort test")
	}

	// Verify ascending order by family name
	if results[0].FamilyName > results[1].FamilyName {
		t.Errorf("Expected ascending order, got %s before %s", results[0].FamilyName, results[1].FamilyName)
	}
}

// TestPatientRepository_Search_SortByBirthDateDescending tests sorting by birth date descending
func TestPatientRepository_Search_SortByBirthDateDescending(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		SortBy:    "birthdate",
		SortOrder: "desc",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) < 2 {
		t.Fatal("Expected at least 2 results for sort test")
	}

	// Verify descending order (newest first)
	if results[0].BirthDate != nil && results[1].BirthDate != nil {
		if results[0].BirthDate.Before(*results[1].BirthDate) {
			t.Errorf("Expected descending order, got %v before %v", results[0].BirthDate, results[1].BirthDate)
		}
	}
}

// TestPatientRepository_Search_Pagination tests pagination with offset
func TestPatientRepository_Search_Pagination(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	// Get first page
	searchParams1 := &models.PatientSearchParams{
		SortBy:    "created_at",
		SortOrder: "asc",
		Limit:     2,
		Offset:    0,
	}

	page1, error1 := repository.Search(context.Background(), searchParams1)
	if error1 != nil {
		t.Fatalf("Expected no error on page 1, got %v", error1)
	}

	if len(page1) != 2 {
		t.Errorf("Expected 2 results on page 1, got %d", len(page1))
	}

	// Get second page
	searchParams2 := &models.PatientSearchParams{
		SortBy:    "created_at",
		SortOrder: "asc",
		Limit:     2,
		Offset:    2,
	}

	page2, error2 := repository.Search(context.Background(), searchParams2)
	if error2 != nil {
		t.Fatalf("Expected no error on page 2, got %v", error2)
	}

	if len(page2) != 2 {
		t.Errorf("Expected 2 results on page 2, got %d", len(page2))
	}

	// Verify pages are different
	if len(page1) > 0 && len(page2) > 0 {
		if page1[0].ID == page2[0].ID {
			t.Error("Expected different results on different pages")
		}
	}
}

// TestPatientRepository_Search_EmptyResults tests search with no matches
func TestPatientRepository_Search_EmptyResults(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		Name:   "NonexistentName",
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

// TestPatientRepository_Search_CaseInsensitive tests case-insensitive search
func TestPatientRepository_Search_CaseInsensitive(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		Name:   "smith", // lowercase
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 patients (case-insensitive), got %d", len(results))
	}
}

// TestPatientRepository_Search_PartialMatch tests partial name matching
func TestPatientRepository_Search_PartialMatch(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		Name:   "Smi", // partial match
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 patients (partial match), got %d", len(results))
	}
}

// TestPatientRepository_Search_DefaultsWithNoFilters tests search with no filters
func TestPatientRepository_Search_DefaultsWithNoFilters(t *testing.T) {
	testDB := setupTestDatabase(t)
	repository := NewPostgresPatientRepository(testDB)
	defer cleanupTestData(t, testDB)

	setupPatientSearchTestData(t, repository)

	searchParams := &models.PatientSearchParams{
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should return all patients
	if len(results) != 5 {
		t.Errorf("Expected all 5 patients, got %d", len(results))
	}
}
