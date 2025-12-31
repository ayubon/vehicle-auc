-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByClerkID :one
SELECT * FROM users WHERE clerk_user_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
    clerk_user_id, email, first_name, last_name, phone, role
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    phone = COALESCE($4, phone)
WHERE id = $1
RETURNING *;

-- name: VerifyUser :one
UPDATE users SET
    id_verified_at = NOW(),
    authorize_payment_profile_id = $2
WHERE id = $1
RETURNING *;

-- name: UserCanBid :one
-- Returns true if user can place bids (verified with payment profile)
SELECT 
    id_verified_at IS NOT NULL 
    AND authorize_payment_profile_id IS NOT NULL AS can_bid
FROM users
WHERE id = $1;

