-- name: GetAuctionByID :one
SELECT * FROM auctions WHERE id = $1;

-- name: GetAuctionForUpdate :one
-- Used by bid engine for OCC - gets current state
SELECT id, vehicle_id, status, current_bid, current_bid_user_id, bid_count, version, ends_at, extension_count, max_extensions, snipe_threshold_minutes, extension_minutes
FROM auctions
WHERE id = $1;

-- name: GetAuctionWithVehicle :one
SELECT 
    a.*,
    v.vin, v.year, v.make, v.model, v.trim, v.mileage, 
    v.starting_price, v.exterior_color, v.location_city, v.location_state,
    u.first_name as seller_first_name, u.last_name as seller_last_name
FROM auctions a
JOIN vehicles v ON a.vehicle_id = v.id
JOIN users u ON v.seller_id = u.id
WHERE a.id = $1;

-- name: ListActiveAuctions :many
SELECT 
    a.*,
    v.year, v.make, v.model, v.trim, v.mileage, v.starting_price,
    v.exterior_color, v.location_city, v.location_state
FROM auctions a
JOIN vehicles v ON a.vehicle_id = v.id
WHERE a.status = 'active'
  AND a.ends_at > NOW()
ORDER BY a.ends_at ASC
LIMIT $1 OFFSET $2;

-- name: ListEndingSoonAuctions :many
SELECT 
    a.*,
    v.year, v.make, v.model, v.trim
FROM auctions a
JOIN vehicles v ON a.vehicle_id = v.id
WHERE a.status = 'active'
  AND a.ends_at > NOW()
  AND a.ends_at < NOW() + INTERVAL '1 hour'
ORDER BY a.ends_at ASC
LIMIT $1;

-- name: CreateAuction :one
INSERT INTO auctions (
    vehicle_id, status, starts_at, ends_at, 
    max_extensions, snipe_threshold_minutes, extension_minutes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateAuctionBidOCC :one
-- OCC update: only succeeds if version matches
-- Returns the updated row or nothing if version mismatch
UPDATE auctions SET
    current_bid = $2,
    current_bid_user_id = $3,
    bid_count = bid_count + 1,
    version = version + 1
WHERE id = $1 AND version = $4
RETURNING *;

-- name: ExtendAuctionOCC :one
-- Extend auction end time (anti-snipe), with OCC
UPDATE auctions SET
    ends_at = ends_at + (extension_minutes * INTERVAL '1 minute'),
    extension_count = extension_count + 1,
    version = version + 1
WHERE id = $1 
  AND version = $2
  AND extension_count < max_extensions
RETURNING *;

-- name: UpdateAuctionStatus :one
UPDATE auctions SET status = $2 WHERE id = $1 RETURNING *;

-- name: EndAuction :one
UPDATE auctions SET
    status = 'ended',
    winner_id = current_bid_user_id,
    winning_bid = current_bid
WHERE id = $1
RETURNING *;

-- name: GetAuctionsToEnd :many
-- Find auctions that should be ended (cron job)
SELECT * FROM auctions
WHERE status = 'active' AND ends_at <= NOW()
LIMIT 100;

-- name: GetAuctionsToStart :many
-- Find auctions that should be started (cron job)
SELECT * FROM auctions
WHERE status = 'scheduled' AND starts_at <= NOW()
LIMIT 100;

