package realtime

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBroker_StartStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)

	broker.Start()
	// Should not panic
	broker.Stop()
}

func TestBroker_Subscribe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	auctionID := int64(42)
	sub := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   1,
		Messages: make(chan []byte, 10),
		Done:     make(chan struct{}),
	}

	broker.Subscribe(auctionID, sub)

	// Should be in subscribers
	broker.mu.RLock()
	subs := broker.subscribers[auctionID]
	broker.mu.RUnlock()
	assert.Len(t, subs, 1)
}

func TestBroker_Unsubscribe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	auctionID := int64(42)
	sub := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   1,
		Messages: make(chan []byte, 10),
		Done:     make(chan struct{}),
	}

	broker.Subscribe(auctionID, sub)
	broker.Unsubscribe(auctionID, sub)

	broker.mu.RLock()
	subs := broker.subscribers[auctionID]
	broker.mu.RUnlock()
	assert.Len(t, subs, 0)
}

func TestBroker_Broadcast(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	auctionID := int64(42)
	sub := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   1,
		Messages: make(chan []byte, 10),
		Done:     make(chan struct{}),
	}

	broker.Subscribe(auctionID, sub)

	event := domain.BidEvent{
		Type:      "bid_accepted",
		AuctionID: auctionID,
		Amount:    decimal.NewFromFloat(100.00),
	}

	broker.Broadcast(event)

	// Should receive event
	select {
	case received := <-sub.Messages:
		assert.Contains(t, string(received), "bid_accepted")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("did not receive event")
	}
}

func TestBroker_BroadcastToMultipleSubscribers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	auctionID := int64(42)

	subs := make([]*Subscriber, 3)
	for i := 0; i < 3; i++ {
		subs[i] = &Subscriber{
			ID:       uuid.New().String(),
			UserID:   int64(i + 1),
			Messages: make(chan []byte, 10),
			Done:     make(chan struct{}),
		}
		broker.Subscribe(auctionID, subs[i])
	}

	event := domain.BidEvent{
		Type:      "bid_accepted",
		AuctionID: auctionID,
		Amount:    decimal.NewFromFloat(100.00),
	}

	broker.Broadcast(event)

	// All should receive
	for i, sub := range subs {
		select {
		case <-sub.Messages:
			// good
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("subscriber %d did not receive event", i)
		}
	}
}

func TestBroker_BroadcastOnlyToTargetAuction(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	auction42 := int64(42)
	auction99 := int64(99)

	sub42 := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   1,
		Messages: make(chan []byte, 10),
		Done:     make(chan struct{}),
	}
	sub99 := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   2,
		Messages: make(chan []byte, 10),
		Done:     make(chan struct{}),
	}

	broker.Subscribe(auction42, sub42)
	broker.Subscribe(auction99, sub99)

	event := domain.BidEvent{
		Type:      "bid_accepted",
		AuctionID: auction42,
		Amount:    decimal.NewFromFloat(100.00),
	}

	broker.Broadcast(event)

	// 42 should receive
	select {
	case <-sub42.Messages:
		// good
	case <-time.After(200 * time.Millisecond):
		t.Fatal("auction 42 did not receive")
	}

	// 99 should NOT receive
	select {
	case <-sub99.Messages:
		t.Fatal("auction 99 should not receive")
	case <-time.After(50 * time.Millisecond):
		// good
	}
}

func TestBroker_Stats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	// Subscribe to different auctions
	for i := 0; i < 2; i++ {
		sub := &Subscriber{
			ID:       uuid.New().String(),
			UserID:   int64(i + 1),
			Messages: make(chan []byte, 10),
			Done:     make(chan struct{}),
		}
		broker.Subscribe(42, sub)
	}

	sub99 := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   3,
		Messages: make(chan []byte, 10),
		Done:     make(chan struct{}),
	}
	broker.Subscribe(99, sub99)

	stats := broker.Stats()

	assert.Equal(t, 3, stats.TotalConnections)
	assert.Len(t, stats.Auctions, 2)
}

func TestBroker_SlowSubscriber(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	broker := NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	auctionID := int64(42)
	sub := &Subscriber{
		ID:       uuid.New().String(),
		UserID:   1,
		Messages: make(chan []byte, 5), // Small buffer
		Done:     make(chan struct{}),
	}

	broker.Subscribe(auctionID, sub)

	// Send many events (should not block)
	for i := 0; i < 20; i++ {
		broker.Broadcast(domain.BidEvent{
			Type:      "bid_accepted",
			AuctionID: auctionID,
			Amount:    decimal.NewFromInt(int64(i * 10)),
		})
	}

	// Should not panic - verify broker is still responsive
	time.Sleep(100 * time.Millisecond)

	// Drain some messages
	count := 0
	for {
		select {
		case <-sub.Messages:
			count++
		case <-time.After(50 * time.Millisecond):
			goto done
		}
	}
done:
	assert.True(t, count > 0)
}
