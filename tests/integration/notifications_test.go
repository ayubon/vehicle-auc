package integration

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestNotification(t *testing.T, db *pgxpool.Pool, userID int64, title, notifType string) int64 {
	var id int64
	err := db.QueryRow(t.Context(),
		`INSERT INTO notifications (user_id, type, title, message) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, notifType, title, "Test message",
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestGetNotifications_Empty(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/notifications", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.GetNotifications(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/notifications", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	notifications := resp["notifications"].([]interface{})
	assert.Len(t, notifications, 0)
	assert.Equal(t, float64(0), resp["total"])
	assert.Equal(t, float64(0), resp["unread"])
}

func TestGetNotifications_WithData(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)

	// Create notifications
	createTestNotification(t, db, userID, "Notification 1", "bid_outbid")
	createTestNotification(t, db, userID, "Notification 2", "auction_won")

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/notifications", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.GetNotifications(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/notifications", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	notifications := resp["notifications"].([]interface{})
	assert.Len(t, notifications, 2)
	assert.Equal(t, float64(2), resp["total"])
	assert.Equal(t, float64(2), resp["unread"])
}

func TestGetUnreadCount(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)

	// Create notifications
	createTestNotification(t, db, userID, "Unread 1", "bid_outbid")
	createTestNotification(t, db, userID, "Unread 2", "auction_won")
	notif3 := createTestNotification(t, db, userID, "Read", "bid_accepted")

	// Mark one as read
	db.Exec(t.Context(), "UPDATE notifications SET read_at = NOW() WHERE id = $1", notif3)

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/notifications/unread-count", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.GetUnreadCount(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/notifications/unread-count", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]int64
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, int64(2), resp["unread"])
}

func TestMarkRead(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	notifID := createTestNotification(t, db, userID, "To Read", "bid_outbid")

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Post("/api/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.MarkRead(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("POST", "/api/notifications/"+strconv.FormatInt(notifID, 10)+"/read", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify marked as read
	var readAt interface{}
	db.QueryRow(t.Context(), "SELECT read_at FROM notifications WHERE id = $1", notifID).Scan(&readAt)
	assert.NotNil(t, readAt)
}

func TestMarkRead_AlreadyRead(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	notifID := createTestNotification(t, db, userID, "Already Read", "bid_outbid")

	// Mark as read
	db.Exec(t.Context(), "UPDATE notifications SET read_at = NOW() WHERE id = $1", notifID)

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Post("/api/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.MarkRead(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("POST", "/api/notifications/"+strconv.FormatInt(notifID, 10)+"/read", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMarkAllRead(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)

	// Create multiple unread notifications
	createTestNotification(t, db, userID, "Unread 1", "bid_outbid")
	createTestNotification(t, db, userID, "Unread 2", "auction_won")
	createTestNotification(t, db, userID, "Unread 3", "bid_accepted")

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Post("/api/notifications/read-all", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.MarkAllRead(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("POST", "/api/notifications/read-all", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify all marked as read
	var unreadCount int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL", userID).Scan(&unreadCount)
	assert.Equal(t, 0, unreadCount)
}

func TestDeleteNotification(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	notifID := createTestNotification(t, db, userID, "To Delete", "bid_outbid")

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Delete("/api/notifications/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.DeleteNotification(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("DELETE", "/api/notifications/"+strconv.FormatInt(notifID, 10), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify deleted
	var count int
	db.QueryRow(t.Context(), "SELECT COUNT(*) FROM notifications WHERE id = $1", notifID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestDeleteNotification_NotOwned(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)
	otherUserID := fixtures.CreateUser(t, db, "other@example.com", "Other", "User")
	notifID := createTestNotification(t, db, otherUserID, "Other's Notification", "bid_outbid")

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Delete("/api/notifications/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.DeleteNotification(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("DELETE", "/api/notifications/"+strconv.FormatInt(notifID, 10), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetNotifications_UnreadOnly(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.BuyerUser(t, db)

	// Create notifications
	createTestNotification(t, db, userID, "Unread", "bid_outbid")
	readNotif := createTestNotification(t, db, userID, "Read", "auction_won")

	// Mark one as read
	db.Exec(t.Context(), "UPDATE notifications SET read_at = NOW() WHERE id = $1", readNotif)

	notifHandler := handler.NewNotificationHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/notifications", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		notifHandler.GetNotifications(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/notifications?unread=true", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	notifications := resp["notifications"].([]interface{})
	assert.Len(t, notifications, 1)
}

