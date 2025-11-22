package repository

import (
	"context"
	"testing"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
)

// setupObservationSearchTestData creates test observations for search testing
func setupObservationSearchTestData(t *testing.T, repository *MongoObservationRepository) {
	testObservations := []*models.Observation{
		{
			PatientID:     "patient-001",
			Status:        "final",
			Category:      "vital-signs",
			Code:          "8480-6", // Systolic BP
			CodeSystem:    "http://loinc.org",
			CodeDisplay:   "Systolic Blood Pressure",
			ValueQuantity: floatPtr(120.0),
			ValueUnit:     "mmHg",
			EffectiveDate: timePtr(parseDate("2024-01-15")),
		},
		{
			PatientID:     "patient-001",
			Status:        "final",
			Category:      "vital-signs",
			Code:          "8462-4", // Diastolic BP
			CodeSystem:    "http://loinc.org",
			CodeDisplay:   "Diastolic Blood Pressure",
			ValueQuantity: floatPtr(80.0),
			ValueUnit:     "mmHg",
			EffectiveDate: timePtr(parseDate("2024-01-15")),
		},
		{
			PatientID:     "patient-001",
			Status:        "preliminary",
			Category:      "laboratory",
			Code:          "2093-3", // Cholesterol
			CodeSystem:    "http://loinc.org",
			CodeDisplay:   "Cholesterol",
			ValueQuantity: floatPtr(180.0),
			ValueUnit:     "mg/dL",
			EffectiveDate: timePtr(parseDate("2024-02-01")),
		},
		{
			PatientID:     "patient-002",
			Status:        "final",
			Category:      "vital-signs",
			Code:          "8867-4", // Heart rate
			CodeSystem:    "http://loinc.org",
			CodeDisplay:   "Heart Rate",
			ValueQuantity: floatPtr(72.0),
			ValueUnit:     "beats/min",
			EffectiveDate: timePtr(parseDate("2024-01-20")),
		},
		{
			PatientID:     "patient-002",
			Status:        "final",
			Category:      "vital-signs",
			Code:          "8310-5", // Body temp
			CodeSystem:    "http://loinc.org",
			CodeDisplay:   "Body Temperature",
			ValueQuantity: floatPtr(98.6),
			ValueUnit:     "degF",
			EffectiveDate: timePtr(parseDate("2024-03-10")),
		},
	}

	for _, observation := range testObservations {
		_, createError := repository.Create(context.Background(), observation)
		if createError != nil {
			t.Fatalf("Failed to create test observation: %v", createError)
		}
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func parseDate(dateStr string) time.Time {
	date, _ := time.Parse("2006-01-02", dateStr)
	return date
}

// TestObservationRepository_Search_ByPatient tests searching by patient ID
func TestObservationRepository_Search_ByPatient(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		PatientID: "patient-001",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 observations for patient-001, got %d", len(results))
	}

	for _, observation := range results {
		if observation.PatientID != "patient-001" {
			t.Errorf("Expected patient ID patient-001, got %s", observation.PatientID)
		}
	}
}

// TestObservationRepository_Search_ByCode tests searching by LOINC code
func TestObservationRepository_Search_ByCode(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		Code:   "8480-6", // Systolic BP
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 observation with code 8480-6, got %d", len(results))
	}

	if len(results) > 0 && results[0].Code != "8480-6" {
		t.Errorf("Expected code 8480-6, got %s", results[0].Code)
	}
}

// TestObservationRepository_Search_ByCategory tests filtering by category
func TestObservationRepository_Search_ByCategory(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		Category: "vital-signs",
		Limit:    10,
		Offset:   0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 vital-signs observations, got %d", len(results))
	}

	for _, observation := range results {
		if observation.Category != "vital-signs" {
			t.Errorf("Expected category vital-signs, got %s", observation.Category)
		}
	}
}

// TestObservationRepository_Search_ByStatus tests filtering by status
func TestObservationRepository_Search_ByStatus(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		Status: "final",
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 4 {
		t.Errorf("Expected 4 final observations, got %d", len(results))
	}

	for _, observation := range results {
		if observation.Status != "final" {
			t.Errorf("Expected status final, got %s", observation.Status)
		}
	}
}

