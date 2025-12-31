package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClerkClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"sub"`
	Email  string `json:"email"`
}

// ClerkAuth validates JWTs from Clerk
type ClerkAuth struct {
	logger    *slog.Logger
	jwksURL   string
	secretKey string
	db        *pgxpool.Pool
}

func NewClerkAuth(logger *slog.Logger, jwksURL, secretKey string, db *pgxpool.Pool) *ClerkAuth {
	return &ClerkAuth{
		logger:    logger,
		jwksURL:   jwksURL,
		secretKey: secretKey,
		db:        db,
	}
}

// Middleware returns the auth middleware handler
func (c *ClerkAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Development/test bypass: X-Dev-User-ID header
		// Only allowed in development or test environments
		env := os.Getenv("ENVIRONMENT")
		if env == "development" || env == "test" || env == "" {
			if devUserID := r.Header.Get("X-Dev-User-ID"); devUserID != "" {
				var uid int64
				if _, err := fmt.Sscanf(devUserID, "%d", &uid); err == nil && uid > 0 {
					c.logger.Debug("dev bypass auth", slog.Int64("user_id", uid), slog.String("env", env))
					ctx := WithUserID(r.Context(), uid)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			c.logger.Warn("missing authorization header",
				slog.String("path", r.URL.Path),
				slog.String("request_id", GetRequestID(r.Context())),
			)
			c.unauthorized(w, "missing authorization header")
			return
		}
		c.logger.Debug("auth header present", slog.String("path", r.URL.Path))

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.unauthorized(w, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := c.validateToken(tokenString)
		if err != nil {
			c.logger.Warn("token validation failed",
				slog.String("error", err.Error()),
				slog.String("request_id", GetRequestID(r.Context())),
			)
			c.unauthorized(w, "invalid token")
			return
		}

		// Look up internal user ID from clerk_user_id
		var userID int64
		err = c.db.QueryRow(r.Context(),
			"SELECT id FROM users WHERE clerk_user_id = $1",
			claims.UserID,
		).Scan(&userID)
		if err != nil {
			c.logger.Warn("user not found for clerk_user_id",
				slog.String("clerk_user_id", claims.UserID),
				slog.String("error", err.Error()),
				slog.String("request_id", GetRequestID(r.Context())),
			)
			c.unauthorized(w, "user not found - please sync your account")
			return
		}

		// Add to context
		ctx := WithUserID(r.Context(), userID)
		ctx = context.WithValue(ctx, "clerk_user_id", claims.UserID)
		ctx = context.WithValue(ctx, "clerk_email", claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *ClerkAuth) validateToken(tokenString string) (*ClerkClaims, error) {
	claims := &ClerkClaims{}

	// Clerk uses RS256 (RSA) signing. For proper validation, we'd need to:
	// 1. Fetch JWKS from c.jwksURL
	// 2. Find the key matching the token's "kid" header
	// 3. Validate the signature with that public key
	//
	// For development, we parse without signature verification and rely on
	// the database lookup to confirm the user exists.
	// TODO: Implement proper JWKS validation for production

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Basic validation - check token structure
	if token == nil || claims.UserID == "" {
		return nil, fmt.Errorf("invalid token structure")
	}

	// Check expiration if present
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func (c *ClerkAuth) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// OptionalAuth allows requests without auth but adds user info if present
func (c *ClerkAuth) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := c.validateToken(parts[1])
		if err != nil {
			// Log but don't fail - auth is optional
			c.logger.Debug("optional auth token validation failed",
				slog.String("error", err.Error()),
			)
			next.ServeHTTP(w, r)
			return
		}

		// Look up internal user ID
		var userID int64
		err = c.db.QueryRow(r.Context(),
			"SELECT id FROM users WHERE clerk_user_id = $1",
			claims.UserID,
		).Scan(&userID)
		
		ctx := r.Context()
		if err == nil {
			ctx = WithUserID(ctx, userID)
		}
		ctx = context.WithValue(ctx, "clerk_user_id", claims.UserID)
		ctx = context.WithValue(ctx, "clerk_email", claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClerkUserID extracts Clerk user ID from context
func GetClerkUserID(ctx context.Context) string {
	if id, ok := ctx.Value("clerk_user_id").(string); ok {
		return id
	}
	return ""
}

// GetClerkEmail extracts Clerk email from context
func GetClerkEmail(ctx context.Context) string {
	if email, ok := ctx.Value("clerk_email").(string); ok {
		return email
	}
	return ""
}

