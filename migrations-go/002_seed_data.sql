-- Seed Data for Vehicle Auction Platform
-- Run with: psql $DATABASE_URL -f migrations-go/002_seed_data.sql

-- =============================================================================
-- USERS (5 test users)
-- =============================================================================
INSERT INTO users (id, clerk_user_id, email, first_name, last_name, phone, role, id_verified_at, created_at) VALUES
(1, 'clerk_seed_seller1', 'seller1@test.com', 'John', 'Dealer', '555-0101', 'seller', NOW(), NOW()),
(2, 'clerk_seed_seller2', 'seller2@test.com', 'Sarah', 'Motors', '555-0102', 'seller', NOW(), NOW()),
(3, 'clerk_seed_buyer1', 'buyer1@test.com', 'Mike', 'Thompson', '555-0201', 'buyer', NOW(), NOW()),
(4, 'clerk_seed_buyer2', 'buyer2@test.com', 'Emily', 'Chen', '555-0202', 'buyer', NOW(), NOW()),
(5, 'clerk_seed_buyer3', 'buyer3@test.com', 'David', 'Wilson', '555-0203', 'buyer', NULL, NOW())
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name;

-- Reset sequence
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

-- =============================================================================
-- VEHICLES (10 vehicles with realistic data)
-- =============================================================================
INSERT INTO vehicles (id, seller_id, vin, year, make, model, trim, body_type, exterior_color, interior_color, mileage, engine, transmission, drivetrain, fuel_type, title_status, condition_grade, description, starting_price, reserve_price, location_city, location_state, location_zip, status) VALUES
-- Seller 1 vehicles
(1, 1, 'JH4KA8260MC000001', 2021, 'Honda', 'Accord', 'Sport', 'Sedan', 'Crystal Black Pearl', 'Black', 28500, '1.5L Turbo I4', 'CVT', 'FWD', 'Gasoline', 'clean', 'A', 'One owner, always garaged. Full service history available. Non-smoker vehicle.', 22000.00, 20000.00, 'Los Angeles', 'CA', '90001', 'active'),
(2, 1, '1HGBH41JXMN000002', 2022, 'Toyota', 'Camry', 'XSE', 'Sedan', 'Wind Chill Pearl', 'Red', 15200, '2.5L I4', 'Automatic', 'FWD', 'Gasoline', 'clean', 'A+', 'Like new condition. All maintenance done at dealership. Extended warranty available.', 26000.00, 24000.00, 'San Francisco', 'CA', '94102', 'active'),
(3, 1, '5YFBURHE8LP000003', 2020, 'BMW', '3 Series', '330i', 'Sedan', 'Alpine White', 'Black', 42000, '2.0L Turbo I4', 'Automatic', 'RWD', 'Gasoline', 'clean', 'B+', 'Well maintained BMW. Minor wear on driver seat. All services up to date.', 32000.00, 30000.00, 'San Diego', 'CA', '92101', 'active'),
(4, 1, 'WBA8E9G50GN000004', 2023, 'Mercedes-Benz', 'C-Class', 'C300', 'Sedan', 'Obsidian Black', 'Macchiato Beige', 8500, '2.0L Turbo I4', 'Automatic', 'AWD', 'Gasoline', 'clean', 'A+', 'Nearly new Mercedes! Still under factory warranty. Premium package included.', 45000.00, 42000.00, 'Beverly Hills', 'CA', '90210', 'active'),
(5, 1, '1G1YY22G465000005', 2019, 'Chevrolet', 'Corvette', 'Stingray', 'Coupe', 'Torch Red', 'Jet Black', 18900, '6.2L V8', 'Manual', 'RWD', 'Gasoline', 'clean', 'A', 'Head turner! Performance exhaust, magnetic ride. Never tracked.', 55000.00, 52000.00, 'Orange County', 'CA', '92618', 'active'),

-- Seller 2 vehicles
(6, 2, '1FTEW1EP5KF000006', 2022, 'Ford', 'F-150', 'XLT', 'Truck', 'Agate Black', 'Medium Earth Gray', 31000, '2.7L EcoBoost V6', 'Automatic', '4WD', 'Gasoline', 'clean', 'A', 'Work truck used for light duty. Bedliner included. Tow package.', 38000.00, 35000.00, 'Phoenix', 'AZ', '85001', 'active'),
(7, 2, '5TDJZRFH9HS000007', 2021, 'Toyota', 'Highlander', 'Limited', 'SUV', 'Midnight Black', 'Glazed Caramel', 25600, '3.5L V6', 'Automatic', 'AWD', 'Gasoline', 'clean', 'A', 'Family SUV in excellent condition. 3rd row seating. Premium audio.', 42000.00, 40000.00, 'Scottsdale', 'AZ', '85251', 'active'),
(8, 2, '5J6RW2H85KL000008', 2020, 'Honda', 'CR-V', 'Touring', 'SUV', 'Modern Steel', 'Gray', 38500, '1.5L Turbo I4', 'CVT', 'AWD', 'Gasoline', 'clean', 'B+', 'Popular CR-V with all the features. Light hail damage on hood (cosmetic only).', 28000.00, 26000.00, 'Tucson', 'AZ', '85701', 'salvage', 'B'),
(9, 2, 'WVWZZZ3CZWE000009', 2023, 'Porsche', '911', 'Carrera', 'Coupe', 'Guards Red', 'Black', 3200, '3.0L Twin-Turbo H6', 'PDK', 'RWD', 'Gasoline', 'clean', 'A+', 'Barely driven 911! Sport Chrono package. Like showroom condition.', 125000.00, 120000.00, 'Las Vegas', 'NV', '89101', 'active'),
(10, 2, '1N4BL4BV4KC000010', 2022, 'Nissan', 'Altima', 'SV', 'Sedan', 'Gun Metallic', 'Charcoal', 22100, '2.5L I4', 'CVT', 'FWD', 'Gasoline', 'clean', 'A', 'Reliable daily driver. Great fuel economy. Safety Shield 360 included.', 21000.00, 19000.00, 'Henderson', 'NV', '89002', 'active')
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    starting_price = EXCLUDED.starting_price;

