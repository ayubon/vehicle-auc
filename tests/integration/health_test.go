package integration

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)

	healthHandler := handler.NewHealthHandler(db)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", resp["status"])
	assert.Contains(t, resp, "timestamp")
	assert.Contains(t, resp, "uptime")
	assert.Contains(t, resp, "checks")
}

func TestReadyEndpoint(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)

	healthHandler := handler.NewHealthHandler(db)

	req := httptest.NewRequest("GET", "/ready", nil)
	rec := httptest.NewRecorder()

	healthHandler.Ready(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ready", rec.Body.String())
}

func TestLiveEndpoint(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)

	healthHandler := handler.NewHealthHandler(db)

	req := httptest.NewRequest("GET", "/live", nil)
	rec := httptest.NewRecorder()

	healthHandler.Live(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "alive", rec.Body.String())
}

func init() {
	// Suppress logs during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})))
}

