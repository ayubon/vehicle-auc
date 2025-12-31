package bidengine

import (
	"context"
	"log/slog"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/domain"
	"github.com/ayubfarah/vehicle-auc/internal/metrics"
	"github.com/ayubfarah/vehicle-auc/internal/tracing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
)

// BidProcessor handles the actual bid processing with OCC
type BidProcessor struct {
	db           *pgxpool.Pool
	logger       *slog.Logger
	broadcaster  Broadcaster
	maxRetries   int
	retryBackoff time.Duration
	onRetry      func()
}

// Process handles a single bid with OCC retry loop
func (p *BidProcessor) Process(ctx context.Context, req domain.BidRequest) domain.BidResult {
	start := time.Now()
	
	// Start tracing span
	ctx, span := tracing.StartSpan(ctx, "bid.process")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("ticket_id", req.TicketID),
		attribute.Int64("auction_id", req.AuctionID),
		attribute.Int64("user_id", req.UserID),
		attribute.String("amount", req.Amount.String()),
	)
	
	p.logger.Info("bid_processing_started",
		slog.String("ticket_id", req.TicketID),
		slog.Int64("auction_id", req.AuctionID),
		slog.Int64("user_id", req.UserID),
		slog.String("amount", req.Amount.String()),
	)
	
	var result domain.BidResult
	var retries int
	
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		result = p.attemptBid(ctx, req, attempt)
		
		if result.Status != "retry" {
			break
		}
		
		retries++
		if p.onRetry != nil {
			p.onRetry()
		}
		
		// Exponential backoff
		backoff := p.retryBackoff * time.Duration(1<<attempt)
		time.Sleep(backoff)
		
		p.logger.Debug("bid_occ_retry",
			slog.String("ticket_id", req.TicketID),
			slog.Int("attempt", attempt+1),
			slog.Duration("backoff", backoff),
		)
	}
	
	// Record metrics
	duration := time.Since(start)
	metrics.BidProcessingDuration.Observe(duration.Seconds())
	metrics.BidOCCRetries.Observe(float64(retries))
	metrics.AuctionBidsTotal.WithLabelValues(result.Status).Inc()
	
	result.Retries = retries
	result.ProcessedAt = time.Now()
	
	// Log final result
	p.logger.Info("bid_processing_completed",
		slog.String("ticket_id", req.TicketID),
		slog.String("status", result.Status),
		slog.Int("retries", retries),
		slog.Duration("duration", duration),
	)
	
	return result
}

func (p *BidProcessor) attemptBid(ctx context.Context, req domain.BidRequest, attempt int) domain.BidResult {
	ctx, span := tracing.StartSpan(ctx, "bid.attempt")
	defer span.End()
	span.SetAttributes(attribute.Int("attempt", attempt))
	
	// 1. Fetch current auction state
	auction, err := p.getAuctionState(ctx, req.AuctionID)
	if err != nil {
		tracing.RecordError(ctx, err)
		return domain.BidResult{
			TicketID:  req.TicketID,
			AuctionID: req.AuctionID,
			Amount:    req.Amount,
			Status:    "error",
			Reason:    "auction_not_found",
		}
	}
	
	// 2. Validate auction is active
	if auction.Status != "active" {
		return domain.BidResult{
			TicketID:  req.TicketID,
			AuctionID: req.AuctionID,
			Amount:    req.Amount,
			Status:    "rejected",
			Reason:    "auction_not_active",
		}
	}
	
	// 3. Validate bid amount
	if req.Amount.LessThanOrEqual(auction.CurrentBid) {
		return domain.BidResult{
			TicketID:        req.TicketID,
			AuctionID:       req.AuctionID,
			Amount:          req.Amount,
			Status:          "rejected",
			Reason:          "bid_too_low",
			PreviousHighBid: auction.CurrentBid,
		}
	}
	
	// 4. Attempt OCC update
	previousBid := auction.CurrentBid
	bidID, extended, err := p.updateAuctionOCC(ctx, req, auction)
	
	if err == ErrVersionConflict {
		metrics.BidOCCConflictsTotal.Inc()
		return domain.BidResult{Status: "retry"}
	}
	
	if err != nil {
		tracing.RecordError(ctx, err)
		return domain.BidResult{
			TicketID:  req.TicketID,
			AuctionID: req.AuctionID,
			Amount:    req.Amount,
			Status:    "error",
			Reason:    err.Error(),
		}
	}
	
	// 5. Broadcast to SSE subscribers
	if p.broadcaster != nil {
		event := domain.BidEvent{
			Type:             "bid_accepted",
			AuctionID:        req.AuctionID,
			Amount:           req.Amount,
			BidderID:         req.UserID,
			BidCount:         auction.BidCount + 1,
			EndsAt:           auction.EndsAt,
			ExtensionApplied: extended,
			Timestamp:        time.Now(),
		}
		p.broadcaster.Broadcast(event)
		metrics.SSEMessagesSent.WithLabelValues("bid_accepted").Inc()
		
		if extended {
			metrics.AuctionExtensions.Inc()
		}
	}
	
	return domain.BidResult{
		TicketID:        req.TicketID,
		Status:          "accepted",
		BidID:           bidID,
		Amount:          req.Amount,
		PreviousHighBid: previousBid,
		NewHighBid:      req.Amount,
		AuctionID:       req.AuctionID,
	}
}

