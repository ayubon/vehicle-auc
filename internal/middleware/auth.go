package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
	// In production, you'd cache the JWKS
}

func NewClerkAuth(logger *slog.Logger, jwksURL, secretKey string) *ClerkAuth {
	return &ClerkAuth{
		logger:    logger,
		jwksURL:   jwksURL,
		secretKey: secretKey,
	}
}

// Middleware returns the auth middleware handler
func (c *ClerkAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			c.unauthorized(w, "missing authorization header")
			return
		}

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

		// Add claims to context
		ctx := context.WithValue(r.Context(), "clerk_user_id", claims.UserID)
		ctx = context.WithValue(ctx, "clerk_email", claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *ClerkAuth) validateToken(tokenString string) (*ClerkClaims, error) {
	// In development, we might skip validation
	// In production, validate against Clerk's JWKS

	claims := &ClerkClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
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

		ctx := context.WithValue(r.Context(), "clerk_user_id", claims.UserID)
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

