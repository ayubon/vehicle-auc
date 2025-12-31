package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/bidengine"
	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/internal/realtime"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBidTestServer(t *testing.T, db *pgxpool.Pool, engine *bidengine.Engine, logger *slog.Logger) *chi.Mux {
	bidHandler := handler.NewBidHandler(engine, logger)

	r := chi.NewRouter()
	r.Post("/api/auctions/{id}/bids", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("test_user_id").(int64)
		ctx := middleware.WithUserID(r.Context(), userID)
		bidHandler.PlaceBid(w, r.WithContext(ctx))
	})
	r.Get("/api/bids/{ticketId}/status", bidHandler.GetBidStatus)
	return r
}

func TestPlaceBid_Success(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Setup
	buyerID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	// Create bid engine in sync mode
	broker := realtime.NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	engine := bidengine.NewEngine(db, logger, broker,
		bidengine.WithSyncMode(true),
		bidengine.WithMaxRetries(3),
	)
	engine.Start()
	defer engine.Stop()

	r := setupBidTestServer(t, db, engine, logger)

	body := map[string]string{"amount": "150.00"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/bids", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "test_user_id", buyerID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp["ticket_id"])
	assert.Equal(t, "queued", resp["status"])

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify auction updated
	var currentBid float64
	db.QueryRow(t.Context(), "SELECT current_bid FROM auctions WHERE id = $1", auctionID).Scan(&currentBid)
	assert.Equal(t, 150.00, currentBid)
}

func TestPlaceBid_InvalidAmount(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	buyerID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	broker := realtime.NewBroker(logger)
	engine := bidengine.NewEngine(db, logger, broker, bidengine.WithSyncMode(true))

	r := setupBidTestServer(t, db, engine, logger)

	// Negative amount
	body := map[string]string{"amount": "-50.00"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/bids", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "test_user_id", buyerID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPlaceBid_TooLow(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	buyerID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	// Set current bid to 200
	_, err := db.Exec(context.Background(), "UPDATE auctions SET current_bid = 200, bid_count = 1 WHERE id = $1", auctionID)
	require.NoError(t, err)

	broker := realtime.NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	engine := bidengine.NewEngine(db, logger, broker, bidengine.WithSyncMode(true))
	engine.Start()
	defer engine.Stop()

	r := setupBidTestServer(t, db, engine, logger)

	// Bid lower than current (should be rejected)
	body := map[string]string{"amount": "150.00"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/bids", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "test_user_id", buyerID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code) // Still accepted (async)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify auction NOT updated (bid was too low)
	var currentBid float64
	db.QueryRow(context.Background(), "SELECT current_bid FROM auctions WHERE id = $1", auctionID).Scan(&currentBid)
	assert.Equal(t, 200.00, currentBid) // Should still be 200, not 150
}

func TestPlaceBid_VerifyBidRecorded(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	buyerID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	broker := realtime.NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	engine := bidengine.NewEngine(db, logger, broker, bidengine.WithSyncMode(true))
	engine.Start()
	defer engine.Stop()

	r := setupBidTestServer(t, db, engine, logger)

	body := map[string]string{"amount": "175.00"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/bids", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "test_user_id", buyerID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	time.Sleep(100 * time.Millisecond)

	// Verify bid was recorded in bids table
	var bidCount int
	var bidAmount float64
	var bidUserID int64
	err := db.QueryRow(context.Background(), `
		SELECT COUNT(*), SUM(amount), MAX(user_id) FROM bids WHERE auction_id = $1
	`, auctionID).Scan(&bidCount, &bidAmount, &bidUserID)
	require.NoError(t, err)

	assert.Equal(t, 1, bidCount)
	assert.Equal(t, 175.00, bidAmount)
	assert.Equal(t, buyerID, bidUserID)

	// Verify auction state
	var auctionBidCount int
	var currentBidUserID *int64
	db.QueryRow(context.Background(), `
		SELECT bid_count, current_bid_user_id FROM auctions WHERE id = $1
	`, auctionID).Scan(&auctionBidCount, &currentBidUserID)
	
	assert.Equal(t, 1, auctionBidCount)
	assert.NotNil(t, currentBidUserID)
	assert.Equal(t, buyerID, *currentBidUserID)
}

func TestPlaceBid_OCC_VersionIncremented(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	buyerID := fixtures.BuyerUser(t, db)
	sellerID := fixtures.SellerUser(t, db)
	vehicleID := fixtures.TestVehicle(t, db, sellerID)
	auctionID := fixtures.TestAuction(t, db, vehicleID)

	// Get initial version
	var initialVersion int
	db.QueryRow(context.Background(), "SELECT version FROM auctions WHERE id = $1", auctionID).Scan(&initialVersion)

	broker := realtime.NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	engine := bidengine.NewEngine(db, logger, broker, bidengine.WithSyncMode(true))
	engine.Start()
	defer engine.Stop()

	r := setupBidTestServer(t, db, engine, logger)

	body := map[string]string{"amount": "100.00"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auctions/"+strconv.FormatInt(auctionID, 10)+"/bids", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "test_user_id", buyerID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	time.Sleep(100 * time.Millisecond)

	// Verify version was incremented (OCC)
	var newVersion int
	db.QueryRow(context.Background(), "SELECT version FROM auctions WHERE id = $1", auctionID).Scan(&newVersion)
	assert.Equal(t, initialVersion+1, newVersion)
}
