package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// BidRequest is sent to the bid engine
type BidRequest struct {
	TicketID  string          `json:"ticket_id"`
	AuctionID int64           `json:"auction_id"`
	UserID    int64           `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	MaxBid    decimal.Decimal `json:"max_bid,omitempty"` // For auto-bidding
	TraceID   string          `json:"trace_id,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// BidResult is the outcome of processing a bid
type BidResult struct {
	TicketID        string          `json:"ticket_id"`
	Status          string          `json:"status"` // "accepted", "rejected", "error"
	Reason          string          `json:"reason,omitempty"`
	BidID           int64           `json:"bid_id,omitempty"`
	Amount          decimal.Decimal `json:"amount"`
	PreviousHighBid decimal.Decimal `json:"previous_high_bid,omitempty"`
	NewHighBid      decimal.Decimal `json:"new_high_bid,omitempty"`
	AuctionID       int64           `json:"auction_id"`
	ProcessedAt     time.Time       `json:"processed_at"`
	Retries         int             `json:"retries,omitempty"`
}

// BidEvent is broadcast to SSE subscribers
type BidEvent struct {
	Type             string          `json:"type"` // "bid_accepted", "bid_outbid", "auction_extended"
	AuctionID        int64           `json:"auction_id"`
	Amount           decimal.Decimal `json:"amount,omitempty"`
	BidderID         int64           `json:"bidder_id,omitempty"`
	BidCount         int             `json:"bid_count,omitempty"`
	EndsAt           time.Time       `json:"ends_at,omitempty"`
	ExtensionApplied bool            `json:"extension_applied,omitempty"`
	Timestamp        time.Time       `json:"timestamp"`
}

// SSEMessage wraps events for SSE transmission
type SSEMessage struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// AuctionState holds the current state for OCC operations
type AuctionState struct {
	ID                 int64
	Status             string
	CurrentBid         decimal.Decimal
	CurrentBidUserID   *int64
	BidCount           int
	Version            int
	EndsAt             time.Time
	ExtensionCount     int
	MaxExtensions      int
	SnipeThresholdMins int
	ExtensionMins      int
}

// User verification status
type UserVerification struct {
	UserID     int64
	CanBid     bool
	Reason     string
	VerifiedAt *time.Time
}

// Pagination
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	HasMore    bool  `json:"has_more"`
}

// API response wrappers
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BidSubmitResponse struct {
	TicketID string `json:"ticket_id"`
	Status   string `json:"status"` // "queued"
	Message  string `json:"message"`
}