-- Reset sequence
SELECT setval('vehicles_id_seq', (SELECT MAX(id) FROM vehicles));

-- =============================================================================
-- VEHICLE IMAGES (sample image URLs - using placeholder images)
-- =============================================================================
INSERT INTO vehicle_images (vehicle_id, s3_key, url, is_primary, display_order) VALUES
(1, 'vehicles/1/main.jpg', 'https://images.unsplash.com/photo-1619767886558-efdc259cde1a?w=800', true, 0),
(1, 'vehicles/1/interior.jpg', 'https://images.unsplash.com/photo-1583267746897-2cf415887172?w=800', false, 1),
(2, 'vehicles/2/main.jpg', 'https://images.unsplash.com/photo-1621007947382-bb3c3994e3fb?w=800', true, 0),
(2, 'vehicles/2/side.jpg', 'https://images.unsplash.com/photo-1550355291-bbee04a92027?w=800', false, 1),
(3, 'vehicles/3/main.jpg', 'https://images.unsplash.com/photo-1555215695-3004980ad54e?w=800', true, 0),
(4, 'vehicles/4/main.jpg', 'https://images.unsplash.com/photo-1618843479313-40f8afb4b4d8?w=800', true, 0),
(5, 'vehicles/5/main.jpg', 'https://images.unsplash.com/photo-1552519507-da3b142c6e3d?w=800', true, 0),
(5, 'vehicles/5/engine.jpg', 'https://images.unsplash.com/photo-1544636331-e26879cd4d9b?w=800', false, 1),
(6, 'vehicles/6/main.jpg', 'https://images.unsplash.com/photo-1590362891991-f776e747a588?w=800', true, 0),
(7, 'vehicles/7/main.jpg', 'https://images.unsplash.com/photo-1619767886558-efdc259cde1a?w=800', true, 0),
(8, 'vehicles/8/main.jpg', 'https://images.unsplash.com/photo-1568844293986-8c1a5e1a5d5b?w=800', true, 0),
(9, 'vehicles/9/main.jpg', 'https://images.unsplash.com/photo-1503376780353-7e6692767b70?w=800', true, 0),
(9, 'vehicles/9/rear.jpg', 'https://images.unsplash.com/photo-1611821064430-0d40291d0f0b?w=800', false, 1),
(10, 'vehicles/10/main.jpg', 'https://images.unsplash.com/photo-1609521263047-f8f205293f24?w=800', true, 0)
ON CONFLICT DO NOTHING;

-- =============================================================================
-- AUCTIONS (6 active auctions with varied end times)
-- =============================================================================
INSERT INTO auctions (id, vehicle_id, status, starts_at, ends_at, current_bid, current_bid_user_id, bid_count, version) VALUES
-- Ending in 2 hours (urgent!)
(1, 1, 'active', NOW() - INTERVAL '5 days', NOW() + INTERVAL '2 hours', 24500.00, 3, 12, 12),
-- Ending in 6 hours
(2, 2, 'active', NOW() - INTERVAL '4 days', NOW() + INTERVAL '6 hours', 27000.00, 4, 8, 8),
-- Ending tomorrow
(3, 3, 'active', NOW() - INTERVAL '3 days', NOW() + INTERVAL '1 day', 33500.00, 3, 5, 5),
-- Ending in 2 days
(4, 5, 'active', NOW() - INTERVAL '2 days', NOW() + INTERVAL '2 days', 57000.00, 4, 4, 4),
-- Ending in 3 days (newly listed)
(5, 6, 'active', NOW() - INTERVAL '1 day', NOW() + INTERVAL '3 days', 39000.00, 3, 2, 2),
-- Ending in 5 days (fresh listing)
(6, 9, 'active', NOW() - INTERVAL '12 hours', NOW() + INTERVAL '5 days', 126000.00, 4, 1, 1)
ON CONFLICT (id) DO UPDATE SET
    current_bid = EXCLUDED.current_bid,
    bid_count = EXCLUDED.bid_count,
    ends_at = EXCLUDED.ends_at;

