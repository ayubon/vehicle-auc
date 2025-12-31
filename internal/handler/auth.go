package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewAuthHandler(db *pgxpool.Pool, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
	}
}

// ClerkSync syncs a Clerk user with the local database
// Called from frontend after Clerk sign-in
func (h *AuthHandler) ClerkSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		ClerkUserID string `json:"clerk_user_id"`
		Email       string `json:"email"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ClerkUserID == "" || req.Email == "" {
		h.jsonError(w, "clerk_user_id and email are required", http.StatusBadRequest)
		return
	}

	// Find or create user
	var userID int64
	var isNew bool

	// Try to find by email first
	err := h.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, req.Email).Scan(&userID)
	if err != nil {
		// User doesn't exist, create new
		err = h.db.QueryRow(ctx, `
			INSERT INTO users (clerk_user_id, email, first_name, last_name, role)
			VALUES ($1, $2, $3, $4, 'buyer')
			RETURNING id
		`, req.ClerkUserID, req.Email, req.FirstName, req.LastName).Scan(&userID)
		if err != nil {
			h.logger.Error("failed to create user", slog.String("error", err.Error()))
			h.jsonError(w, "failed to create user", http.StatusInternalServerError)
			return
		}
		isNew = true
		h.logger.Info("user_created",
			slog.Int64("user_id", userID),
			slog.String("email", req.Email),
		)
	} else {
		// Update existing user with Clerk ID if not set
		_, err = h.db.Exec(ctx, `
			UPDATE users SET
				clerk_user_id = COALESCE(clerk_user_id, $1),
				first_name = COALESCE(NULLIF($2, ''), first_name),
				last_name = COALESCE(NULLIF($3, ''), last_name)
			WHERE id = $4
		`, req.ClerkUserID, req.FirstName, req.LastName, userID)
		if err != nil {
			h.logger.Error("failed to update user", slog.String("error", err.Error()))
		}
	}

	// Get full user data
	var user struct {
		ID                int64      `json:"id"`
		Email             string     `json:"email"`
		FirstName         *string    `json:"first_name"`
		LastName          *string    `json:"last_name"`
		Role              string     `json:"role"`
		IDVerifiedAt      *time.Time `json:"id_verified_at"`
		HasPaymentMethod  bool       `json:"has_payment_method"`
	}

	var paymentProfileID *string
	err = h.db.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, role, id_verified_at, authorize_payment_profile_id
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.IDVerifiedAt, &paymentProfileID)
	if err != nil {
		h.jsonError(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}
	user.HasPaymentMethod = paymentProfileID != nil && *paymentProfileID != ""

	h.logger.Info("clerk_sync",
		slog.Int64("user_id", userID),
		slog.Bool("is_new", isNew),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":                 user.ID,
			"email":              user.Email,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
			"role":               user.Role,
			"is_id_verified":     user.IDVerifiedAt != nil,
			"has_payment_method": user.HasPaymentMethod,
			"can_bid":            user.IDVerifiedAt != nil && user.HasPaymentMethod,
		},
	})
}

// Me returns the current user's profile
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var user struct {
		ID                int64      `json:"id"`
		Email             string     `json:"email"`
		FirstName         *string    `json:"first_name"`
		LastName          *string    `json:"last_name"`
		Phone             *string    `json:"phone"`
		Role              string     `json:"role"`
		IDVerifiedAt      *time.Time `json:"id_verified_at"`
		CreatedAt         time.Time  `json:"created_at"`
	}
	var paymentProfileID *string

	err := h.db.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, phone, role, id_verified_at, authorize_payment_profile_id, created_at
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Role, &user.IDVerifiedAt, &paymentProfileID, &user.CreatedAt)
	if err != nil {
		h.jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	hasPaymentMethod := paymentProfileID != nil && *paymentProfileID != ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                 user.ID,
		"email":              user.Email,
		"first_name":         user.FirstName,
		"last_name":          user.LastName,
		"phone":              user.Phone,
		"role":               user.Role,
		"is_id_verified":     user.IDVerifiedAt != nil,
		"has_payment_method": hasPaymentMethod,
		"can_bid":            user.IDVerifiedAt != nil && hasPaymentMethod,
		"created_at":         user.CreatedAt.Format(time.RFC3339),
	})
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		h.jsonError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Phone     *string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE users SET
			first_name = COALESCE($2, first_name),
			last_name = COALESCE($3, last_name),
			phone = COALESCE($4, phone)
		WHERE id = $1
	`, userID, req.FirstName, req.LastName, req.Phone)

	if err != nil {
		h.logger.Error("failed to update profile", slog.String("error", err.Error()))
		h.jsonError(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	h.logger.Info("profile_updated", slog.Int64("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated"})
}

// VerifyUser marks a user as ID verified (admin endpoint or webhook)
func (h *AuthHandler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		UserID            int64  `json:"user_id"`
		PaymentProfileID  string `json:"payment_profile_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE users SET
			id_verified_at = NOW(),
			authorize_payment_profile_id = $2
		WHERE id = $1
	`, req.UserID, req.PaymentProfileID)

	if err != nil {
		h.jsonError(w, "failed to verify user", http.StatusInternalServerError)
		return
	}

	h.logger.Info("user_verified", slog.Int64("user_id", req.UserID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User verified"})
}

func (h *AuthHandler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

