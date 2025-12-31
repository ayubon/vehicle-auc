package bidengine

import "errors"

var (
	// ErrQueueFull is returned when the bid queue is at capacity
	ErrQueueFull = errors.New("bid queue is full")
	
	// ErrVersionConflict is returned when OCC detects a concurrent modification
	ErrVersionConflict = errors.New("version conflict - concurrent modification")
	
	// ErrTimeout is returned when waiting for a result times out
	ErrTimeout = errors.New("timeout waiting for bid result")
	
	// ErrAuctionNotActive is returned when bidding on a non-active auction
	ErrAuctionNotActive = errors.New("auction is not active")
	
	// ErrBidTooLow is returned when bid amount is not higher than current bid
	ErrBidTooLow = errors.New("bid amount must be higher than current bid")
	
	// ErrUserCannotBid is returned when user is not verified to bid
	ErrUserCannotBid = errors.New("user is not verified to place bids")
)

