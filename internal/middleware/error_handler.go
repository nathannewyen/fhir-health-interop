package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	apperrors "github.com/nathannewyen/fhir-health-interop/internal/errors"
	"github.com/rs/zerolog/log"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
)

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information returned to clients
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// RequestID middleware generates a unique request ID for each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Continue with request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ErrorHandler middleware catches panics and handles errors
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestID(r.Context())

				log.Error().
					Interface("panic", err).
					Str("request_id", requestID).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("Panic recovered")

				// Send 500 error response
				sendErrorResponse(w, r, apperrors.Internal("Internal server error", nil))
			}
		}()

		// Create custom response writer to capture status
		errorWriter := &errorResponseWriter{
			ResponseWriter: w,
			request:        r,
		}

		// Continue with request
		next.ServeHTTP(errorWriter, r)
	})
}

// errorResponseWriter wraps http.ResponseWriter to intercept error responses
type errorResponseWriter struct {
	http.ResponseWriter
	request *http.Request
}

// getRequestID retrieves request ID from context
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, r *http.Request, err *apperrors.AppError) {
	requestID := getRequestID(r.Context())

	// Log error with context
	logEvent := log.Error()
	if err.StatusCode < 500 {
		logEvent = log.Warn()
	}

	logEvent.
		Str("request_id", requestID).
		Str("path", r.URL.Path).
		Str("method", r.Method).
		Str("error_code", err.Code).
		Int("status", err.StatusCode).
		Err(err.Err).
		Msg(err.Message)

	// Prepare error response
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:      err.Code,
			Message:   err.Message,
			RequestID: requestID,
		},
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// WriteError is a helper function for handlers to write error responses
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	// Convert to AppError if not already
	var appErr *apperrors.AppError
	if e, ok := err.(*apperrors.AppError); ok {
		appErr = e
	} else {
		// Unknown error, treat as internal server error
		appErr = apperrors.Internal("Internal server error", err)
	}

	sendErrorResponse(w, r, appErr)
}
