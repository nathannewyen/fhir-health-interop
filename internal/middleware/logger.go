package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int
}

// newResponseWriter creates a new responseWriter wrapper
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status is 200
	}
}

// WriteHeader captures the status code before writing
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the number of bytes written
func (rw *responseWriter) Write(b []byte) (int, error) {
	bytesWritten, writeError := rw.ResponseWriter.Write(b)
	rw.bytesWritten += bytesWritten
	return bytesWritten, writeError
}

// Logger middleware logs HTTP requests with structured logging using zerolog
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Wrap the response writer to capture status code and bytes
			wrappedWriter := newResponseWriter(w)

			// Process request
			next.ServeHTTP(wrappedWriter, r)

			// Calculate request duration
			duration := time.Since(startTime)

			// Determine log level based on status code
			logEvent := logger.Info()
			if wrappedWriter.statusCode >= 500 {
				logEvent = logger.Error()
			} else if wrappedWriter.statusCode >= 400 {
				logEvent = logger.Warn()
			}

			// Log structured request data
			logEvent.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", wrappedWriter.statusCode).
				Int("bytes", wrappedWriter.bytesWritten).
				Dur("duration_ms", duration).
				Str("user_agent", r.UserAgent()).
				Msg("HTTP request")
		})
	}
}
