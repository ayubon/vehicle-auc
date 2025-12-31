-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, message, data)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotifications :many
SELECT * FROM notifications
WHERE user_id = $1 AND read_at IS NULL
ORDER BY created_at DESC
LIMIT $2;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND read_at IS NULL;

-- name: MarkNotificationRead :exec
UPDATE notifications SET read_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET read_at = NOW()
WHERE user_id = $1 AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1 AND user_id = $2;

