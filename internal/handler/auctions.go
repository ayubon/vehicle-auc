package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuctionHandler struct {
	db       *pgxpool.Pool
	logger   *slog.Logger
	validate *validator.Validate
}

func NewAuctionHandler(db *pgxpool.Pool, logger *slog.Logger) *AuctionHandler {
	return &AuctionHandler{
		db:       db,
		logger:   logger,
		validate: validator.New(),
	}
}

type AuctionResponse struct {
	ID                int64   `json:"id"`
	VehicleID         int64   `json:"vehicle_id"`
	Status            string  `json:"status"`
	StartsAt          string  `json:"starts_at"`
	EndsAt            string  `json:"ends_at"`
	CurrentBid        string  `json:"current_bid"`
	CurrentBidUserID  *int64  `json:"current_bid_user_id,omitempty"`
	BidCount          int     `json:"bid_count"`
	
	// Vehicle info (joined)
	Year              int     `json:"year,omitempty"`
	Make              string  `json:"make,omitempty"`
	Model             string  `json:"model,omitempty"`
	Trim              *string `json:"trim,omitempty"`
	Mileage           *int    `json:"mileage,omitempty"`
	StartingPrice     string  `json:"starting_price,omitempty"`
	ExteriorColor     *string `json:"exterior_color,omitempty"`
	LocationCity      *string `json:"location_city,omitempty"`
	LocationState     *string `json:"location_state,omitempty"`
}

// ListAuctions returns active auctions
func (h *AuctionHandler) ListAuctions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	limit := 20
	offset := 0
	
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}
	
	query := `
		SELECT a.id, a.vehicle_id, a.status::text, a.starts_at, a.ends_at,
		       a.current_bid, a.current_bid_user_id, a.bid_count,
		       v.year, v.make, v.model, v.trim, v.mileage,
		       v.starting_price, v.exterior_color, v.location_city, v.location_state
		FROM auctions a
		JOIN vehicles v ON a.vehicle_id = v.id
		WHERE a.status::text = $1
		ORDER BY a.ends_at ASC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := h.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		h.logger.Error("failed to query auctions", slog.String("error", err.Error()))
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	auctions := make([]AuctionResponse, 0)
	for rows.Next() {
		var a AuctionResponse
		var startsAt, endsAt time.Time
		var currentBid, startingPrice float64
		
		err := rows.Scan(
			&a.ID, &a.VehicleID, &a.Status, &startsAt, &endsAt,
			&currentBid, &a.CurrentBidUserID, &a.BidCount,
			&a.Year, &a.Make, &a.Model, &a.Trim, &a.Mileage,
			&startingPrice, &a.ExteriorColor, &a.LocationCity, &a.LocationState,
		)
		if err != nil {
			h.logger.Error("failed to scan auction", slog.String("error", err.Error()))
			continue
		}
		
		a.StartsAt = startsAt.Format(time.RFC3339)
		a.EndsAt = endsAt.Format(time.RFC3339)
		a.CurrentBid = strconv.FormatFloat(currentBid, 'f', 2, 64)
		a.StartingPrice = strconv.FormatFloat(startingPrice, 'f', 2, 64)
		
		auctions = append(auctions, a)
	}
	
	// Get total count
	var total int64
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM auctions WHERE status::text = $1`, status).Scan(&total)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"auctions": auctions,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": int64(offset+len(auctions)) < total,
	})
}

// GetAuction returns a single auction with full details
func (h *AuctionHandler) GetAuction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	
	query := `
		SELECT a.id, a.vehicle_id, a.status::text, a.starts_at, a.ends_at,
		       a.current_bid, a.current_bid_user_id, a.bid_count,
		       a.extension_count, a.max_extensions,
		       v.vin, v.year, v.make, v.model, v.trim, v.mileage,
		       v.starting_price, v.exterior_color, v.description,
		       v.location_city, v.location_state,
		       u.first_name as seller_first_name, u.last_name as seller_last_name
		FROM auctions a
		JOIN vehicles v ON a.vehicle_id = v.id
		JOIN users u ON v.seller_id = u.id
		WHERE a.id = $1
	`
	
	var auction struct {
		AuctionResponse
		VIN             string  `json:"vin"`
		Description     *string `json:"description,omitempty"`
		ExtensionCount  int     `json:"extension_count"`
		MaxExtensions   int     `json:"max_extensions"`
		SellerFirstName *string `json:"seller_first_name,omitempty"`
		SellerLastName  *string `json:"seller_last_name,omitempty"`
	}
	
	var startsAt, endsAt time.Time
	var currentBid, startingPrice float64
	
	err = h.db.QueryRow(ctx, query, id).Scan(
		&auction.ID, &auction.VehicleID, &auction.Status, &startsAt, &endsAt,
		&currentBid, &auction.CurrentBidUserID, &auction.BidCount,
		&auction.ExtensionCount, &auction.MaxExtensions,
		&auction.VIN, &auction.Year, &auction.Make, &auction.Model,
		&auction.Trim, &auction.Mileage, &startingPrice,
		&auction.ExteriorColor, &auction.Description,
		&auction.LocationCity, &auction.LocationState,
		&auction.SellerFirstName, &auction.SellerLastName,
	)
	
	if err != nil {
		h.jsonError(w, "auction not found", http.StatusNotFound)
		return
	}
	
	auction.StartsAt = startsAt.Format(time.RFC3339)
	auction.EndsAt = endsAt.Format(time.RFC3339)
	auction.CurrentBid = strconv.FormatFloat(currentBid, 'f', 2, 64)
	auction.StartingPrice = strconv.FormatFloat(startingPrice, 'f', 2, 64)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"auction": auction,
	})
}

