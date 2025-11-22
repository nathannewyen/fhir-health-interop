package utils

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nathannewyen/fhir-health-interop/internal/models"
)

// ParsePatientSearchParams extracts and validates patient search parameters from HTTP request
func ParsePatientSearchParams(request *http.Request) (*models.PatientSearchParams, error) {
	queryParams := request.URL.Query()

	searchParams := &models.PatientSearchParams{
		Limit:  10,  // Default limit
		Offset: 0,   // Default offset
	}

	// Parse name parameter
	if name := queryParams.Get("name"); name != "" {
		searchParams.Name = name
	}

	// Parse family name parameter
	if familyName := queryParams.Get("family"); familyName != "" {
		searchParams.FamilyName = familyName
	}

	// Parse given name parameter
	if givenName := queryParams.Get("given"); givenName != "" {
		searchParams.GivenName = givenName
	}

	// Parse gender parameter
	if gender := queryParams.Get("gender"); gender != "" {
		searchParams.Gender = gender
	}

	// Parse birthdate parameter with prefixes
	if birthdate := queryParams.Get("birthdate"); birthdate != "" {
		parsedDate, prefix := parseDateWithPrefix(birthdate)
		if parsedDate != nil {
			switch prefix {
			case "ge", "gt":
				searchParams.BirthDateGreaterThan = parsedDate
			case "le", "lt":
				searchParams.BirthDateLessThan = parsedDate
			case "eq", "":
				searchParams.BirthDate = parsedDate
			}
		}
	}

	// Parse active parameter
	if active := queryParams.Get("active"); active != "" {
		if activeBool, parseError := strconv.ParseBool(active); parseError == nil {
			searchParams.Active = &activeBool
		}
	}

	// Parse sort parameter
	if sortBy := queryParams.Get("_sort"); sortBy != "" {
		// Handle descending sort (prefix with -)
		if strings.HasPrefix(sortBy, "-") {
			searchParams.SortBy = strings.TrimPrefix(sortBy, "-")
			searchParams.SortOrder = "desc"
		} else {
			searchParams.SortBy = sortBy
			searchParams.SortOrder = "asc"
		}
	}

	// Parse limit parameter
	if limit := queryParams.Get("_count"); limit != "" {
		if limitInt, parseError := strconv.Atoi(limit); parseError == nil && limitInt > 0 {
			if limitInt > 100 {
				limitInt = 100 // Maximum limit
			}
			searchParams.Limit = limitInt
		}
	}

	// Parse offset parameter
	if offset := queryParams.Get("_offset"); offset != "" {
		if offsetInt, parseError := strconv.Atoi(offset); parseError == nil && offsetInt >= 0 {
			searchParams.Offset = offsetInt
		}
	}

	return searchParams, nil
}

// ParseObservationSearchParams extracts and validates observation search parameters from HTTP request
func ParseObservationSearchParams(request *http.Request) (*models.ObservationSearchParams, error) {
	queryParams := request.URL.Query()

	searchParams := &models.ObservationSearchParams{
		Limit:  10,  // Default limit
		Offset: 0,   // Default offset
	}

	// Parse patient parameter
	if patientID := queryParams.Get("patient"); patientID != "" {
		// Handle both "patient=123" and "patient=Patient/123" formats
		if strings.HasPrefix(patientID, "Patient/") {
			searchParams.PatientID = strings.TrimPrefix(patientID, "Patient/")
		} else {
			searchParams.PatientID = patientID
		}
	}

	// Parse code parameter
	if code := queryParams.Get("code"); code != "" {
		searchParams.Code = code
	}

	// Parse category parameter
	if category := queryParams.Get("category"); category != "" {
		searchParams.Category = category
	}

	// Parse status parameter
	if status := queryParams.Get("status"); status != "" {
		searchParams.Status = status
	}

	// Parse date parameter with prefixes
	if date := queryParams.Get("date"); date != "" {
		parsedDate, prefix := parseDateWithPrefix(date)
		if parsedDate != nil {
			switch prefix {
			case "ge", "gt":
				searchParams.DateGreaterThan = parsedDate
			case "le", "lt":
				searchParams.DateLessThan = parsedDate
			}
		}
	}

	// Parse sort parameter
	if sortBy := queryParams.Get("_sort"); sortBy != "" {
		// Handle descending sort (prefix with -)
		if strings.HasPrefix(sortBy, "-") {
			searchParams.SortBy = strings.TrimPrefix(sortBy, "-")
			searchParams.SortOrder = "desc"
		} else {
			searchParams.SortBy = sortBy
			searchParams.SortOrder = "asc"
		}
	}

	// Parse limit parameter
	if limit := queryParams.Get("_count"); limit != "" {
		if limitInt, parseError := strconv.Atoi(limit); parseError == nil && limitInt > 0 {
			if limitInt > 100 {
				limitInt = 100 // Maximum limit
			}
			searchParams.Limit = limitInt
		}
	}

	// Parse offset parameter
	if offset := queryParams.Get("_offset"); offset != "" {
		if offsetInt, parseError := strconv.Atoi(offset); parseError == nil && offsetInt >= 0 {
			searchParams.Offset = offsetInt
		}
	}

	return searchParams, nil
}

// parseDateWithPrefix extracts date prefix (ge, le, etc.) and parses the date
func parseDateWithPrefix(dateString string) (*time.Time, string) {
	prefix := ""

	// Check for FHIR comparison prefixes
	if len(dateString) > 2 {
		potentialPrefix := dateString[:2]
		if potentialPrefix == "ge" || potentialPrefix == "le" || potentialPrefix == "gt" || potentialPrefix == "lt" || potentialPrefix == "eq" {
			prefix = potentialPrefix
			dateString = dateString[2:]
		}
	}

	// Try multiple date formats
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD
		"2006-01-02T15:04:05Z", // ISO 8601
		"2006-01-02T15:04:05",  // ISO 8601 without timezone
	}

	for _, format := range formats {
		if parsedTime, parseError := time.Parse(format, dateString); parseError == nil {
			return &parsedTime, prefix
		}
	}

	return nil, prefix
}
