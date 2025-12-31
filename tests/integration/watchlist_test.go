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
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWatchlist_Empty(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	watchlistHandler := handler.NewWatchlistHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/watchlist", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		watchlistHandler.GetWatchlist(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/watchlist", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	watchlist := resp["watchlist"].([]interface{})
	assert.Len(t, watchlist, 0)
	assert.Equal(t, float64(0), resp["total"])
}

func TestAddToWatchlist(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	watchlistHandler := handler.NewWatchlistHandler(db, logger)

	r := chi.NewRouter()
	r.Post("/api/auctions/{id}/watch", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		watchlistHandler.AddToWatchlist(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/watch", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify in database
	var count int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM watchlist WHERE user_id = $1 AND auction_id = $2", userID, auctionID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestAddToWatchlist_Duplicate(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	// Add to watchlist first
	db.Exec(t.Context(), "INSERT INTO watchlist (user_id, auction_id) VALUES ($1, $2)", userID, auctionID)

	watchlistHandler := handler.NewWatchlistHandler(db, logger)

	r := chi.NewRouter()
	r.Post("/api/auctions/{id}/watch", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		watchlistHandler.AddToWatchlist(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/watch", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Should succeed (ON CONFLICT DO NOTHING)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Still only one entry
	var count int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM watchlist WHERE user_id = $1 AND auction_id = $2", userID, auctionID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestRemoveFromWatchlist(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	// Add to watchlist first
	db.Exec(t.Context(), "INSERT INTO watchlist (user_id, auction_id) VALUES ($1, $2)", userID, auctionID)

	watchlistHandler := handler.NewWatchlistHandler(db, logger)

	r := chi.NewRouter()
	r.Delete("/api/auctions/{id}/watch", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		watchlistHandler.RemoveFromWatchlist(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("DELETE", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/watch", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify removed
	var count int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM watchlist WHERE user_id = $1 AND auction_id = $2", userID, auctionID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestIsWatching(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	watchlistHandler := handler.NewWatchlistHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/auctions/{id}/watching", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		watchlistHandler.IsWatching(w, r.WithContext(ctx))
	})

	// Not watching
	req := httptest.NewRequest("GET", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/watching", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]bool
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.False(t, resp["watching"])

	// Add to watchlist
	db.Exec(t.Context(), "INSERT INTO watchlist (user_id, auction_id) VALUES ($1, $2)", userID, auctionID)

	// Now watching
	req = httptest.NewRequest("GET", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/watching", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.True(t, resp["watching"])
}

func TestGetWatchlist_WithData(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	// Add to watchlist
	db.Exec(t.Context(), "INSERT INTO watchlist (user_id, auction_id) VALUES ($1, $2)", userID, auctionID)

	watchlistHandler := handler.NewWatchlistHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/watchlist", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		watchlistHandler.GetWatchlist(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/watchlist", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	watchlist := resp["watchlist"].([]interface{})
	assert.Len(t, watchlist, 1)
	assert.Equal(t, float64(1), resp["total"])

	item := watchlist[0].(map[string]interface{})
	assert.Equal(t, float64(auctionID), item["auction_id"])
}

