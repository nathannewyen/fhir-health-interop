package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/nathannewyen/fhir-health-interop/internal/errors"
)

// TestRequestID_GeneratesUniqueID verifies request ID generation
func TestRequestID_GeneratesUniqueID(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())
		if requestID == "" {
			t.Error("Expected request ID to be set in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	// Verify X-Request-ID header is set
	requestIDHeader := recorder.Header().Get("X-Request-ID")
	if requestIDHeader == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

// TestRequestID_UniquePerRequest verifies different requests get different IDs
func TestRequestID_UniquePerRequest(t *testing.T) {
	var requestID1, requestID2 string

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())
		if requestID1 == "" {
			requestID1 = requestID
		} else {
			requestID2 = requestID
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(testHandler)

	// First request
	request1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder1 := httptest.NewRecorder()
	middleware.ServeHTTP(recorder1, request1)

	// Second request
	request2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder2 := httptest.NewRecorder()
	middleware.ServeHTTP(recorder2, request2)

	if requestID1 == requestID2 {
		t.Error("Expected different request IDs for different requests")
	}
}

// TestErrorHandler_RecoversPanic verifies panic recovery
func TestErrorHandler_RecoversPanic(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := RequestID(ErrorHandler(testHandler))

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	// Should not panic - middleware should recover
	middleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 after panic, got %d", recorder.Code)
	}

	var errorResponse ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)

	if errorResponse.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("Expected error code INTERNAL_ERROR, got %s", errorResponse.Error.Code)
	}
}

// TestWriteError_NotFoundError verifies 404 error response
func TestWriteError_NotFoundError(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, apperrors.NotFound("Patient", "123"))
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", recorder.Code)
	}

	var errorResponse ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)

	if errorResponse.Error.Code != "RESOURCE_NOT_FOUND" {
		t.Errorf("Expected code RESOURCE_NOT_FOUND, got %s", errorResponse.Error.Code)
	}

	if errorResponse.Error.RequestID == "" {
		t.Error("Expected request ID in error response")
	}

	expectedMsg := "Patient with ID '123' not found"
	if errorResponse.Error.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, errorResponse.Error.Message)
	}
}

// TestWriteError_ValidationError verifies 400 error response
func TestWriteError_ValidationError(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, apperrors.ValidationError("Name is required"))
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}

	var errorResponse ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)

	if errorResponse.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("Expected code VALIDATION_ERROR, got %s", errorResponse.Error.Code)
	}
}

// TestWriteError_InternalError verifies 500 error response
func TestWriteError_InternalError(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalErr := errors.New("database connection failed")
		WriteError(w, r, apperrors.Internal("Failed to process", originalErr))
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", recorder.Code)
	}

	var errorResponse ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)

	if errorResponse.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("Expected code INTERNAL_ERROR, got %s", errorResponse.Error.Code)
	}
}

// TestWriteError_GenericError verifies generic error converted to AppError
func TestWriteError_GenericError(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		genericErr := errors.New("some error")
		WriteError(w, r, genericErr)
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for generic error, got %d", recorder.Code)
	}

	var errorResponse ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)

	if errorResponse.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("Expected code INTERNAL_ERROR, got %s", errorResponse.Error.Code)
	}
}

// TestWriteError_ContentType verifies JSON content type
func TestWriteError_ContentType(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, apperrors.NotFound("Resource", "1"))
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// TestGetRequestID_NoContext verifies empty string when no request ID
func TestGetRequestID_NoContext(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	requestID := getRequestID(request.Context())

	if requestID != "" {
		t.Errorf("Expected empty string when no request ID, got %s", requestID)
	}
}

// TestErrorResponse_JSONStructure verifies error response JSON structure
func TestErrorResponse_JSONStructure(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, apperrors.InvalidInput("email", "invalid format"))
	})

	middleware := RequestID(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/test", nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	responseBody := recorder.Body.String()

	// Verify JSON structure
	if !strings.Contains(responseBody, `"error"`) {
		t.Error("Expected response to contain 'error' field")
	}
	if !strings.Contains(responseBody, `"code"`) {
		t.Error("Expected response to contain 'code' field")
	}
	if !strings.Contains(responseBody, `"message"`) {
		t.Error("Expected response to contain 'message' field")
	}
	if !strings.Contains(responseBody, `"request_id"`) {
		t.Error("Expected response to contain 'request_id' field")
	}
}