func (p *BidProcessor) getAuctionState(ctx context.Context, auctionID int64) (*domain.AuctionState, error) {
	ctx, span := tracing.StartSpan(ctx, "db.auction.read")
	defer span.End()
	
	query := `
		SELECT id, status::text, current_bid, current_bid_user_id, bid_count, version, 
		       ends_at, extension_count, max_extensions, snipe_threshold_minutes, extension_minutes
		FROM auctions WHERE id = $1
	`
	
	var auction domain.AuctionState
	var status string
	err := p.db.QueryRow(ctx, query, auctionID).Scan(
		&auction.ID,
		&status,
		&auction.CurrentBid,
		&auction.CurrentBidUserID,
		&auction.BidCount,
		&auction.Version,
		&auction.EndsAt,
		&auction.ExtensionCount,
		&auction.MaxExtensions,
		&auction.SnipeThresholdMins,
		&auction.ExtensionMins,
	)
	
	if err != nil {
		return nil, err
	}
	
	auction.Status = status
	return &auction, nil
}

func (p *BidProcessor) updateAuctionOCC(ctx context.Context, req domain.BidRequest, auction *domain.AuctionState) (int64, bool, error) {
	ctx, span := tracing.StartSpan(ctx, "db.auction.update.occ")
	defer span.End()
	
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)
	
	// Check for snipe extension
	extended := false
	newEndsAt := auction.EndsAt
	if auction.ExtensionCount < auction.MaxExtensions {
		snipeThreshold := time.Duration(auction.SnipeThresholdMins) * time.Minute
		if time.Until(auction.EndsAt) < snipeThreshold {
			extended = true
			newEndsAt = auction.EndsAt.Add(time.Duration(auction.ExtensionMins) * time.Minute)
		}
	}
	
	// OCC update - only succeeds if version matches
	var updateQuery string
	var args []interface{}
	
	if extended {
		updateQuery = `
			UPDATE auctions SET
				current_bid = $1,
				current_bid_user_id = $2,
				bid_count = bid_count + 1,
				version = version + 1,
				ends_at = $3,
				extension_count = extension_count + 1
			WHERE id = $4 AND version = $5
			RETURNING id
		`
		args = []interface{}{req.Amount, req.UserID, newEndsAt, req.AuctionID, auction.Version}
	} else {
		updateQuery = `
			UPDATE auctions SET
				current_bid = $1,
				current_bid_user_id = $2,
				bid_count = bid_count + 1,
				version = version + 1
			WHERE id = $3 AND version = $4
			RETURNING id
		`
		args = []interface{}{req.Amount, req.UserID, req.AuctionID, auction.Version}
	}
	
	var updatedID int64
	err = tx.QueryRow(ctx, updateQuery, args...).Scan(&updatedID)
	
	if err == pgx.ErrNoRows {
		// Version mismatch - another bid won the race
		return 0, false, ErrVersionConflict
	}
	if err != nil {
		return 0, false, err
	}
	
	// Record the bid in history
	bidQuery := `
		INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, max_bid, is_auto_bid)
		VALUES ($1, $2, $3, 'accepted', $4, $5, $6)
		RETURNING id
	`
	
	var bidID int64
	err = tx.QueryRow(ctx, bidQuery,
		req.AuctionID,
		req.UserID,
		req.Amount,
		auction.CurrentBid,
		decimalOrNil(req.MaxBid),
		false,
	).Scan(&bidID)
	
	if err != nil {
		return 0, false, err
	}
	
	// Mark previous high bidder's bid as outbid
	if auction.CurrentBidUserID != nil && *auction.CurrentBidUserID != req.UserID {
		_, err = tx.Exec(ctx, `
			UPDATE bids SET status = 'outbid'
			WHERE auction_id = $1 AND user_id = $2 AND status = 'accepted'
		`, req.AuctionID, *auction.CurrentBidUserID)
		if err != nil {
			return 0, false, err
		}
	}
	
	if err := tx.Commit(ctx); err != nil {
		return 0, false, err
	}
	
	return bidID, extended, nil
}

func decimalOrNil(d decimal.Decimal) interface{} {
	if d.IsZero() {
		return nil
	}
	return d
}

