package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ayubfarah/vehicle-auc/internal/middleware"
)

// VINHandler handles VIN decoding
type VINHandler struct {
	logger   *slog.Logger
	decoder  VINDecoder
}

// VINDecoder interface for VIN decoding services
type VINDecoder interface {
	DecodeVIN(ctx context.Context, vin string) (*VINData, error)
}

// VINData represents decoded VIN information
type VINData struct {
	VIN          string  `json:"vin"`
	Year         int     `json:"year"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Trim         string  `json:"trim,omitempty"`
	BodyType     string  `json:"body_type,omitempty"`
	Engine       string  `json:"engine,omitempty"`
	Transmission string  `json:"transmission,omitempty"`
	Drivetrain   string  `json:"drivetrain,omitempty"`
	FuelType     string  `json:"fuel_type,omitempty"`
	Doors        int     `json:"doors,omitempty"`
	MSRP         float64 `json:"msrp,omitempty"`
}

func NewVINHandler(logger *slog.Logger, decoder VINDecoder) *VINHandler {
	return &VINHandler{
		logger:  logger,
		decoder: decoder,
	}
}

// DecodeVIN decodes a VIN and returns vehicle information
func (h *VINHandler) DecodeVIN(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Require auth
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		VIN string `json:"vin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate VIN
	if len(req.VIN) != 17 {
		h.jsonError(w, "invalid VIN - must be 17 characters", http.StatusBadRequest)
		return
	}

	// Check if decoder is configured
	if h.decoder == nil {
		// Return mock data in development
		h.logger.Warn("VIN decoder not configured, returning mock data")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": VINData{
				VIN:      req.VIN,
				Year:     2021,
				Make:     "Honda",
				Model:    "Accord",
				Trim:     "Sport",
				BodyType: "Sedan",
				Engine:   "1.5L Turbo I4",
				Transmission: "CVT",
				Drivetrain: "FWD",
				FuelType: "Gasoline",
				Doors:    4,
			},
			"mock": true,
		})
		return
	}

	// Decode VIN
	data, err := h.decoder.DecodeVIN(ctx, req.VIN)
	if err != nil {
		h.logger.Error("VIN decode failed",
			slog.String("vin", req.VIN),
			slog.String("error", err.Error()),
		)
		h.jsonError(w, "failed to decode VIN: "+err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("vin_decoded",
		slog.String("vin", req.VIN),
		slog.Int64("user_id", userID),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *VINHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

