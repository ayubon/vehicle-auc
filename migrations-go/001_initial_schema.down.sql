-- Rollback: Drop all tables and types

DROP TRIGGER IF EXISTS update_fulfillments_updated_at ON fulfillments;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_auctions_updated_at ON auctions;
DROP TRIGGER IF EXISTS update_vehicles_updated_at ON vehicles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS watchlist;
DROP TABLE IF EXISTS fulfillments;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS bids;
DROP TABLE IF EXISTS auctions;
DROP TABLE IF EXISTS vehicle_images;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS fulfillment_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS bid_status;
DROP TYPE IF EXISTS auction_status;
DROP TYPE IF EXISTS vehicle_status;
DROP TYPE IF EXISTS user_role;

