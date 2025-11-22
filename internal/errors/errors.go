package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Err
}

// NotFound creates a 404 Not Found error
func NotFound(resourceType string, resourceID string) *AppError {
	return &AppError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    fmt.Sprintf("%s with ID '%s' not found", resourceType, resourceID),
		StatusCode: http.StatusNotFound,
	}
}

// ValidationError creates a 400 Bad Request error for validation failures
func ValidationError(message string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// InvalidInput creates a 400 Bad Request error for invalid input
func InvalidInput(field string, reason string) *AppError {
	return &AppError{
		Code:       "INVALID_INPUT",
		Message:    fmt.Sprintf("Invalid input for field '%s': %s", field, reason),
		StatusCode: http.StatusBadRequest,
	}
}

// Internal creates a 500 Internal Server Error
func Internal(message string, err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// Conflict creates a 409 Conflict error
func Conflict(resourceType string, reason string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    fmt.Sprintf("%s conflict: %s", resourceType, reason),
		StatusCode: http.StatusConflict,
	}
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// If already an AppError, preserve it but add context
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:       appErr.Code,
			Message:    fmt.Sprintf("%s: %s", message, appErr.Message),
			StatusCode: appErr.StatusCode,
			Err:        appErr.Err,
		}
	}

	// Otherwise create internal error
	return Internal(message, err)
}
