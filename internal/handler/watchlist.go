package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WatchlistHandler handles watchlist operations
type WatchlistHandler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewWatchlistHandler(db *pgxpool.Pool, logger *slog.Logger) *WatchlistHandler {
	return &WatchlistHandler{
		db:     db,
		logger: logger,
	}
}

// GetWatchlist returns user's watchlist
func (h *WatchlistHandler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, _ := strconv.Atoi(l); parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, _ := strconv.Atoi(o); parsed >= 0 {
			offset = parsed
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT w.id, w.auction_id, w.created_at,
		       a.status::text, a.current_bid, a.ends_at,
		       v.year, v.make, v.model, v.trim
		FROM watchlist w
		JOIN auctions a ON w.auction_id = a.id
		JOIN vehicles v ON a.vehicle_id = v.id
		WHERE w.user_id = $1
		ORDER BY a.ends_at ASC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id, auctionID                       int64
			createdAt, endsAt                   time.Time
			status                              string
			currentBid                          float64
			year                                int
			vehicleMake, model                  string
			trim                                *string
		)
		rows.Scan(&id, &auctionID, &createdAt, &status, &currentBid, &endsAt, &year, &vehicleMake, &model, &trim)
		items = append(items, map[string]interface{}{
			"id":          id,
			"auction_id":  auctionID,
			"status":      status,
			"current_bid": strconv.FormatFloat(currentBid, 'f', 2, 64),
			"ends_at":     endsAt.Format(time.RFC3339),
			"vehicle": map[string]interface{}{
				"year":  year,
				"make":  vehicleMake,
				"model": model,
				"trim":  trim,
			},
			"added_at": createdAt.Format(time.RFC3339),
		})
	}

	// Get total count
	var total int64
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM watchlist WHERE user_id = $1`, userID).Scan(&total)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"watchlist": items,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// AddToWatchlist adds an auction to user's watchlist
func (h *WatchlistHandler) AddToWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	auctionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid auction id", http.StatusBadRequest)
		return
	}

	// Check auction exists
	var exists bool
	h.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auctions WHERE id = $1)`, auctionID).Scan(&exists)
	if !exists {
		h.jsonError(w, "auction not found", http.StatusNotFound)
		return
	}

	// Add to watchlist (ignore if already exists)
	_, err = h.db.Exec(ctx, `
		INSERT INTO watchlist (user_id, auction_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, auction_id) DO NOTHING
	`, userID, auctionID)
	if err != nil {
		h.jsonError(w, "failed to add to watchlist", http.StatusInternalServerError)
		return
	}

	h.logger.Info("watchlist_added",
		slog.Int64("user_id", userID),
		slog.Int64("auction_id", auctionID),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Added to watchlist"})
}

// RemoveFromWatchlist removes an auction from user's watchlist
func (h *WatchlistHandler) RemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	auctionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid auction id", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(ctx, `DELETE FROM watchlist WHERE user_id = $1 AND auction_id = $2`, userID, auctionID)
	if err != nil {
		h.jsonError(w, "failed to remove from watchlist", http.StatusInternalServerError)
		return
	}

	h.logger.Info("watchlist_removed",
		slog.Int64("user_id", userID),
		slog.Int64("auction_id", auctionID),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Removed from watchlist"})
}

// IsWatching checks if user is watching an auction
func (h *WatchlistHandler) IsWatching(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	auctionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid auction id", http.StatusBadRequest)
		return
	}

	var watching bool
	h.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM watchlist WHERE user_id = $1 AND auction_id = $2)
	`, userID, auctionID).Scan(&watching)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"watching": watching})
}

func (h *WatchlistHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

