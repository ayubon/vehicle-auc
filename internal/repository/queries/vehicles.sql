-- name: GetVehicleByID :one
SELECT * FROM vehicles WHERE id = $1;

-- name: GetVehicleByVIN :one
SELECT * FROM vehicles WHERE vin = $1;

-- name: ListVehicles :many
SELECT * FROM vehicles
WHERE status = COALESCE(sqlc.narg('status'), status)
  AND make = COALESCE(sqlc.narg('make'), make)
  AND model = COALESCE(sqlc.narg('model'), model)
  AND year >= COALESCE(sqlc.narg('min_year'), year)
  AND year <= COALESCE(sqlc.narg('max_year'), year)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountVehicles :one
SELECT COUNT(*) FROM vehicles
WHERE status = COALESCE(sqlc.narg('status'), status)
  AND make = COALESCE(sqlc.narg('make'), make)
  AND model = COALESCE(sqlc.narg('model'), model);

-- name: CreateVehicle :one
INSERT INTO vehicles (
    seller_id, vin, year, make, model, trim, body_type,
    exterior_color, interior_color, mileage, engine, transmission,
    drivetrain, fuel_type, title_status, condition_grade, description,
    starting_price, reserve_price, buy_now_price,
    location_city, location_state, location_zip, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
) RETURNING *;

-- name: UpdateVehicle :one
UPDATE vehicles SET
    year = COALESCE($2, year),
    make = COALESCE($3, make),
    model = COALESCE($4, model),
    trim = COALESCE($5, trim),
    mileage = COALESCE($6, mileage),
    description = COALESCE($7, description),
    starting_price = COALESCE($8, starting_price),
    status = COALESCE($9, status)
WHERE id = $1
RETURNING *;

-- name: UpdateVehicleStatus :one
UPDATE vehicles SET status = $2 WHERE id = $1 RETURNING *;

-- name: GetVehicleImages :many
SELECT * FROM vehicle_images 
WHERE vehicle_id = $1 
ORDER BY display_order;

-- name: CreateVehicleImage :one
INSERT INTO vehicle_images (vehicle_id, s3_key, url, is_primary, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteVehicleImage :exec
DELETE FROM vehicle_images WHERE id = $1;

