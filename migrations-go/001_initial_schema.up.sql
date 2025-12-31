-- Vehicle Auction Platform - PostgreSQL Schema
-- Designed for OCC (Optimistic Concurrency Control) bid processing

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Custom ENUM types
CREATE TYPE user_role AS ENUM ('buyer', 'seller', 'admin');
CREATE TYPE vehicle_status AS ENUM ('draft', 'pending', 'active', 'sold', 'archived');
CREATE TYPE auction_status AS ENUM ('scheduled', 'active', 'ended', 'cancelled');
CREATE TYPE bid_status AS ENUM ('accepted', 'rejected', 'outbid');
CREATE TYPE order_status AS ENUM ('pending_payment', 'paid', 'in_transit', 'delivered', 'cancelled', 'disputed');
CREATE TYPE fulfillment_status AS ENUM ('pending', 'pickup_scheduled', 'picked_up', 'in_transit', 'delivered');

-- =============================================================================
-- USERS
-- =============================================================================
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    clerk_user_id VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    role user_role NOT NULL DEFAULT 'buyer',
    
    -- Verification
    id_verified_at TIMESTAMPTZ,
    authorize_payment_profile_id VARCHAR(100),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_clerk_id ON users(clerk_user_id);

-- =============================================================================
-- VEHICLES
-- =============================================================================
CREATE TABLE vehicles (
    id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL REFERENCES users(id),
    
    -- Core details
    vin VARCHAR(17) UNIQUE NOT NULL,
    year SMALLINT NOT NULL,
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    trim VARCHAR(100),
    body_type VARCHAR(50),
    exterior_color VARCHAR(50),
    interior_color VARCHAR(50),
    
    -- Specs
    mileage INT,
    engine VARCHAR(100),
    transmission VARCHAR(50),
    drivetrain VARCHAR(20),
    fuel_type VARCHAR(30),
    
    -- Condition
    title_status VARCHAR(50) DEFAULT 'clean',
    condition_grade VARCHAR(10),
    description TEXT,
    
    -- Pricing
    starting_price NUMERIC(10, 2) NOT NULL,
    reserve_price NUMERIC(10, 2),
    buy_now_price NUMERIC(10, 2),
    
    -- Location
    location_city VARCHAR(100),
    location_state VARCHAR(50),
    location_zip VARCHAR(20),
    
    -- Status
    status vehicle_status NOT NULL DEFAULT 'draft',
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vehicles_seller ON vehicles(seller_id);
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_vehicles_make_model ON vehicles(make, model);
CREATE INDEX idx_vehicles_year ON vehicles(year);

-- =============================================================================
-- VEHICLE IMAGES
-- =============================================================================
CREATE TABLE vehicle_images (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    s3_key VARCHAR(500) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    display_order SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vehicle_images_vehicle ON vehicle_images(vehicle_id);

-- =============================================================================
-- AUCTIONS
-- =============================================================================
CREATE TABLE auctions (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT UNIQUE NOT NULL REFERENCES vehicles(id),
    status auction_status NOT NULL DEFAULT 'scheduled',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    
    -- Denormalized current state (fast reads, updated atomically with OCC)
    current_bid NUMERIC(10, 2) NOT NULL DEFAULT 0,
    current_bid_user_id BIGINT REFERENCES users(id),
    bid_count INT NOT NULL DEFAULT 0,
    
    -- OCC version for concurrent bid handling
    version INT NOT NULL DEFAULT 0,
    
    -- Anti-snipe protection
    extension_count SMALLINT NOT NULL DEFAULT 0,
    max_extensions SMALLINT NOT NULL DEFAULT 10,
    snipe_threshold_minutes SMALLINT NOT NULL DEFAULT 2,
    extension_minutes SMALLINT NOT NULL DEFAULT 2,
    
    -- Winner (set when auction ends)
    winner_id BIGINT REFERENCES users(id),
    winning_bid NUMERIC(10, 2),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_auctions_ends_at ON auctions(ends_at);
CREATE INDEX idx_auctions_vehicle ON auctions(vehicle_id);

-- =============================================================================
-- BIDS (Full history - never lose bid data)
-- =============================================================================
CREATE TABLE bids (
    id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT NOT NULL REFERENCES auctions(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    amount NUMERIC(10, 2) NOT NULL,
    
    -- Outcome tracking
    status bid_status NOT NULL,
    previous_high_bid NUMERIC(10, 2),  -- What was high bid when this was placed
    
    -- Auto-bid support (future)
    max_bid NUMERIC(10, 2),
    is_auto_bid BOOLEAN NOT NULL DEFAULT false,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bids_auction ON bids(auction_id);
CREATE INDEX idx_bids_user ON bids(user_id);
CREATE INDEX idx_bids_auction_amount ON bids(auction_id, amount DESC);

-- =============================================================================
-- ORDERS
-- =============================================================================
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT UNIQUE NOT NULL REFERENCES auctions(id),
    buyer_id BIGINT NOT NULL REFERENCES users(id),
    seller_id BIGINT NOT NULL REFERENCES users(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
    
    -- Pricing
    sale_price NUMERIC(10, 2) NOT NULL,
    buyer_premium NUMERIC(10, 2) NOT NULL DEFAULT 0,
    seller_fee NUMERIC(10, 2) NOT NULL DEFAULT 0,
    total_price NUMERIC(10, 2) NOT NULL,
    
    -- Status
    status order_status NOT NULL DEFAULT 'pending_payment',
    
    -- Payment
    payment_intent_id VARCHAR(255),
    paid_at TIMESTAMPTZ,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer ON orders(buyer_id);
CREATE INDEX idx_orders_seller ON orders(seller_id);
CREATE INDEX idx_orders_status ON orders(status);

-- =============================================================================
-- FULFILLMENT
-- =============================================================================
CREATE TABLE fulfillments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT UNIQUE NOT NULL REFERENCES orders(id),
    
    -- Status
    status fulfillment_status NOT NULL DEFAULT 'pending',
    
    -- Transport details
    carrier VARCHAR(100),
    tracking_number VARCHAR(255),
    estimated_pickup TIMESTAMPTZ,
    actual_pickup TIMESTAMPTZ,
    estimated_delivery TIMESTAMPTZ,
    actual_delivery TIMESTAMPTZ,
    
    -- Addresses
    pickup_address JSONB,
    delivery_address JSONB,
    
    -- Notes
    notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- WATCHLIST
-- =============================================================================
CREATE TABLE watchlist (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    auction_id BIGINT NOT NULL REFERENCES auctions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, auction_id)
);

CREATE INDEX idx_watchlist_user ON watchlist(user_id);

-- =============================================================================
-- NOTIFICATIONS
-- =============================================================================
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    data JSONB,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, read_at) WHERE read_at IS NULL;

-- =============================================================================
-- TRIGGERS
-- =============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_auctions_updated_at BEFORE UPDATE ON auctions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_fulfillments_updated_at BEFORE UPDATE ON fulfillments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

