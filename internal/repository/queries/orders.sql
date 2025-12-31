-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1;

-- name: GetOrderByAuctionID :one
SELECT * FROM orders WHERE auction_id = $1;

-- name: GetOrdersForBuyer :many
SELECT 
    o.*,
    v.year, v.make, v.model, v.vin
FROM orders o
JOIN vehicles v ON o.vehicle_id = v.id
WHERE o.buyer_id = $1
ORDER BY o.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOrdersForSeller :many
SELECT 
    o.*,
    v.year, v.make, v.model, v.vin,
    u.first_name as buyer_first_name, u.last_name as buyer_last_name
FROM orders o
JOIN vehicles v ON o.vehicle_id = v.id
JOIN users u ON o.buyer_id = u.id
WHERE o.seller_id = $1
ORDER BY o.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateOrder :one
INSERT INTO orders (
    auction_id, buyer_id, seller_id, vehicle_id,
    sale_price, buyer_premium, seller_fee, total_price, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE orders SET status = $2 WHERE id = $1 RETURNING *;

-- name: MarkOrderPaid :one
UPDATE orders SET
    status = 'paid',
    payment_intent_id = $2,
    paid_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetFulfillmentByOrderID :one
SELECT * FROM fulfillments WHERE order_id = $1;

-- name: CreateFulfillment :one
INSERT INTO fulfillments (
    order_id, status, pickup_address, delivery_address
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateFulfillmentStatus :one
UPDATE fulfillments SET status = $2 WHERE id = $1 RETURNING *;

-- name: UpdateFulfillmentTracking :one
UPDATE fulfillments SET
    carrier = $2,
    tracking_number = $3,
    estimated_delivery = $4
WHERE id = $1
RETURNING *;

