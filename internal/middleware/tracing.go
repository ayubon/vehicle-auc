package middleware

import (
	"context"
	"net/http"

	"github.com/ayubfarah/vehicle-auc/internal/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel"
)

// Tracing middleware adds OpenTelemetry spans to requests
func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Start span
		ctx, span := tracing.StartSpan(ctx, r.URL.Path)
		defer span.End()

		// Add HTTP attributes
		span.SetAttributes(
			semconv.HTTPMethod(r.Method),
			semconv.HTTPURL(r.URL.String()),
			semconv.HTTPRoute(r.URL.Path),
			attribute.String("http.client_ip", r.RemoteAddr),
		)

		// Add trace ID to context for logging
		ctx = context.WithValue(ctx, TraceIDKey, tracing.TraceIDFromContext(ctx))

		// Wrap response writer to capture status
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Add response attributes
		span.SetAttributes(
			semconv.HTTPStatusCode(wrapped.status),
		)
	})
}