-- Reset sequence
SELECT setval('auctions_id_seq', (SELECT MAX(id) FROM auctions));

-- =============================================================================
-- BIDS (Sample bid history)
-- =============================================================================
-- Auction 1 bids (Honda Accord - hot auction)
INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, created_at) VALUES
(1, 3, 22500.00, 'outbid', 22000.00, NOW() - INTERVAL '4 days'),
(1, 4, 23000.00, 'outbid', 22500.00, NOW() - INTERVAL '4 days' + INTERVAL '2 hours'),
(1, 3, 23500.00, 'outbid', 23000.00, NOW() - INTERVAL '3 days'),
(1, 4, 24000.00, 'outbid', 23500.00, NOW() - INTERVAL '3 days' + INTERVAL '1 hour'),
(1, 3, 24500.00, 'accepted', 24000.00, NOW() - INTERVAL '2 days')
ON CONFLICT DO NOTHING;

-- Auction 2 bids (Toyota Camry)
INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, created_at) VALUES
(2, 4, 26500.00, 'outbid', 26000.00, NOW() - INTERVAL '3 days'),
(2, 3, 26750.00, 'outbid', 26500.00, NOW() - INTERVAL '2 days'),
(2, 4, 27000.00, 'accepted', 26750.00, NOW() - INTERVAL '1 day')
ON CONFLICT DO NOTHING;

-- Auction 3 bids (BMW 3 Series)
INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, created_at) VALUES
(3, 3, 32500.00, 'outbid', 32000.00, NOW() - INTERVAL '2 days'),
(3, 5, 33000.00, 'outbid', 32500.00, NOW() - INTERVAL '1 day'),
(3, 3, 33500.00, 'accepted', 33000.00, NOW() - INTERVAL '12 hours')
ON CONFLICT DO NOTHING;

-- Auction 4 bids (Corvette)
INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, created_at) VALUES
(4, 4, 56000.00, 'outbid', 55000.00, NOW() - INTERVAL '1 day'),
(4, 3, 56500.00, 'outbid', 56000.00, NOW() - INTERVAL '18 hours'),
(4, 4, 57000.00, 'accepted', 56500.00, NOW() - INTERVAL '6 hours')
ON CONFLICT DO NOTHING;

-- Auction 5 bids (F-150)
INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, created_at) VALUES
(5, 3, 38500.00, 'outbid', 38000.00, NOW() - INTERVAL '12 hours'),
(5, 3, 39000.00, 'accepted', 38500.00, NOW() - INTERVAL '6 hours')
ON CONFLICT DO NOTHING;

-- Auction 6 bids (Porsche 911)
INSERT INTO bids (auction_id, user_id, amount, status, previous_high_bid, created_at) VALUES
(6, 4, 126000.00, 'accepted', 125000.00, NOW() - INTERVAL '2 hours')
ON CONFLICT DO NOTHING;

-- =============================================================================
-- WATCHLIST (sample watchlist items)
-- =============================================================================
INSERT INTO watchlist (user_id, auction_id) VALUES
(3, 2),  -- Mike watching Camry
(3, 4),  -- Mike watching Corvette
(4, 1),  -- Emily watching Accord
(4, 3),  -- Emily watching BMW
(5, 6)   -- David watching Porsche
ON CONFLICT DO NOTHING;

-- =============================================================================
-- NOTIFICATIONS (sample notifications)
-- =============================================================================
INSERT INTO notifications (user_id, type, title, message, data, created_at) VALUES
(3, 'outbid', 'You''ve been outbid!', 'Someone placed a higher bid on the 2022 Toyota Camry', '{"auction_id": 2, "new_bid": 27000}', NOW() - INTERVAL '1 day'),
(4, 'outbid', 'You''ve been outbid!', 'Someone placed a higher bid on the 2021 Honda Accord', '{"auction_id": 1, "new_bid": 24500}', NOW() - INTERVAL '2 days'),
(3, 'auction_ending', 'Auction ending soon!', 'The 2021 Honda Accord auction ends in 2 hours', '{"auction_id": 1}', NOW() - INTERVAL '1 hour'),
(4, 'bid_accepted', 'Bid placed successfully', 'Your bid of $27,000 on the 2022 Toyota Camry was accepted', '{"auction_id": 2, "amount": 27000}', NOW() - INTERVAL '1 day')
ON CONFLICT DO NOTHING;

-- Print summary
DO $$
BEGIN
    RAISE NOTICE 'Seed data loaded:';
    RAISE NOTICE '  - % users', (SELECT COUNT(*) FROM users);
    RAISE NOTICE '  - % vehicles', (SELECT COUNT(*) FROM vehicles);
    RAISE NOTICE '  - % vehicle images', (SELECT COUNT(*) FROM vehicle_images);
    RAISE NOTICE '  - % auctions', (SELECT COUNT(*) FROM auctions);
    RAISE NOTICE '  - % bids', (SELECT COUNT(*) FROM bids);
    RAISE NOTICE '  - % watchlist items', (SELECT COUNT(*) FROM watchlist);
    RAISE NOTICE '  - % notifications', (SELECT COUNT(*) FROM notifications);
END $$;

