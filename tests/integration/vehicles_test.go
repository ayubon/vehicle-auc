package integration

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListVehiclesEmpty(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	req := httptest.NewRequest("GET", "/api/vehicles", nil)
	rec := httptest.NewRecorder()

	vehicleHandler.ListVehicles(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Contains(t, resp, "vehicles")
	assert.Contains(t, resp, "total")
	assert.Contains(t, resp, "limit")
	assert.Contains(t, resp, "offset")
}

func TestListVehiclesWithData(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data
	sellerID := fixtures.SellerUser(t, db)
	fixtures.TestVehicle(t, db, sellerID)
	fixtures.TestVehicleWithDetails(t, db, sellerID, 2022, "Toyota", "Camry", 20000)

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	req := httptest.NewRequest("GET", "/api/vehicles", nil)
	rec := httptest.NewRecorder()

	vehicleHandler.ListVehicles(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	vehicles := resp["vehicles"].([]interface{})
	assert.Len(t, vehicles, 2)
}

func TestListVehiclesFilterByMake(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data
	sellerID := fixtures.SellerUser(t, db)
	fixtures.TestVehicle(t, db, sellerID)                                        // Honda
	fixtures.TestVehicleWithDetails(t, db, sellerID, 2022, "Toyota", "Camry", 20000) // Toyota

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	req := httptest.NewRequest("GET", "/api/vehicles?make=Honda", nil)
	rec := httptest.NewRecorder()

	vehicleHandler.ListVehicles(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	vehicles := resp["vehicles"].([]interface{})
	assert.Len(t, vehicles, 1)

	vehicle := vehicles[0].(map[string]interface{})
	assert.Equal(t, "Honda", vehicle["make"])
}

func TestGetVehicle(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	// Setup router to extract URL params
	r := chi.NewRouter()
	r.Get("/api/vehicles/{id}", vehicleHandler.GetVehicle)

	req := httptest.NewRequest("GET", "/api/vehicles/"+itoa(vehicleID), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	vehicle := resp["vehicle"].(map[string]interface{})
	assert.Equal(t, "Honda", vehicle["make"])
	assert.Equal(t, "Accord", vehicle["model"])
}

func TestGetVehicleNotFound(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/vehicles/{id}", vehicleHandler.GetVehicle)

	req := httptest.NewRequest("GET", "/api/vehicles/99999", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListVehiclesPagination(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data
	sellerID := fixtures.SellerUser(t, db)
	for i := 0; i < 5; i++ {
		fixtures.TestVehicleWithDetails(t, db, sellerID, 2020+i, "Test", "Model", float64(10000+i*1000))
	}

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	// Test limit
	req := httptest.NewRequest("GET", "/api/vehicles?limit=2", nil)
	rec := httptest.NewRecorder()
	vehicleHandler.ListVehicles(rec, req)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	vehicles := resp["vehicles"].([]interface{})
	assert.Len(t, vehicles, 2)
	assert.True(t, resp["has_more"].(bool))

	// Test offset
	req = httptest.NewRequest("GET", "/api/vehicles?limit=2&offset=2", nil)
	rec = httptest.NewRecorder()
	vehicleHandler.ListVehicles(rec, req)

	json.Unmarshal(rec.Body.Bytes(), &resp)
	vehicles = resp["vehicles"].([]interface{})
	assert.Len(t, vehicles, 2)
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

