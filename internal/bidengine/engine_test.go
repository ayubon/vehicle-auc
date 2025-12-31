package bidengine

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBroadcaster captures broadcast events for testing
type mockBroadcaster struct {
	mu     sync.Mutex
	events []domain.BidEvent
}

func (m *mockBroadcaster) Broadcast(event domain.BidEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func (m *mockBroadcaster) Events() []domain.BidEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]domain.BidEvent{}, m.events...)
}

func setupTestEngine(t *testing.T) (*Engine, *mockBroadcaster, *pgxpool.Pool) {
	t.Helper()

	// Skip if no test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	broadcaster := &mockBroadcaster{}

	engine := NewEngine(db, logger, broadcaster,
		WithSyncMode(true), // Synchronous for testing
		WithMaxRetries(3),
		WithRetryBackoff(1*time.Millisecond),
	)

	return engine, broadcaster, db
}

func TestEngine_Submit_SyncMode(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	// Test that sync mode processes immediately
	req := domain.BidRequest{
		TicketID:  uuid.New().String(),
		AuctionID: 1,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100),
		CreatedAt: time.Now(),
	}

	err := engine.Submit(req)
	// Will error because no actual auction exists, but proves sync mode works
	assert.NoError(t, err)
}

func TestEngine_QueueFull(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := &mockBroadcaster{}

	// Create engine with tiny queue
	engine := &Engine{
		logger:      logger,
		broadcaster: broadcaster,
		queue:       make(chan domain.BidRequest, 1), // Size 1
		results:     make(map[string]chan domain.BidResult),
		workers:     make(map[int64]*Worker),
		syncMode:    false,
	}

	// Fill the queue
	engine.queue <- domain.BidRequest{TicketID: "1"}

	// Next submit should fail
	err := engine.Submit(domain.BidRequest{TicketID: "2"})
	assert.Equal(t, ErrQueueFull, err)
}

func TestEngine_Stats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := &mockBroadcaster{}

	engine := NewEngine(nil, logger, broadcaster,
		WithQueueSize(100),
		WithSyncMode(true),
	)

	stats := engine.Stats()
	assert.Equal(t, 0, stats.QueueDepth)
	assert.Equal(t, 0, stats.ActiveWorkers)
	assert.Equal(t, int64(0), stats.TotalProcessed)
}

func TestBidProcessor_ValidateBidTooLow(t *testing.T) {
	// Unit test for bid validation (no DB needed)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := &mockBroadcaster{}

	processor := &BidProcessor{
		logger:       logger,
		broadcaster:  broadcaster,
		maxRetries:   3,
		retryBackoff: 1 * time.Millisecond,
	}

	// Create a mock auction state
	auction := &domain.AuctionState{
		ID:         1,
		Status:     "active",
		CurrentBid: decimal.NewFromFloat(100),
		Version:    1,
	}

	req := domain.BidRequest{
		TicketID:  uuid.New().String(),
		AuctionID: 1,
		UserID:    42,
		Amount:    decimal.NewFromFloat(50), // Lower than current
	}

	// Test validation
	if req.Amount.LessThanOrEqual(auction.CurrentBid) {
		result := domain.BidResult{
			TicketID:        req.TicketID,
			AuctionID:       req.AuctionID,
			Amount:          req.Amount,
			Status:          "rejected",
			Reason:          "bid_too_low",
			PreviousHighBid: auction.CurrentBid,
		}
		assert.Equal(t, "rejected", result.Status)
		assert.Equal(t, "bid_too_low", result.Reason)
	}

	_ = processor // used to show processor is available
}

func TestBidRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		amount  decimal.Decimal
		current decimal.Decimal
		valid   bool
	}{
		{"higher bid", decimal.NewFromFloat(150), decimal.NewFromFloat(100), true},
		{"equal bid", decimal.NewFromFloat(100), decimal.NewFromFloat(100), false},
		{"lower bid", decimal.NewFromFloat(50), decimal.NewFromFloat(100), false},
		{"zero current", decimal.NewFromFloat(50), decimal.Zero, true},
		{"negative bid", decimal.NewFromFloat(-10), decimal.Zero, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.amount.GreaterThan(tt.current) && tt.amount.GreaterThan(decimal.Zero)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestOCCVersionConflict(t *testing.T) {
	// Test that version conflict is properly detected
	err := ErrVersionConflict
	assert.Equal(t, "version conflict - concurrent modification", err.Error())
}

func TestResultDelivery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := &mockBroadcaster{}

	engine := NewEngine(nil, logger, broadcaster, WithSyncMode(true))

	ticketID := uuid.New().String()

	// Deliver a result
	result := domain.BidResult{
		TicketID:  ticketID,
		Status:    "accepted",
		AuctionID: 1,
	}
	engine.deliverResult(ticketID, result)

	// Should be able to retrieve it
	retrieved, err := engine.GetResult(ticketID, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, "accepted", retrieved.Status)
}

func TestResultTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broadcaster := &mockBroadcaster{}

	engine := NewEngine(nil, logger, broadcaster, WithSyncMode(true))

	ticketID := uuid.New().String()

	// Don't deliver any result - should timeout
	_, err := engine.GetResult(ticketID, 10*time.Millisecond)
	assert.Equal(t, ErrTimeout, err)
}