// CreateAuction creates a new auction for a vehicle
func (h *AuctionHandler) CreateAuction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	
	var req struct {
		VehicleID     int64  `json:"vehicle_id" validate:"required"`
		StartsAt      string `json:"starts_at" validate:"required"`
		EndsAt        string `json:"ends_at" validate:"required"`
		MaxExtensions int    `json:"max_extensions"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.validate.Struct(req); err != nil {
		h.jsonError(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		h.jsonError(w, "invalid starts_at format (use RFC3339)", http.StatusBadRequest)
		return
	}
	
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		h.jsonError(w, "invalid ends_at format (use RFC3339)", http.StatusBadRequest)
		return
	}
	
	if endsAt.Before(startsAt) {
		h.jsonError(w, "ends_at must be after starts_at", http.StatusBadRequest)
		return
	}
	
	// Verify user owns the vehicle
	var vehicleOwnerID int64
	err = h.db.QueryRow(ctx, `SELECT seller_id FROM vehicles WHERE id = $1`, req.VehicleID).Scan(&vehicleOwnerID)
	if err != nil {
		h.jsonError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	
	if vehicleOwnerID != userID {
		h.jsonError(w, "not authorized to auction this vehicle", http.StatusForbidden)
		return
	}
	
	// Determine initial status
	status := "scheduled"
	if startsAt.Before(time.Now()) {
		status = "active"
	}
	
	maxExtensions := req.MaxExtensions
	if maxExtensions == 0 {
		maxExtensions = 10
	}
	
	query := `
		INSERT INTO auctions (vehicle_id, status, starts_at, ends_at, max_extensions)
		VALUES ($1, $2::auction_status, $3, $4, $5)
		RETURNING id
	`
	
	var auctionID int64
	err = h.db.QueryRow(ctx, query, req.VehicleID, status, startsAt, endsAt, maxExtensions).Scan(&auctionID)
	if err != nil {
		h.logger.Error("failed to create auction", slog.String("error", err.Error()))
		h.jsonError(w, "failed to create auction", http.StatusInternalServerError)
		return
	}
	
	// Update vehicle status
	h.db.Exec(ctx, `UPDATE vehicles SET status = 'active' WHERE id = $1`, req.VehicleID)
	
	h.logger.Info("auction_created",
		slog.Int64("auction_id", auctionID),
		slog.Int64("vehicle_id", req.VehicleID),
		slog.Int64("seller_id", userID),
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"auction_id": auctionID,
		"status":     status,
		"message":    "Auction created successfully",
	})
}

// GetBidHistory returns bid history for an auction
func (h *AuctionHandler) GetBidHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	idStr := chi.URLParam(r, "id")
	auctionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	query := `
		SELECT b.id, b.amount, b.status::text, b.previous_high_bid, b.created_at,
		       u.first_name, u.last_name
		FROM bids b
		JOIN users u ON b.user_id = u.id
		WHERE b.auction_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2
	`
	
	rows, err := h.db.Query(ctx, query, auctionID, limit)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	type BidHistoryItem struct {
		ID              int64   `json:"id"`
		Amount          string  `json:"amount"`
		Status          string  `json:"status"`
		PreviousHighBid *string `json:"previous_high_bid,omitempty"`
		CreatedAt       string  `json:"created_at"`
		BidderFirstName *string `json:"bidder_first_name,omitempty"`
		BidderLastName  *string `json:"bidder_last_name,omitempty"`
	}
	
	bids := make([]BidHistoryItem, 0)
	for rows.Next() {
		var b BidHistoryItem
		var amount float64
		var previousHighBid *float64
		var createdAt time.Time
		
		err := rows.Scan(
			&b.ID, &amount, &b.Status, &previousHighBid, &createdAt,
			&b.BidderFirstName, &b.BidderLastName,
		)
		if err != nil {
			continue
		}
		
		b.Amount = strconv.FormatFloat(amount, 'f', 2, 64)
		b.CreatedAt = createdAt.Format(time.RFC3339)
		if previousHighBid != nil {
			s := strconv.FormatFloat(*previousHighBid, 'f', 2, 64)
			b.PreviousHighBid = &s
		}
		
		bids = append(bids, b)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bids": bids,
	})
}

func (h *AuctionHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

