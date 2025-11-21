package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// TestLogger_Success verifies logging for successful requests (2xx)
func TestLogger_Success(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)

	// Create test handler that returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap handler with logging middleware
	loggedHandler := Logger(testLogger)(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, request)

	// Verify log output contains expected fields
	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, `"method":"GET"`) {
		t.Error("Expected log to contain method")
	}
	if !strings.Contains(logOutput, `"path":"/test"`) {
		t.Error("Expected log to contain path")
	}
	if !strings.Contains(logOutput, `"status":200`) {
		t.Error("Expected log to contain status 200")
	}
	if !strings.Contains(logOutput, `"level":"info"`) {
		t.Error("Expected log level to be info for 2xx status")
	}
}

// TestLogger_ClientError verifies logging for 4xx errors as warnings
func TestLogger_ClientError(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)

	// Create test handler that returns 404 Not Found
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	loggedHandler := Logger(testLogger)(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, request)

	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, `"status":404`) {
		t.Error("Expected log to contain status 404")
	}
	if !strings.Contains(logOutput, `"level":"warn"`) {
		t.Error("Expected log level to be warn for 4xx status")
	}
}

// TestLogger_ServerError verifies logging for 5xx errors
func TestLogger_ServerError(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)

	// Create test handler that returns 500 Internal Server Error
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	loggedHandler := Logger(testLogger)(testHandler)

	request := httptest.NewRequest(http.MethodPost, "/error", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, request)

	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, `"status":500`) {
		t.Error("Expected log to contain status 500")
	}
	if !strings.Contains(logOutput, `"level":"error"`) {
		t.Error("Expected log level to be error for 5xx status")
	}
}

// TestLogger_BytesWritten verifies byte count is logged
func TestLogger_BytesWritten(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)

	responseBody := "test response body"
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(responseBody))
	})

	loggedHandler := Logger(testLogger)(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, request)

	logOutput := logBuffer.String()

	// Verify bytes written matches response body length
	expectedBytes := len(responseBody)
	if !strings.Contains(logOutput, `"bytes":`+string(rune(expectedBytes+'0'))) {
		if !strings.Contains(logOutput, `"bytes":18`) { // length of "test response body"
			t.Errorf("Expected log to contain bytes=%d", expectedBytes)
		}
	}
}

// TestLogger_Duration verifies duration is logged
func TestLogger_Duration(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := Logger(testLogger)(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, request)

	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, `"duration_ms"`) {
		t.Error("Expected log to contain duration_ms field")
	}
}

// TestLogger_UserAgent verifies user agent is logged
func TestLogger_UserAgent(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := zerolog.New(&logBuffer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := Logger(testLogger)(testHandler)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("User-Agent", "TestAgent/1.0")
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, request)

	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, `"user_agent":"TestAgent/1.0"`) {
		t.Error("Expected log to contain user agent")
	}
}

// TestResponseWriter_WriteHeader verifies status code capture
func TestResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrappedWriter := newResponseWriter(recorder)

	wrappedWriter.WriteHeader(http.StatusCreated)

	if wrappedWriter.statusCode != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", wrappedWriter.statusCode)
	}
}

// TestResponseWriter_Write verifies byte count capture
func TestResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrappedWriter := newResponseWriter(recorder)

	testData := []byte("test data")
	bytesWritten, writeError := wrappedWriter.Write(testData)

	if writeError != nil {
		t.Errorf("Expected no error, got %v", writeError)
	}
	if bytesWritten != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), bytesWritten)
	}
	if wrappedWriter.bytesWritten != len(testData) {
		t.Errorf("Expected bytesWritten to be %d, got %d", len(testData), wrappedWriter.bytesWritten)
	}
}

// TestResponseWriter_DefaultStatus verifies default status is 200
func TestResponseWriter_DefaultStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrappedWriter := newResponseWriter(recorder)

	if wrappedWriter.statusCode != http.StatusOK {
		t.Errorf("Expected default status 200, got %d", wrappedWriter.statusCode)
	}
}
