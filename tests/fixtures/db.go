package fixtures

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates a connection pool for testing
// Uses TEST_DATABASE_URL env var or falls back to a default
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/vehicle_auc_test?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Verify connection
	err = db.Ping(ctx)
	require.NoError(t, err, "Failed to ping test database")

	// Clean up on test completion
	t.Cleanup(func() {
		CleanupTestData(t, db)
		db.Close()
	})

	return db
}

// SetupTestDBWithMigrations sets up DB and ensures schema exists
func SetupTestDBWithMigrations(t *testing.T) *pgxpool.Pool {
	t.Helper()

	db := SetupTestDB(t)

	// Check if schema exists by checking for users table
	ctx := context.Background()
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'users'
		)
	`).Scan(&exists)
	require.NoError(t, err)

	if !exists {
		t.Skip("Database schema not initialized. Run migrations first: make migrate-test")
	}

	return db
}

