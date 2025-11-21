package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthHandler_Check verifies the health check endpoint returns correct response
func TestHealthHandler_Check(t *testing.T) {
	// Create a new health handler instance
	healthHandler := NewHealthHandler()

	// Create a mock HTTP request for GET /health
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Create a response recorder to capture the response
	responseRecorder := httptest.NewRecorder()

	// Call the health check handler
	healthHandler.Check(responseRecorder, request)

	// Verify HTTP status code is 200 OK
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	// Verify Content-Type header is application/json
	contentType := responseRecorder.Header().Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	// Parse the JSON response body
	var healthResponse HealthResponse
	decodeError := json.NewDecoder(responseRecorder.Body).Decode(&healthResponse)
	if decodeError != nil {
		t.Fatalf("Failed to decode response body: %v", decodeError)
	}

	// Verify the status field is "healthy"
	expectedStatus := "healthy"
	if healthResponse.Status != expectedStatus {
		t.Errorf("Expected status %s, got %s", expectedStatus, healthResponse.Status)
	}

	// Verify the service field is correct
	expectedService := "fhir-health-interop"
	if healthResponse.Service != expectedService {
		t.Errorf("Expected service %s, got %s", expectedService, healthResponse.Service)
	}

	// Verify timestamp is not empty
	if healthResponse.Timestamp == "" {
		t.Error("Expected timestamp to be non-empty")
	}
}

// TestNewHealthHandler verifies the constructor creates a valid instance
func TestNewHealthHandler(t *testing.T) {
	healthHandler := NewHealthHandler()

	if healthHandler == nil {
		t.Error("Expected NewHealthHandler to return non-nil instance")
	}
}
