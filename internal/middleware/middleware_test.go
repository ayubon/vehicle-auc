package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID_GeneratesID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		assert.NotEmpty(t, reqID)
		w.Write([]byte(reqID))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should set header
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestRequestID_UsesProvidedID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		w.Write([]byte(reqID))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "custom-id-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	assert.Equal(t, "custom-id-123", string(body))
	assert.Equal(t, "custom-id-123", rec.Header().Get("X-Request-ID"))
}

func TestLogging_LogsRequest(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserContext_WithAndGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	// Initially no user
	userID := GetUserID(req.Context())
	assert.Equal(t, int64(0), userID)

	// Set user
	ctx := WithUserID(req.Context(), 42)
	userID = GetUserID(ctx)
	assert.Equal(t, int64(42), userID)
}

func TestGetRequestID_ReturnsEmpty(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	reqID := GetRequestID(req.Context())
	assert.Empty(t, reqID)
}

