package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/tests/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClerkSync_NewUser(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	authHandler := handler.NewAuthHandler(db, logger)

	body := map[string]string{
		"clerk_user_id": "clerk_test_123",
		"email":         "newuser@example.com",
		"first_name":    "New",
		"last_name":     "User",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/clerk-sync", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.ClerkSync(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	user := resp["user"].(map[string]interface{})
	assert.Equal(t, "newuser@example.com", user["email"])
	assert.Equal(t, "New", user["first_name"])
	assert.Equal(t, "User", user["last_name"])
	assert.Equal(t, "buyer", user["role"])
	assert.Equal(t, false, user["can_bid"])
}

func TestClerkSync_ExistingUser(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create existing user
	existingEmail := "existing@example.com"
	fixtures.CreateUser(t, db, existingEmail, "Existing", "User")

	authHandler := handler.NewAuthHandler(db, logger)

	body := map[string]string{
		"clerk_user_id": "clerk_existing_123",
		"email":         existingEmail,
		"first_name":    "Updated",
		"last_name":     "Name",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/clerk-sync", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.ClerkSync(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	user := resp["user"].(map[string]interface{})
	assert.Equal(t, existingEmail, user["email"])
}

func TestClerkSync_MissingFields(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	authHandler := handler.NewAuthHandler(db, logger)

	// Missing clerk_user_id
	body := map[string]string{
		"email": "test@example.com",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/clerk-sync", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.ClerkSync(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMe_Authenticated(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.CreateUser(t, db, "me@example.com", "Test", "User")

	authHandler := handler.NewAuthHandler(db, logger)

	r := chi.NewRouter()
	r.Get("/api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		authHandler.Me(w, r.WithContext(ctx))
	})

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "me@example.com", resp["email"])
	assert.Equal(t, "Test", resp["first_name"])
}

func TestMe_Unauthenticated(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	authHandler := handler.NewAuthHandler(db, logger)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	authHandler.Me(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUpdateProfile(t *testing.T) {
	db := fixtures.SetupTestDBWithMigrations(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	userID := fixtures.CreateUser(t, db, "update@example.com", "Old", "Name")

	authHandler := handler.NewAuthHandler(db, logger)

	r := chi.NewRouter()
	r.Put("/api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		ctx := middleware.WithUserID(r.Context(), userID)
		authHandler.UpdateProfile(w, r.WithContext(ctx))
	})

	body := map[string]string{
		"first_name": "New",
		"last_name":  "Updated",
		"phone":      "555-1234",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/auth/me", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify update
	var firstName, lastName, phone string
	db.QueryRow(t.Context(), "SELECT first_name, last_name, phone FROM users WHERE id = $1", userID).
		Scan(&firstName, &lastName, &phone)
	assert.Equal(t, "New", firstName)
	assert.Equal(t, "Updated", lastName)
	assert.Equal(t, "555-1234", phone)
}

