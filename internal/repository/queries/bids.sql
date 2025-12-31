-- name: CreateBid :one
-- Record a bid with its outcome
INSERT INTO bids (
    auction_id, user_id, amount, status, previous_high_bid, max_bid, is_auto_bid
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetBidByID :one
SELECT * FROM bids WHERE id = $1;

-- name: GetBidsForAuction :many
SELECT 
    b.*,
    u.first_name, u.last_name
FROM bids b
JOIN users u ON b.user_id = u.id
WHERE b.auction_id = $1
ORDER BY b.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAcceptedBidsForAuction :many
-- Only accepted bids (the "real" bid history)
SELECT 
    b.*,
    u.first_name, u.last_name
FROM bids b
JOIN users u ON b.user_id = u.id
WHERE b.auction_id = $1 AND b.status = 'accepted'
ORDER BY b.amount DESC
LIMIT $2;

-- name: GetUserBidsForAuction :many
SELECT * FROM bids
WHERE auction_id = $1 AND user_id = $2
ORDER BY created_at DESC;

-- name: GetHighestBid :one
SELECT * FROM bids
WHERE auction_id = $1 AND status = 'accepted'
ORDER BY amount DESC
LIMIT 1;

-- name: CountBidsForAuction :one
SELECT COUNT(*) FROM bids WHERE auction_id = $1;

-- name: CountAcceptedBids :one
SELECT COUNT(*) FROM bids WHERE auction_id = $1 AND status = 'accepted';

-- name: MarkBidOutbid :exec
-- Mark a user's previous accepted bid as outbid
UPDATE bids SET status = 'outbid'
WHERE auction_id = $1 
  AND user_id = $2 
  AND status = 'accepted'
  AND id != $3;

-- name: GetUserActiveBids :many
-- Get all auctions where user currently has the high bid
SELECT 
    b.*,
    a.ends_at,
    v.year, v.make, v.model
FROM bids b
JOIN auctions a ON b.auction_id = a.id
JOIN vehicles v ON a.vehicle_id = v.id
WHERE b.user_id = $1
  AND b.status = 'accepted'
  AND a.status = 'active'
  AND a.current_bid_user_id = $1
ORDER BY a.ends_at ASC;