// TestObservationRepository_Search_ByDateGreaterThan tests date >= filter
func TestObservationRepository_Search_ByDateGreaterThan(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	cutoffDate := parseDate("2024-02-01")
	searchParams := &models.ObservationSearchParams{
		DateGreaterThan: &cutoffDate,
		Limit:           10,
		Offset:          0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get observations from Feb 1 onwards (2 observations)
	if len(results) != 2 {
		t.Errorf("Expected 2 observations after Feb 1, got %d", len(results))
	}

	for _, observation := range results {
		if observation.EffectiveDate != nil && observation.EffectiveDate.Before(cutoffDate) {
			t.Errorf("Expected date >= 2024-02-01, got %v", observation.EffectiveDate)
		}
	}
}

// TestObservationRepository_Search_ByDateLessThan tests date <= filter
func TestObservationRepository_Search_ByDateLessThan(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	cutoffDate := parseDate("2024-01-31")
	searchParams := &models.ObservationSearchParams{
		DateLessThan: &cutoffDate,
		Limit:        10,
		Offset:       0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get observations up to Jan 31 (3 observations)
	if len(results) != 3 {
		t.Errorf("Expected 3 observations before Jan 31, got %d", len(results))
	}

	for _, observation := range results {
		if observation.EffectiveDate != nil && observation.EffectiveDate.After(cutoffDate) {
			t.Errorf("Expected date <= 2024-01-31, got %v", observation.EffectiveDate)
		}
	}
}

// TestObservationRepository_Search_CombinedFilters tests multiple filters
func TestObservationRepository_Search_CombinedFilters(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		PatientID: "patient-001",
		Category:  "vital-signs",
		Status:    "final",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get final vital signs for patient-001 (2 observations: systolic and diastolic BP)
	if len(results) != 2 {
		t.Errorf("Expected 2 observations, got %d", len(results))
	}

	for _, observation := range results {
		if observation.PatientID != "patient-001" {
			t.Errorf("Expected patient-001, got %s", observation.PatientID)
		}
		if observation.Category != "vital-signs" {
			t.Errorf("Expected vital-signs, got %s", observation.Category)
		}
		if observation.Status != "final" {
			t.Errorf("Expected final, got %s", observation.Status)
		}
	}
}

// TestObservationRepository_Search_SortByEffectiveDate tests sorting by date
func TestObservationRepository_Search_SortByEffectiveDate(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		SortBy:    "effective_date",
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

	// Verify ascending order
	for i := 0; i < len(results)-1; i++ {
		if results[i].EffectiveDate != nil && results[i+1].EffectiveDate != nil {
			if results[i].EffectiveDate.After(*results[i+1].EffectiveDate) {
				t.Errorf("Expected ascending order, got %v after %v", results[i].EffectiveDate, results[i+1].EffectiveDate)
			}
		}
	}
}

// TestObservationRepository_Search_SortDescending tests descending sort
func TestObservationRepository_Search_SortDescending(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		SortBy:    "effective_date",
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
	for i := 0; i < len(results)-1; i++ {
		if results[i].EffectiveDate != nil && results[i+1].EffectiveDate != nil {
			if results[i].EffectiveDate.Before(*results[i+1].EffectiveDate) {
				t.Errorf("Expected descending order, got %v before %v", results[i].EffectiveDate, results[i+1].EffectiveDate)
			}
		}
	}
}

// TestObservationRepository_Search_Pagination tests pagination
func TestObservationRepository_Search_Pagination(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	// Get first page
	searchParams1 := &models.ObservationSearchParams{
		SortBy:    "effective_date",
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
	searchParams2 := &models.ObservationSearchParams{
		SortBy:    "effective_date",
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

// TestObservationRepository_Search_EmptyResults tests search with no matches
func TestObservationRepository_Search_EmptyResults(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		PatientID: "patient-nonexistent",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

// TestObservationRepository_Search_DefaultsWithNoFilters tests search with no filters
func TestObservationRepository_Search_DefaultsWithNoFilters(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		Limit:  10,
		Offset: 0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should return all observations
	if len(results) != 5 {
		t.Errorf("Expected all 5 observations, got %d", len(results))
	}
}

// TestObservationRepository_Search_PatientAndCode tests combining patient and code filters
func TestObservationRepository_Search_PatientAndCode(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	searchParams := &models.ObservationSearchParams{
		PatientID: "patient-001",
		Code:      "8480-6",
		Limit:     10,
		Offset:    0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 observation, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].PatientID != "patient-001" {
			t.Errorf("Expected patient-001, got %s", results[0].PatientID)
		}
		if results[0].Code != "8480-6" {
			t.Errorf("Expected code 8480-6, got %s", results[0].Code)
		}
	}
}

// TestObservationRepository_Search_DateRange tests date range filtering
func TestObservationRepository_Search_DateRange(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	startDate := parseDate("2024-01-10")
	endDate := parseDate("2024-02-15")

	searchParams := &models.ObservationSearchParams{
		DateGreaterThan: &startDate,
		DateLessThan:    &endDate,
		Limit:           10,
		Offset:          0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get observations between Jan 10 and Feb 15
	if len(results) != 4 {
		t.Errorf("Expected 4 observations in date range, got %d", len(results))
	}

	for _, observation := range results {
		if observation.EffectiveDate != nil {
			if observation.EffectiveDate.Before(startDate) || observation.EffectiveDate.After(endDate) {
				t.Errorf("Expected date between %v and %v, got %v", startDate, endDate, observation.EffectiveDate)
			}
		}
	}
}

// TestObservationRepository_Search_AllFilters tests search with all possible filters
func TestObservationRepository_Search_AllFilters(t *testing.T) {
	mongoDatabase := setupTestMongoDB(t)
	repository := NewMongoObservationRepository(mongoDatabase)
	defer cleanupMongoTestData(t, repository.collection)

	setupObservationSearchTestData(t, repository)

	startDate := parseDate("2024-01-01")
	endDate := parseDate("2024-01-31")

	searchParams := &models.ObservationSearchParams{
		PatientID:       "patient-001",
		Code:            "8480-6",
		Category:        "vital-signs",
		Status:          "final",
		DateGreaterThan: &startDate,
		DateLessThan:    &endDate,
		SortBy:          "effective_date",
		SortOrder:       "desc",
		Limit:           10,
		Offset:          0,
	}

	results, searchError := repository.Search(context.Background(), searchParams)

	if searchError != nil {
		t.Fatalf("Expected no error, got %v", searchError)
	}

	// Should get 1 observation matching all criteria
	if len(results) != 1 {
		t.Errorf("Expected 1 observation with all filters, got %d", len(results))
	}

	if len(results) > 0 {
		obs := results[0]
		if obs.PatientID != "patient-001" || obs.Code != "8480-6" ||
			obs.Category != "vital-signs" || obs.Status != "final" {
			t.Error("Result doesn't match all filter criteria")
		}
	}
}

