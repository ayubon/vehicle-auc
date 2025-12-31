package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/bidengine"
	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/internal/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BidHandler struct {
	engine   *bidengine.Engine
	logger   *slog.Logger
	validate *validator.Validate
}

func NewBidHandler(engine *bidengine.Engine, logger *slog.Logger) *BidHandler {
	return &BidHandler{
		engine:   engine,
		logger:   logger,
		validate: validator.New(),
	}
}

type PlaceBidRequest struct {
	Amount json.Number `json:"amount" validate:"required"` // Accepts both "150.00" and 150.00
	MaxBid json.Number `json:"max_bid,omitempty"`          // For auto-bidding (future)
}

type PlaceBidResponse struct {
	TicketID string `json:"ticket_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// PlaceBid submits a bid to the engine and returns immediately
func (h *BidHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse auction ID
	auctionIDStr := chi.URLParam(r, "id")
	auctionID, err := strconv.ParseInt(auctionIDStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid auction id", http.StatusBadRequest)
		return
	}
	
	// Get user ID (from auth middleware)
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}
	
	// Parse request body
	var req PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate
	if err := h.validate.Struct(req); err != nil {
		h.jsonError(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Parse amount (json.Number handles both string "150.00" and number 150.00)
	amount, err := decimal.NewFromString(req.Amount.String())
	if err != nil {
		h.jsonError(w, "invalid bid amount", http.StatusBadRequest)
		return
	}
	
	if amount.LessThanOrEqual(decimal.Zero) {
		h.jsonError(w, "bid amount must be positive", http.StatusBadRequest)
		return
	}
	
	// Generate ticket ID for tracking
	ticketID := uuid.New().String()
	
	// Create bid request
	bidReq := domain.BidRequest{
		TicketID:  ticketID,
		AuctionID: auctionID,
		UserID:    userID,
		Amount:    amount,
		TraceID:   tracing.TraceIDFromContext(ctx),
		CreatedAt: time.Now(),
	}
	
	// Parse max bid if provided
	if req.MaxBid.String() != "" {
		maxBid, err := decimal.NewFromString(req.MaxBid.String())
		if err == nil && maxBid.GreaterThan(amount) {
			bidReq.MaxBid = maxBid
		}
	}
	
	// Submit to engine
	if err := h.engine.Submit(bidReq); err != nil {
		if err == bidengine.ErrQueueFull {
			h.jsonError(w, "system busy, please retry", http.StatusServiceUnavailable)
			return
		}
		h.jsonError(w, "failed to submit bid", http.StatusInternalServerError)
		return
	}
	
	h.logger.Info("bid_submitted",
		slog.String("ticket_id", ticketID),
		slog.Int64("auction_id", auctionID),
		slog.Int64("user_id", userID),
		slog.String("amount", amount.String()),
		slog.String("request_id", middleware.GetRequestID(ctx)),
	)
	
	// Return 202 Accepted with ticket
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(PlaceBidResponse{
		TicketID: ticketID,
		Status:   "queued",
		Message:  "Bid submitted for processing",
	})
}

// GetBidStatus checks the status of a submitted bid
func (h *BidHandler) GetBidStatus(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		h.jsonError(w, "ticket_id required", http.StatusBadRequest)
		return
	}
	
	// Wait for result with short timeout
	result, err := h.engine.GetResult(ticketID, 5*time.Second)
	if err == bidengine.ErrTimeout {
		// Still processing
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"ticket_id": ticketID,
			"status":    "processing",
		})
		return
	}
	
	if err != nil {
		h.jsonError(w, "failed to get result", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *BidHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

