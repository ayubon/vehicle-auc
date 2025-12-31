package bidengine

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Worker processes bids for a single auction
type Worker struct {
	auctionID    int64
	db           *pgxpool.Pool
	logger       *slog.Logger
	broadcaster  Broadcaster
	maxRetries   int
	retryBackoff time.Duration
	
	// Internal queue
	queue        chan domain.BidRequest
	
	// Callbacks
	OnResult     func(ticketID string, result domain.BidResult)
	OnComplete   func()
	OnRetry      func()
	
	// Stats
	processed    atomic.Int64
	lastBidAt    atomic.Int64 // Unix timestamp
	
	// Lifecycle
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// WorkerStats for debug endpoints
type WorkerStats struct {
	AuctionID   int64  `json:"auction_id"`
	QueueDepth  int    `json:"queue_depth"`
	Processed   int64  `json:"processed"`
	LastBidAt   string `json:"last_bid_at,omitempty"`
	IdleFor     string `json:"idle_for,omitempty"`
}

// NewWorker creates a new auction worker
func NewWorker(auctionID int64, db *pgxpool.Pool, logger *slog.Logger, broadcaster Broadcaster, maxRetries int, retryBackoff time.Duration) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Worker{
		auctionID:    auctionID,
		db:           db,
		logger:       logger,
		broadcaster:  broadcaster,
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
		queue:        make(chan domain.BidRequest, 100),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the worker goroutine
func (w *Worker) Start() {
	w.wg.Add(1)
	go w.run()
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	w.cancel()
	w.wg.Wait()
}

// Submit sends a bid to this worker
func (w *Worker) Submit(req domain.BidRequest) {
	select {
	case w.queue <- req:
	case <-w.ctx.Done():
	}
}

// Stats returns worker statistics
func (w *Worker) Stats() WorkerStats {
	lastBid := time.Unix(w.lastBidAt.Load(), 0)
	
	stats := WorkerStats{
		AuctionID:  w.auctionID,
		QueueDepth: len(w.queue),
		Processed:  w.processed.Load(),
	}
	
	if !lastBid.IsZero() && lastBid.Unix() > 0 {
		stats.LastBidAt = lastBid.Format(time.RFC3339)
		stats.IdleFor = time.Since(lastBid).Round(time.Second).String()
	}
	
	return stats
}

func (w *Worker) run() {
	defer w.wg.Done()
	
	processor := &BidProcessor{
		db:           w.db,
		logger:       w.logger,
		broadcaster:  w.broadcaster,
		maxRetries:   w.maxRetries,
		retryBackoff: w.retryBackoff,
		onRetry:      w.OnRetry,
	}
	
	for {
		select {
		case <-w.ctx.Done():
			return
		case req := <-w.queue:
			result := processor.Process(w.ctx, req)
			
			w.processed.Add(1)
			w.lastBidAt.Store(time.Now().Unix())
			
			if w.OnResult != nil {
				w.OnResult(req.TicketID, result)
			}
			if w.OnComplete != nil {
				w.OnComplete()
			}
		}
	}
}

