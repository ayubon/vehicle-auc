package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/config"
	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUploadURL(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{
		AWSS3Bucket: "test-bucket",
		AWSS3Region: "us-east-1",
	}

	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	imageHandler := handler.NewImageHandler(db, logger, cfg, nil)

	r := chi.NewRouter()
	r.Post("/api/vehicles/{id}/upload-url", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), sellerID)
		imageHandler.GetUploadURL(w, r.WithContext(ctx))
	})

	body := map[string]string{
		"filename":     "test-image.jpg",
		"content_type": "image/jpeg",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/upload-url", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Contains(t, resp["upload_url"], "test-bucket")
	assert.Contains(t, resp["s3_key"], "vehicles/")
	assert.Contains(t, resp["url"], "test-bucket")
}

func TestGetUploadURL_NotOwner(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{
		AWSS3Bucket: "test-bucket",
		AWSS3Region: "us-east-1",
	}

	sellerID := fixtures.SellerUser(t, db)
	otherUserID := fixtures.BuyerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	imageHandler := handler.NewImageHandler(db, logger, cfg, nil)

	r := chi.NewRouter()
	r.Post("/api/vehicles/{id}/upload-url", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), otherUserID)
		imageHandler.GetUploadURL(w, r.WithContext(ctx))
	})

	body := map[string]string{"filename": "test.jpg"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/upload-url", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAddImage(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{}

	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	imageHandler := handler.NewImageHandler(db, logger, cfg, nil)

	r := chi.NewRouter()
	r.Post("/api/vehicles/{id}/images", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), sellerID)
		imageHandler.AddImage(w, r.WithContext(ctx))
	})

	body := map[string]interface{}{
		"s3_key":     "vehicles/1/test.jpg",
		"url":        "https://bucket.s3.amazonaws.com/vehicles/1/test.jpg",
		"is_primary": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/images", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NotNil(t, resp["image_id"])
	assert.Equal(t, true, resp["is_primary"])

	// Verify in database
	var count int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM vehicle_images WHERE vehicle_id = $1", vehicleID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestAddImage_MissingFields(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{}

	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	imageHandler := handler.NewImageHandler(db, logger, cfg, nil)

	r := chi.NewRouter()
	r.Post("/api/vehicles/{id}/images", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), sellerID)
		imageHandler.AddImage(w, r.WithContext(ctx))
	})

	body := map[string]interface{}{
		"s3_key": "vehicles/1/test.jpg",
		// missing url
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/images", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteImage(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{}

	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	// Create image
	var imageID int64
	db.QueryRow(t.Context(), `
		INSERT INTO vehicle_images (vehicle_id, s3_key, url, is_primary, display_order)
		VALUES ($1, 'test.jpg', 'https://example.com/test.jpg', true, 1)
		RETURNING id
	`, vehicleID).Scan(&imageID)

	imageHandler := handler.NewImageHandler(db, logger, cfg, nil)

	r := chi.NewRouter()
	r.Delete("/api/vehicles/{id}/images/{imageId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), sellerID)
		imageHandler.DeleteImage(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("DELETE", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/images/"+strconv.FormatInt(imageID, 10), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify deleted
	var count int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM vehicle_images WHERE id = $1", imageID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestDeleteImage_NotOwner(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{}

	sellerID := fixtures.SellerUser(t, db)
	otherUserID := fixtures.BuyerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	// Create image
	var imageID int64
	db.QueryRow(t.Context(), `
		INSERT INTO vehicle_images (vehicle_id, s3_key, url, is_primary, display_order)
		VALUES ($1, 'test.jpg', 'https://example.com/test.jpg', true, 1)
		RETURNING id
	`, vehicleID).Scan(&imageID)

	imageHandler := handler.NewImageHandler(db, logger, cfg, nil)

	r := chi.NewRouter()
	r.Delete("/api/vehicles/{id}/images/{imageId}", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), otherUserID)
		imageHandler.DeleteImage(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("DELETE", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/images/"+strconv.FormatInt(imageID, 10), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGetVehicleImages(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)

	// Create images
	db.Exec(t.Context(), `
		INSERT INTO vehicle_images (vehicle_id, s3_key, url, is_primary, display_order)
		VALUES ($1, 'img1.jpg', 'https://example.com/img1.jpg', true, 1),
		       ($1, 'img2.jpg', 'https://example.com/img2.jpg', false, 2)
	`, vehicleID)

	vehicleHandler := handler.NewVehicleHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/vehicles/{id}/images", vehicleHandler.GetVehicleImages)

	req := httptest.NewRequest("GET", "/api/vehicles/"+strconv.FormatInt(vehicleID, 10)+"/images", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	images := resp["images"].([]interface{})
	assert.Len(t, images, 2)
}

