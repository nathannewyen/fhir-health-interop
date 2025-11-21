package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
}

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new instance of HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check returns the health status of the service
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	// Build health response with current timestamp and service status
	healthResponse := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "fhir-health-interop",
	}

	// Set response headers for JSON content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and write the JSON response
	json.NewEncoder(w).Encode(healthResponse)
}
