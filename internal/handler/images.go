package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/config"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImageHandler handles vehicle image operations
type ImageHandler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	cfg    *config.Config
	s3     S3Presigner
}

// S3Presigner interface for generating presigned URLs
type S3Presigner interface {
	GenerateUploadURL(ctx context.Context, bucket, key, contentType string, expires time.Duration) (string, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}

func NewImageHandler(db *pgxpool.Pool, logger *slog.Logger, cfg *config.Config, s3 S3Presigner) *ImageHandler {
	return &ImageHandler{
		db:     db,
		logger: logger,
		cfg:    cfg,
		s3:     s3,
	}
}

// GetUploadURL generates a presigned S3 URL for uploading
func (h *ImageHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	// Check ownership
	var sellerID int64
	err = h.db.QueryRow(ctx, `SELECT seller_id FROM vehicles WHERE id = $1`, vehicleID).Scan(&sellerID)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	if sellerID != userID {
		h.jsonError(w, "not authorized", http.StatusForbidden)
		return
	}

	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Filename == "" {
		req.Filename = "image.jpg"
	}
	if req.ContentType == "" {
		req.ContentType = "image/jpeg"
	}

	// Generate unique S3 key
	s3Key := fmt.Sprintf("vehicles/%d/%s-%s", vehicleID, uuid.New().String()[:8], req.Filename)

	// Generate presigned URL (if S3 client configured)
	var uploadURL string
	if h.s3 != nil {
		uploadURL, err = h.s3.GenerateUploadURL(ctx, h.cfg.AWSS3Bucket, s3Key, req.ContentType, 15*time.Minute)
		if err != nil {
			h.logger.Error("failed to generate upload URL", slog.String("error", err.Error()))
			h.jsonError(w, "failed to generate upload URL", http.StatusInternalServerError)
			return
		}
	} else {
		// Development mode - return mock URL
		uploadURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s?mock=true", h.cfg.AWSS3Bucket, h.cfg.AWSS3Region, s3Key)
	}

	// Construct the final URL (without query params)
	finalURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", h.cfg.AWSS3Bucket, h.cfg.AWSS3Region, s3Key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"upload_url": uploadURL,
		"s3_key":     s3Key,
		"url":        finalURL,
		"public_url": finalURL, // Frontend expects this field name
	})
}

// AddImage registers an uploaded image with a vehicle
func (h *ImageHandler) AddImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	// Check ownership
	var sellerID int64
	err = h.db.QueryRow(ctx, `SELECT seller_id FROM vehicles WHERE id = $1`, vehicleID).Scan(&sellerID)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	if sellerID != userID {
		h.jsonError(w, "not authorized", http.StatusForbidden)
		return
	}

	var req struct {
		S3Key     string `json:"s3_key"`
		URL       string `json:"url"`
		IsPrimary bool   `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.S3Key == "" || req.URL == "" {
		h.jsonError(w, "s3_key and url are required", http.StatusBadRequest)
		return
	}

	// If marking as primary, unset other primary images
	if req.IsPrimary {
		h.db.Exec(ctx, `UPDATE vehicle_images SET is_primary = false WHERE vehicle_id = $1`, vehicleID)
	}

	// Get next display order
	var maxOrder int
	h.db.QueryRow(ctx, `SELECT COALESCE(MAX(display_order), 0) FROM vehicle_images WHERE vehicle_id = $1`, vehicleID).Scan(&maxOrder)

	var imageID int64
	err = h.db.QueryRow(ctx, `
		INSERT INTO vehicle_images (vehicle_id, s3_key, url, is_primary, display_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, vehicleID, req.S3Key, req.URL, req.IsPrimary, maxOrder+1).Scan(&imageID)

	if err != nil {
		h.logger.Error("failed to add image", slog.String("error", err.Error()))
		h.jsonError(w, "failed to add image", http.StatusInternalServerError)
		return
	}

	h.logger.Info("image_added",
		slog.Int64("image_id", imageID),
		slog.Int64("vehicle_id", vehicleID),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Image added",
		"image_id":   imageID,
		"is_primary": req.IsPrimary,
	})
}

// DeleteImage removes an image from a vehicle
func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	vehicleIDStr := chi.URLParam(r, "id")
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid vehicle id", http.StatusBadRequest)
		return
	}

	imageIDStr := chi.URLParam(r, "imageId")
	imageID, err := strconv.ParseInt(imageIDStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid image id", http.StatusBadRequest)
		return
	}

	// Check ownership
	var sellerID int64
	err = h.db.QueryRow(ctx, `SELECT seller_id FROM vehicles WHERE id = $1`, vehicleID).Scan(&sellerID)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	if sellerID != userID {
		h.jsonError(w, "not authorized", http.StatusForbidden)
		return
	}

	// Get image s3_key for deletion
	var s3Key string
	var imgVehicleID int64
	err = h.db.QueryRow(ctx, `SELECT s3_key, vehicle_id FROM vehicle_images WHERE id = $1`, imageID).Scan(&s3Key, &imgVehicleID)
	if err != nil || imgVehicleID != vehicleID {
		h.jsonError(w, "image not found", http.StatusNotFound)
		return
	}

	// Delete from S3 if client configured
	if h.s3 != nil {
		if err := h.s3.DeleteObject(ctx, h.cfg.AWSS3Bucket, s3Key); err != nil {
			h.logger.Warn("failed to delete from S3", slog.String("error", err.Error()), slog.String("s3_key", s3Key))
		}
	}

	// Delete from database
	_, err = h.db.Exec(ctx, `DELETE FROM vehicle_images WHERE id = $1`, imageID)
	if err != nil {
		h.jsonError(w, "failed to delete image", http.StatusInternalServerError)
		return
	}

	h.logger.Info("image_deleted", slog.Int64("image_id", imageID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Image deleted"})
}

func (h *ImageHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

