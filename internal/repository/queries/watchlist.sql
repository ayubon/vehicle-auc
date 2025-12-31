-- name: AddToWatchlist :one
INSERT INTO watchlist (user_id, auction_id)
VALUES ($1, $2)
ON CONFLICT (user_id, auction_id) DO NOTHING
RETURNING *;

-- name: RemoveFromWatchlist :exec
DELETE FROM watchlist WHERE user_id = $1 AND auction_id = $2;

-- name: GetUserWatchlist :many
SELECT 
    w.*,
    a.status as auction_status, a.current_bid, a.ends_at,
    v.year, v.make, v.model, v.trim
FROM watchlist w
JOIN auctions a ON w.auction_id = a.id
JOIN vehicles v ON a.vehicle_id = v.id
WHERE w.user_id = $1
ORDER BY a.ends_at ASC
LIMIT $2 OFFSET $3;

-- name: IsWatching :one
SELECT EXISTS(
    SELECT 1 FROM watchlist WHERE user_id = $1 AND auction_id = $2
) as watching;

-- name: GetWatchersForAuction :many
SELECT u.id, u.email, u.first_name
FROM watchlist w
JOIN users u ON w.user_id = u.id
WHERE w.auction_id = $1;

