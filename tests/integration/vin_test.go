package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeVIN_MockMode(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.SellerUser(t, db)

	// No decoder configured - will use mock
	vinHandler := handler.NewVINHandler(logger, nil)

	r := chi.NewRouter()
	r.Post("/api/decode-vin", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		vinHandler.DecodeVIN(w, r.WithContext(ctx))
	})

	body := map[string]string{
		"vin": "1HGBH41JXMN109186",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/decode-vin", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp["success"].(bool))
	assert.True(t, resp["mock"].(bool))

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "1HGBH41JXMN109186", data["vin"])
	assert.Equal(t, "Honda", data["make"])
	assert.Equal(t, "Accord", data["model"])
}

func TestDecodeVIN_InvalidVIN(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.SellerUser(t, db)
	vinHandler := handler.NewVINHandler(logger, nil)

	r := chi.NewRouter()
	r.Post("/api/decode-vin", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		vinHandler.DecodeVIN(w, r.WithContext(ctx))
	})

	body := map[string]string{
		"vin": "TOOSHORT",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/decode-vin", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "17 characters")
}

func TestDecodeVIN_Unauthenticated(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	vinHandler := handler.NewVINHandler(logger, nil)

	body := map[string]string{
		"vin": "1HGBH41JXMN109186",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/decode-vin", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	vinHandler.DecodeVIN(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

