package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/metrics"
	"github.com/ayubfarah/vehicle-auc/internal/tracing"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	size        int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.size += len(b)
	return rw.ResponseWriter.Write(b)
}

// Logging middleware with structured logging
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			// Extract IDs for logging
			requestID := GetRequestID(r.Context())
			traceID := tracing.TraceIDFromContext(r.Context())

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Record metrics
			metrics.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				http.StatusText(wrapped.status),
			).Inc()

			metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration.Seconds())

			// Log request
			logLevel := slog.LevelInfo
			if wrapped.status >= 500 {
				logLevel = slog.LevelError
			} else if wrapped.status >= 400 {
				logLevel = slog.LevelWarn
			}

			logger.Log(r.Context(), logLevel, "http_request",
				slog.String("request_id", requestID),
				slog.String("trace_id", traceID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.Int("status", wrapped.status),
				slog.Int("size", wrapped.size),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

