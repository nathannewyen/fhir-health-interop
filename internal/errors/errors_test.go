package errors

import (
	"errors"
	"net/http"
	"testing"
)

// TestNotFound verifies NotFound error creation
func TestNotFound(t *testing.T) {
	err := NotFound("Patient", "123")

	if err.Code != "RESOURCE_NOT_FOUND" {
		t.Errorf("Expected code RESOURCE_NOT_FOUND, got %s", err.Code)
	}
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", err.StatusCode)
	}
	if err.Message != "Patient with ID '123' not found" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

// TestValidationError verifies ValidationError creation
func TestValidationError(t *testing.T) {
	err := ValidationError("Name is required")

	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("Expected code VALIDATION_ERROR, got %s", err.Code)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", err.StatusCode)
	}
	if err.Message != "Name is required" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

// TestInvalidInput verifies InvalidInput error creation
func TestInvalidInput(t *testing.T) {
	err := InvalidInput("email", "must be valid email address")

	if err.Code != "INVALID_INPUT" {
		t.Errorf("Expected code INVALID_INPUT, got %s", err.Code)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", err.StatusCode)
	}
	expectedMsg := "Invalid input for field 'email': must be valid email address"
	if err.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, err.Message)
	}
}

// TestInternal verifies Internal error creation
func TestInternal(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := Internal("Failed to process request", originalErr)

	if err.Code != "INTERNAL_ERROR" {
		t.Errorf("Expected code INTERNAL_ERROR, got %s", err.Code)
	}
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", err.StatusCode)
	}
	if err.Err != originalErr {
		t.Error("Expected wrapped error to be preserved")
	}
}

// TestConflict verifies Conflict error creation
func TestConflict(t *testing.T) {
	err := Conflict("Patient", "duplicate identifier")

	if err.Code != "CONFLICT" {
		t.Errorf("Expected code CONFLICT, got %s", err.Code)
	}
	if err.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", err.StatusCode)
	}
	expectedMsg := "Patient conflict: duplicate identifier"
	if err.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, err.Message)
	}
}

// TestUnauthorized verifies Unauthorized error creation
func TestUnauthorized(t *testing.T) {
	err := Unauthorized("Invalid credentials")

	if err.Code != "UNAUTHORIZED" {
		t.Errorf("Expected code UNAUTHORIZED, got %s", err.Code)
	}
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", err.StatusCode)
	}
}

// TestForbidden verifies Forbidden error creation
func TestForbidden(t *testing.T) {
	err := Forbidden("Access denied")

	if err.Code != "FORBIDDEN" {
		t.Errorf("Expected code FORBIDDEN, got %s", err.Code)
	}
	if err.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", err.StatusCode)
	}
}

// TestAppError_Error verifies Error() method
func TestAppError_Error(t *testing.T) {
	originalErr := errors.New("original error")
	err := Internal("Something went wrong", originalErr)

	expected := "Something went wrong: original error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

// TestAppError_ErrorWithoutWrapped verifies Error() without wrapped error
func TestAppError_ErrorWithoutWrapped(t *testing.T) {
	err := ValidationError("Test error")

	if err.Error() != "Test error" {
		t.Errorf("Expected 'Test error', got '%s'", err.Error())
	}
}

// TestAppError_Unwrap verifies Unwrap() method
func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original")
	err := Internal("Wrapped", originalErr)

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Error("Expected Unwrap to return original error")
	}
}

// TestWrap_NilError verifies Wrap with nil error
func TestWrap_NilError(t *testing.T) {
	wrapped := Wrap(nil, "context")

	if wrapped != nil {
		t.Error("Expected Wrap of nil to return nil")
	}
}

// TestWrap_AppError verifies Wrap preserves AppError
func TestWrap_AppError(t *testing.T) {
	original := NotFound("Patient", "123")
	wrapped := Wrap(original, "Failed to fetch patient")

	if wrapped.Code != "RESOURCE_NOT_FOUND" {
		t.Errorf("Expected code to be preserved, got %s", wrapped.Code)
	}
	if wrapped.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status to be preserved, got %d", wrapped.StatusCode)
	}
	expectedMsg := "Failed to fetch patient: Patient with ID '123' not found"
	if wrapped.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, wrapped.Message)
	}
}

// TestWrap_GenericError verifies Wrap converts generic error to AppError
func TestWrap_GenericError(t *testing.T) {
	originalErr := errors.New("database error")
	wrapped := Wrap(originalErr, "Failed to query")

	if wrapped.Code != "INTERNAL_ERROR" {
		t.Errorf("Expected code INTERNAL_ERROR, got %s", wrapped.Code)
	}
	if wrapped.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", wrapped.StatusCode)
	}
	if wrapped.Err != originalErr {
		t.Error("Expected wrapped error to contain original")
	}
}
