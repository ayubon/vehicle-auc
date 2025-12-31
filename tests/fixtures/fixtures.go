package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// TestUser creates a basic test user
func TestUser(t *testing.T, db *pgxpool.Pool) int64 {
	t.Helper()
	ctx := context.Background()

	email := fmt.Sprintf("testuser-%s@example.com", uuid.New().String()[:8])
	clerkID := fmt.Sprintf("clerk_%s", uuid.New().String()[:8])

	var userID int64
	err := db.QueryRow(ctx, `
		INSERT INTO users (clerk_user_id, email, first_name, last_name, role)
		VALUES ($1, $2, 'Test', 'User', 'buyer')
		RETURNING id
	`, clerkID, email).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// VerifiedUser creates a user who can place bids
func VerifiedUser(t *testing.T, db *pgxpool.Pool) int64 {
	t.Helper()
	ctx := context.Background()

	userID := TestUser(t, db)

	_, err := db.Exec(ctx, `
		UPDATE users SET
			id_verified_at = NOW(),
			authorize_payment_profile_id = $1
		WHERE id = $2
	`, fmt.Sprintf("profile_%s", uuid.New().String()[:8]), userID)
	require.NoError(t, err)

	return userID
}

// SellerUser creates a user with seller role
func SellerUser(t *testing.T, db *pgxpool.Pool) int64 {
	t.Helper()
	ctx := context.Background()

	email := fmt.Sprintf("seller-%s@example.com", uuid.New().String()[:8])
	clerkID := fmt.Sprintf("clerk_%s", uuid.New().String()[:8])

	var userID int64
	err := db.QueryRow(ctx, `
		INSERT INTO users (clerk_user_id, email, first_name, last_name, role, id_verified_at, authorize_payment_profile_id)
		VALUES ($1, $2, 'Test', 'Seller', 'seller', NOW(), 'seller-profile')
		RETURNING id
	`, clerkID, email).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// TestVehicle creates a test vehicle
func TestVehicle(t *testing.T, db *pgxpool.Pool, sellerID int64) int64 {
	t.Helper()
	ctx := context.Background()

	vin := fmt.Sprintf("1HGBH41JX%s", uuid.New().String()[:8])

	var vehicleID int64
	err := db.QueryRow(ctx, `
		INSERT INTO vehicles (
			seller_id, vin, year, make, model, trim, mileage,
			starting_price, status, location_city, location_state
		) VALUES (
			$1, $2, 2021, 'Honda', 'Accord', 'Sport', 35000,
			15000.00, 'active', 'Los Angeles', 'CA'
		)
		RETURNING id
	`, sellerID, vin).Scan(&vehicleID)
	require.NoError(t, err)

	return vehicleID
}

// TestVehicleWithDetails creates a vehicle with custom details
func TestVehicleWithDetails(t *testing.T, db *pgxpool.Pool, sellerID int64, year int, make, model string, startingPrice float64) int64 {
	t.Helper()
	ctx := context.Background()

	vin := fmt.Sprintf("1HGBH41JX%s", uuid.New().String()[:8])

	var vehicleID int64
	err := db.QueryRow(ctx, `
		INSERT INTO vehicles (
			seller_id, vin, year, make, model, starting_price, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, 'active'
		)
		RETURNING id
	`, sellerID, vin, year, make, model, startingPrice).Scan(&vehicleID)
	require.NoError(t, err)

	return vehicleID
}

// TestAuction creates an active auction
func TestAuction(t *testing.T, db *pgxpool.Pool, vehicleID int64) int64 {
	t.Helper()
	ctx := context.Background()

	startsAt := time.Now().Add(-1 * time.Hour)
	endsAt := time.Now().Add(23 * time.Hour)

	var auctionID int64
	err := db.QueryRow(ctx, `
		INSERT INTO auctions (
			vehicle_id, status, starts_at, ends_at,
			current_bid, bid_count, version
		) VALUES (
			$1, 'active', $2, $3, 0, 0, 0
		)
		RETURNING id
	`, vehicleID, startsAt, endsAt).Scan(&auctionID)
	require.NoError(t, err)

	return auctionID
}

// TestAuctionEndingSoon creates an auction ending within snipe threshold
func TestAuctionEndingSoon(t *testing.T, db *pgxpool.Pool, vehicleID int64) int64 {
	t.Helper()
	ctx := context.Background()

	startsAt := time.Now().Add(-1 * time.Hour)
	endsAt := time.Now().Add(1 * time.Minute) // Ending soon!

	var auctionID int64
	err := db.QueryRow(ctx, `
		INSERT INTO auctions (
			vehicle_id, status, starts_at, ends_at,
			current_bid, bid_count, version,
			snipe_threshold_minutes, extension_minutes
		) VALUES (
			$1, 'active', $2, $3, 0, 0, 0, 2, 2
		)
		RETURNING id
	`, vehicleID, startsAt, endsAt).Scan(&auctionID)
	require.NoError(t, err)

	return auctionID
}

// TestAuctionWithBid creates an auction with an existing bid
func TestAuctionWithBid(t *testing.T, db *pgxpool.Pool, vehicleID int64, currentBid float64, bidderID int64) int64 {
	t.Helper()
	ctx := context.Background()

	startsAt := time.Now().Add(-1 * time.Hour)
	endsAt := time.Now().Add(23 * time.Hour)

	var auctionID int64
	err := db.QueryRow(ctx, `
		INSERT INTO auctions (
			vehicle_id, status, starts_at, ends_at,
			current_bid, current_bid_user_id, bid_count, version
		) VALUES (
			$1, 'active', $2, $3, $4, $5, 1, 1
		)
		RETURNING id
	`, vehicleID, startsAt, endsAt, currentBid, bidderID).Scan(&auctionID)
	require.NoError(t, err)

	// Record the bid
	_, err = db.Exec(ctx, `
		INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid)
		VALUES ($1, $2, $3, 'accepted', 0)
	`, auctionID, bidderID, currentBid)
	require.NoError(t, err)

	return auctionID
}

// TestBid records a bid for an auction
func TestBid(t *testing.T, db *pgxpool.Pool, auctionID, userID int64, amount decimal.Decimal, status string) int64 {
	t.Helper()
	ctx := context.Background()

	var bidID int64
	err := db.QueryRow(ctx, `
		INSERT INTO bids (auction_id, user_id, amount, status)
		VALUES ($1, $2, $3, $4::bid_status)
		RETURNING id
	`, auctionID, userID, amount, status).Scan(&bidID)
	require.NoError(t, err)

	return bidID
}

// BuyerUser creates a verified buyer user
func BuyerUser(t *testing.T, db *pgxpool.Pool) int64 {
	t.Helper()
	return VerifiedUser(t, db)
}

// CreateUser creates a user with specific email and name
func CreateUser(t *testing.T, db *pgxpool.Pool, email, firstName, lastName string) int64 {
	t.Helper()
	ctx := context.Background()

	clerkID := fmt.Sprintf("clerk_%s", uuid.New().String()[:8])

	var userID int64
	err := db.QueryRow(ctx, `
		INSERT INTO users (clerk_user_id, email, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, 'buyer')
		RETURNING id
	`, clerkID, email, firstName, lastName).Scan(&userID)
	require.NoError(t, err)

	return userID
}

// CleanupTestData removes all test data (call in cleanup)
func CleanupTestData(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Delete in reverse order of dependencies
	tables := []string{
		"notifications",
		"watchlist",
		"fulfillments",
		"orders",
		"bids",
		"auctions",
		"vehicle_images",
		"vehicles",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}
}

