package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationHandler handles notification operations
type NotificationHandler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewNotificationHandler(db *pgxpool.Pool, logger *slog.Logger) *NotificationHandler {
	return &NotificationHandler{
		db:     db,
		logger: logger,
	}
}

// GetNotifications returns user's notifications
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, _ := strconv.Atoi(l); parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, _ := strconv.Atoi(o); parsed >= 0 {
			offset = parsed
		}
	}

	unreadOnly := r.URL.Query().Get("unread") == "true"

	var query string
	var args []interface{}
	if unreadOnly {
		query = `
			SELECT id, type, title, message, data, read_at, created_at
			FROM notifications
			WHERE user_id = $1 AND read_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{userID, limit, offset}
	} else {
		query = `
			SELECT id, type, title, message, data, read_at, created_at
			FROM notifications
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{userID, limit, offset}
	}

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	notifications := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id                    int64
			notifType, title      string
			message               *string
			data                  []byte
			readAt                *time.Time
			createdAt             time.Time
		)
		rows.Scan(&id, &notifType, &title, &message, &data, &readAt, &createdAt)

		notif := map[string]interface{}{
			"id":         id,
			"type":       notifType,
			"title":      title,
			"message":    message,
			"read":       readAt != nil,
			"created_at": createdAt.Format(time.RFC3339),
		}
		if data != nil {
			var parsedData interface{}
			if json.Unmarshal(data, &parsedData) == nil {
				notif["data"] = parsedData
			}
		}
		notifications = append(notifications, notif)
	}

	// Get counts
	var total, unread int64
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&total)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&unread)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
		"total":         total,
		"unread":        unread,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetUnreadCount returns count of unread notifications
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var count int64
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"unread": count})
}

// MarkRead marks a notification as read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	notifID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(ctx, `
		UPDATE notifications SET read_at = NOW()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL
	`, notifID, userID)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		h.jsonError(w, "notification not found or already read", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
}

// MarkAllRead marks all notifications as read
func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE notifications SET read_at = NOW()
		WHERE user_id = $1 AND read_at IS NULL
	`, userID)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "All notifications marked as read"})
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	notifID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.jsonError(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(ctx, `DELETE FROM notifications WHERE id = $1 AND user_id = $2`, notifID, userID)
	if err != nil {
		h.jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		h.jsonError(w, "notification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification deleted"})
}

func (h *NotificationHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

