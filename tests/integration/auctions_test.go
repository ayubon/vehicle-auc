package integration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAuctionsEmpty(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	auctionHandler := handler.NewAuctionHandler(db, logger)

	req := httptest.NewRequest("GET", "/api/auctions", nil)
	rec := httptest.NewRecorder()

	auctionHandler.ListAuctions(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Contains(t, resp, "auctions")
	assert.Contains(t, resp, "total")
}

func TestListAuctionsWithData(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	fixtures.TestAuction(t, db, vehicleID)

	auctionHandler := handler.NewAuctionHandler(db, logger)

	req := httptest.NewRequest("GET", "/api/auctions", nil)
	rec := httptest.NewRecorder()

	auctionHandler.ListAuctions(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	auctions := resp["auctions"].([]interface{})
	assert.Len(t, auctions, 1)

	auction := auctions[0].(map[string]interface{})
	assert.Equal(t, "active", auction["status"])
	assert.Contains(t, auction, "current_bid")
	assert.Contains(t, auction, "make")
	assert.Contains(t, auction, "model")
}

func TestGetAuction(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	auctionHandler := handler.NewAuctionHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/auctions/{id}", auctionHandler.GetAuction)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/auctions/%d", auctionID), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	auction := resp["auction"].(map[string]interface{})
	assert.Equal(t, "active", auction["status"])
	assert.Equal(t, "Honda", auction["make"])
	assert.Equal(t, "Accord", auction["model"])
	assert.Contains(t, auction, "seller_first_name")
}

func TestGetAuctionNotFound(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	auctionHandler := handler.NewAuctionHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/auctions/{id}", auctionHandler.GetAuction)

	req := httptest.NewRequest("GET", "/api/auctions/99999", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetBidHistory(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test data with bids
	sellerID := fixtures.SellerUser(t, db)
	bidderID := fixtures.VerifiedUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuctionWithBid(t, db, vehicleID, 100, bidderID)

	auctionHandler := handler.NewAuctionHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/auctions/{id}/bids", auctionHandler.GetBidHistory)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/auctions/%d/bids", auctionID), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	bids := resp["bids"].([]interface{})
	assert.Len(t, bids, 1)

	bid := bids[0].(map[string]interface{})
	assert.Equal(t, "100.00", bid["amount"])
	assert.Equal(t, "accepted", bid["status"])
}

func TestAuctionWithCurrentBid(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create auction with existing bid
	sellerID := fixtures.SellerUser(t, db)
	bidderID := fixtures.VerifiedUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuctionWithBid(t, db, vehicleID, 5000, bidderID)

	auctionHandler := handler.NewAuctionHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/auctions/{id}", auctionHandler.GetAuction)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/auctions/%d", auctionID), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	auction := resp["auction"].(map[string]interface{})
	assert.Equal(t, "5000.00", auction["current_bid"])
	assert.Equal(t, float64(1), auction["bid_count"])
}

