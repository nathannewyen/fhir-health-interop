package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// FHIRValidator middleware validates FHIR resource structure
func FHIRValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate POST and PUT requests with bodies
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Only validate FHIR endpoints
		if !isFHIREndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Read the request body
		bodyBytes, readError := io.ReadAll(r.Body)
		if readError != nil {
			log.Warn().Err(readError).Msg("Failed to read request body")
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Restore the body for downstream handlers
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Determine resource type from path
		resourceType := extractResourceType(r.URL.Path)

		// Validate based on resource type
		validationError := validateFHIRResource(bodyBytes, resourceType)
		if validationError != nil {
			log.Warn().
				Err(validationError).
				Str("resource_type", resourceType).
				Str("path", r.URL.Path).
				Msg("FHIR validation failed")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Invalid FHIR resource",
				"message": validationError.Error(),
			})
			return
		}

		// Validation passed, continue to next handler
		next.ServeHTTP(w, r)
	})
}

// isFHIREndpoint checks if the path is a FHIR endpoint
func isFHIREndpoint(path string) bool {
	return len(path) >= 5 && path[:5] == "/fhir"
}

// extractResourceType extracts the FHIR resource type from the URL path
// Example: /fhir/Patient -> "Patient"
func extractResourceType(path string) string {
	// Remove /fhir/ prefix
	if len(path) < 6 {
		return ""
	}

	resourcePath := path[6:] // Skip "/fhir/"

	// Find the first slash or end of string
	for i, char := range resourcePath {
		if char == '/' || char == '?' {
			return resourcePath[:i]
		}
	}

	return resourcePath
}

// validateFHIRResource validates a FHIR resource based on its type
func validateFHIRResource(bodyBytes []byte, resourceType string) error {
	switch resourceType {
	case "Patient":
		var patient fhir.Patient
		if unmarshalError := json.Unmarshal(bodyBytes, &patient); unmarshalError != nil {
			return unmarshalError
		}
		return validatePatient(&patient)
	default:
		// For unknown resource types, just validate it's valid JSON
		var genericResource map[string]interface{}
		return json.Unmarshal(bodyBytes, &genericResource)
	}
}

// validatePatient validates a FHIR Patient resource
func validatePatient(patient *fhir.Patient) error {
	// Check that at least a name is provided
	if len(patient.Name) == 0 {
		return &ValidationError{Message: "Patient must have at least one name"}
	}

	// Check that name has either family or given name with non-empty values
	hasValidName := false
	for _, name := range patient.Name {
		// Check for valid family name
		if name.Family != nil && *name.Family != "" {
			hasValidName = true
			break
		}

		// Check for valid given name (at least one non-empty string)
		for _, givenName := range name.Given {
			if givenName != "" {
				hasValidName = true
				break
			}
		}

		if hasValidName {
			break
		}
	}

	if !hasValidName {
		return &ValidationError{Message: "Patient name must have family or given name"}
	}

	return nil
}

// ValidationError represents a FHIR validation error
type ValidationError struct {
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}
