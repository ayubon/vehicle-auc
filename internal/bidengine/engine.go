package bidengine

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/ayubfarah/vehicle-auc/internal/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Engine processes bids using goroutine workers with OCC
type Engine struct {
	db            *pgxpool.Pool
	logger        *slog.Logger
	broadcaster   Broadcaster
	
	// Incoming bid queue
	queue         chan domain.BidRequest
	queueSize     int
	
	// Worker management
	workers       map[int64]*Worker
	workersMu     sync.RWMutex
	maxRetries    int
	retryBackoff  time.Duration
	
	// Result delivery
	results       map[string]chan domain.BidResult
	resultsMu     sync.RWMutex
	
	// Stats
	totalProcessed atomic.Int64
	totalRetries   atomic.Int64
	
	// Lifecycle
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	
	// Testing mode
	syncMode      bool
}

// Broadcaster interface for SSE integration
type Broadcaster interface {
	Broadcast(event domain.BidEvent)
}

// EngineOption configures the engine
type EngineOption func(*Engine)

// WithSyncMode enables synchronous processing for testing
func WithSyncMode(sync bool) EngineOption {
	return func(e *Engine) {
		e.syncMode = sync
	}
}

// WithQueueSize sets the bid queue buffer size
func WithQueueSize(size int) EngineOption {
	return func(e *Engine) {
		e.queueSize = size
	}
}

// WithMaxRetries sets max OCC retries
func WithMaxRetries(retries int) EngineOption {
	return func(e *Engine) {
		e.maxRetries = retries
	}
}

// WithRetryBackoff sets the OCC retry backoff duration
func WithRetryBackoff(d time.Duration) EngineOption {
	return func(e *Engine) {
		e.retryBackoff = d
	}
}

// NewEngine creates a new bid processing engine
func NewEngine(db *pgxpool.Pool, logger *slog.Logger, broadcaster Broadcaster, opts ...EngineOption) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	
	e := &Engine{
		db:           db,
		logger:       logger,
		broadcaster:  broadcaster,
		queueSize:    10000,
		maxRetries:   3,
		retryBackoff: 10 * time.Millisecond,
		workers:      make(map[int64]*Worker),
		results:      make(map[string]chan domain.BidResult),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	for _, opt := range opts {
		opt(e)
	}
	
	e.queue = make(chan domain.BidRequest, e.queueSize)
	
	return e
}

// Start begins the dispatcher goroutine
func (e *Engine) Start() {
	if e.syncMode {
		e.logger.Info("bid_engine_started", slog.Bool("sync_mode", true))
		return
	}
	
	e.wg.Add(1)
	go e.dispatcher()
	
	e.logger.Info("bid_engine_started",
		slog.Int("queue_size", e.queueSize),
		slog.Int("max_retries", e.maxRetries),
	)
}

// Stop gracefully shuts down the engine
func (e *Engine) Stop() {
	e.logger.Info("bid_engine_stopping")
	e.cancel()
	
	// Wait for dispatcher to finish
	e.wg.Wait()
	
	// Stop all workers
	e.workersMu.Lock()
	for _, w := range e.workers {
		w.Stop()
	}
	e.workersMu.Unlock()
	
	e.logger.Info("bid_engine_stopped",
		slog.Int64("total_processed", e.totalProcessed.Load()),
	)
}

// Submit queues a bid for processing
// Returns immediately with a ticket ID
func (e *Engine) Submit(req domain.BidRequest) error {
	// In sync mode, process immediately
	if e.syncMode {
		result := e.processBidSync(req)
		e.deliverResult(req.TicketID, result)
		return nil
	}
	
	// Non-blocking send to queue
	select {
	case e.queue <- req:
		metrics.BidEngineQueueDepth.Set(float64(len(e.queue)))
		e.logger.Debug("bid_queued",
			slog.String("ticket_id", req.TicketID),
			slog.Int64("auction_id", req.AuctionID),
		)
		return nil
	default:
		return ErrQueueFull
	}
}

// GetResult waits for a bid result with timeout
func (e *Engine) GetResult(ticketID string, timeout time.Duration) (domain.BidResult, error) {
	e.resultsMu.Lock()
	ch, exists := e.results[ticketID]
	if !exists {
		ch = make(chan domain.BidResult, 1)
		e.results[ticketID] = ch
	}
	e.resultsMu.Unlock()
	
	select {
	case result := <-ch:
		e.cleanupResult(ticketID)
		return result, nil
	case <-time.After(timeout):
		e.cleanupResult(ticketID)
		return domain.BidResult{}, ErrTimeout
	}
}

func (e *Engine) cleanupResult(ticketID string) {
	e.resultsMu.Lock()
	delete(e.results, ticketID)
	e.resultsMu.Unlock()
}

func (e *Engine) deliverResult(ticketID string, result domain.BidResult) {
	e.resultsMu.Lock()
	ch, exists := e.results[ticketID]
	if !exists {
		ch = make(chan domain.BidResult, 1)
		e.results[ticketID] = ch
	}
	e.resultsMu.Unlock()
	
	// Non-blocking send
	select {
	case ch <- result:
	default:
	}
}

// dispatcher routes bids to per-auction workers
func (e *Engine) dispatcher() {
	defer e.wg.Done()
	
	for {
		select {
		case <-e.ctx.Done():
			return
		case req := <-e.queue:
			metrics.BidEngineQueueDepth.Set(float64(len(e.queue)))
			e.routeToWorker(req)
		}
	}
}

func (e *Engine) routeToWorker(req domain.BidRequest) {
	e.workersMu.Lock()
	worker, exists := e.workers[req.AuctionID]
	if !exists {
		worker = NewWorker(req.AuctionID, e.db, e.logger, e.broadcaster, e.maxRetries, e.retryBackoff)
		worker.OnResult = e.deliverResult
		worker.OnComplete = func() {
			e.totalProcessed.Add(1)
		}
		worker.OnRetry = func() {
			e.totalRetries.Add(1)
		}
		e.workers[req.AuctionID] = worker
		worker.Start()
		metrics.BidEngineWorkersActive.Set(float64(len(e.workers)))
	}
	e.workersMu.Unlock()
	
	worker.Submit(req)
}

// processBidSync processes a bid synchronously (for testing)
func (e *Engine) processBidSync(req domain.BidRequest) domain.BidResult {
	processor := &BidProcessor{
		db:           e.db,
		logger:       e.logger,
		broadcaster:  e.broadcaster,
		maxRetries:   e.maxRetries,
		retryBackoff: e.retryBackoff,
	}
	return processor.Process(context.Background(), req)
}

// Stats returns engine statistics
func (e *Engine) Stats() EngineStats {
	e.workersMu.RLock()
	workerCount := len(e.workers)
	workerStats := make([]WorkerStats, 0, workerCount)
	for _, w := range e.workers {
		workerStats = append(workerStats, w.Stats())
	}
	e.workersMu.RUnlock()
	
	return EngineStats{
		QueueDepth:     len(e.queue),
		ActiveWorkers:  workerCount,
		TotalProcessed: e.totalProcessed.Load(),
		TotalRetries:   e.totalRetries.Load(),
		Workers:        workerStats,
	}
}

// EngineStats holds engine statistics for debug endpoints
type EngineStats struct {
	QueueDepth     int           `json:"queue_depth"`
	ActiveWorkers  int           `json:"active_workers"`
	TotalProcessed int64         `json:"total_processed"`
	TotalRetries   int64         `json:"total_retries"`
	Workers        []WorkerStats `json:"workers"`
}

