package realtime

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/ayubfarah/vehicle-auc/internal/metrics"
)

// Broker manages SSE connections and broadcasts events
type Broker struct {
	logger *slog.Logger
	
	// Per-auction subscribers
	subscribers map[int64]map[*Subscriber]struct{}
	mu          sync.RWMutex
	
	// Event channel for broadcasting
	events chan domain.BidEvent
	
	// Lifecycle
	done chan struct{}
}

// Subscriber represents an SSE client connection
type Subscriber struct {
	ID       string
	UserID   int64
	Messages chan []byte
	Done     chan struct{}
}

// NewBroker creates a new SSE broker
func NewBroker(logger *slog.Logger) *Broker {
	b := &Broker{
		logger:      logger,
		subscribers: make(map[int64]map[*Subscriber]struct{}),
		events:      make(chan domain.BidEvent, 1000),
		done:        make(chan struct{}),
	}
	return b
}

// Start begins the broadcast loop
func (b *Broker) Start() {
	go b.broadcastLoop()
	b.logger.Info("sse_broker_started")
}

// Stop gracefully shuts down the broker
func (b *Broker) Stop() {
	close(b.done)
	b.logger.Info("sse_broker_stopped")
}

// Subscribe adds a subscriber for an auction
func (b *Broker) Subscribe(auctionID int64, sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.subscribers[auctionID] == nil {
		b.subscribers[auctionID] = make(map[*Subscriber]struct{})
	}
	b.subscribers[auctionID][sub] = struct{}{}
	
	metrics.SSEConnectionsActive.Inc()
	
	b.logger.Debug("sse_subscriber_added",
		slog.Int64("auction_id", auctionID),
		slog.String("subscriber_id", sub.ID),
	)
}

// Unsubscribe removes a subscriber
func (b *Broker) Unsubscribe(auctionID int64, sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if subs, ok := b.subscribers[auctionID]; ok {
		delete(subs, sub)
		if len(subs) == 0 {
			delete(b.subscribers, auctionID)
		}
	}
	
	metrics.SSEConnectionsActive.Dec()
	
	b.logger.Debug("sse_subscriber_removed",
		slog.Int64("auction_id", auctionID),
		slog.String("subscriber_id", sub.ID),
	)
}

// Broadcast sends an event to all subscribers of an auction
func (b *Broker) Broadcast(event domain.BidEvent) {
	select {
	case b.events <- event:
	default:
		b.logger.Warn("sse_event_dropped_queue_full",
			slog.Int64("auction_id", event.AuctionID),
		)
	}
}

func (b *Broker) broadcastLoop() {
	for {
		select {
		case <-b.done:
			return
		case event := <-b.events:
			b.broadcastEvent(event)
		}
	}
}

func (b *Broker) broadcastEvent(event domain.BidEvent) {
	b.mu.RLock()
	subs := b.subscribers[event.AuctionID]
	count := len(subs)
	b.mu.RUnlock()
	
	if count == 0 {
		return
	}
	
	// Serialize event once
	data, err := json.Marshal(event)
	if err != nil {
		b.logger.Error("sse_event_marshal_error",
			slog.String("error", err.Error()),
		)
		return
	}
	
	// Format as SSE
	message := formatSSE(event.Type, data)
	
	// Fan out to subscribers
	b.mu.RLock()
	for sub := range b.subscribers[event.AuctionID] {
		select {
		case sub.Messages <- message:
		default:
			// Subscriber buffer full, skip
		}
	}
	b.mu.RUnlock()
	
	metrics.SSESubscribersPerAuction.Observe(float64(count))
	
	b.logger.Debug("sse_event_broadcast",
		slog.Int64("auction_id", event.AuctionID),
		slog.String("event_type", event.Type),
		slog.Int("subscribers", count),
	)
}

func formatSSE(eventType string, data []byte) []byte {
	// SSE format: "event: <type>\ndata: <json>\n\n"
	result := make([]byte, 0, len(eventType)+len(data)+20)
	result = append(result, "event: "...)
	result = append(result, eventType...)
	result = append(result, '\n')
	result = append(result, "data: "...)
	result = append(result, data...)
	result = append(result, '\n', '\n')
	return result
}

// Stats returns broker statistics
func (b *Broker) Stats() BrokerStats {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	total := 0
	auctionStats := make([]AuctionSubscribers, 0, len(b.subscribers))
	
	for auctionID, subs := range b.subscribers {
		count := len(subs)
		total += count
		auctionStats = append(auctionStats, AuctionSubscribers{
			AuctionID:   auctionID,
			Subscribers: count,
		})
	}
	
	return BrokerStats{
		TotalConnections: total,
		Auctions:         auctionStats,
	}
}

// BrokerStats for debug endpoints
type BrokerStats struct {
	TotalConnections int                  `json:"total_connections"`
	Auctions         []AuctionSubscribers `json:"auctions"`
}

type AuctionSubscribers struct {
	AuctionID   int64 `json:"auction_id"`
	Subscribers int   `json:"subscribers"`
}

