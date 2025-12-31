import { test, expect, waitForBackend } from './fixtures/test-utils';
import { seedSeller, seedActiveVehicle, seedAuction } from './fixtures/seed';

/**
 * API Contract Tests
 * 
 * These tests validate that the backend API responses match
 * the shape expected by the frontend. This catches mismatches
 * like the public_url vs url bug early.
 */

test.describe('Vehicle API Contracts', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('GET /api/vehicles returns expected shape', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/vehicles');
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    // Validate response shape
    expect(data).toHaveProperty('vehicles');
    expect(data).toHaveProperty('total');
    expect(data).toHaveProperty('limit');
    expect(data).toHaveProperty('offset');
    expect(Array.isArray(data.vehicles)).toBeTruthy();
  });

  test('GET /api/vehicles/:id returns expected shape', async ({ request }) => {
    // Create a vehicle first
    const seller = await seedSeller('ContractTestSeller');
    const vehicle = await seedActiveVehicle(seller.id, {
      make: 'ContractTest',
      model: 'Model1',
    });
    
    const response = await request.get(`http://localhost:8080/api/vehicles/${vehicle.id}`, {
      headers: { 'X-Dev-User-ID': '1' },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    // Response is wrapped in { vehicle: {...} }
    expect(data).toHaveProperty('vehicle');
    const vehicleData = data.vehicle;
    
    // Validate required fields that frontend expects
    expect(vehicleData).toHaveProperty('id');
    expect(vehicleData).toHaveProperty('vin');
    expect(vehicleData).toHaveProperty('year');
    expect(vehicleData).toHaveProperty('make');
    expect(vehicleData).toHaveProperty('model');
    expect(vehicleData).toHaveProperty('starting_price');
    expect(vehicleData).toHaveProperty('status');
  });

  test('POST /api/vehicles returns expected shape', async ({ request }) => {
    const response = await request.post('http://localhost:8080/api/vehicles', {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': '1',
      },
      data: {
        vin: `CONTRACT${Date.now()}`.slice(0, 17).padEnd(17, '0'),
        year: 2021,
        make: 'ContractTest',
        model: 'CreateTest',
        starting_price: 15000,
      },
    });
    
    expect(response.status()).toBe(201);
    const data = await response.json();
    
    // Frontend expects these fields
    expect(data).toHaveProperty('vehicle_id');
    expect(data).toHaveProperty('message');
    expect(typeof data.vehicle_id).toBe('number');
  });
});

test.describe('Image Upload API Contracts', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('POST /api/vehicles/:id/upload-url returns all required fields', async ({ request }) => {
    // Create a vehicle first
    const seller = await seedSeller('UploadContractSeller');
    const vehicle = await seedActiveVehicle(seller.id, {
      make: 'UploadTest',
    });
    
    const response = await request.post(`http://localhost:8080/api/vehicles/${vehicle.id}/upload-url`, {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': String(seller.id),
      },
      data: {
        filename: 'test.jpg',
        content_type: 'image/jpeg',
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    // Frontend expects these specific fields
    // This test would have caught the public_url bug!
    expect(data).toHaveProperty('upload_url');
    expect(data).toHaveProperty('s3_key');
    expect(data).toHaveProperty('url');
    expect(data).toHaveProperty('public_url'); // Critical: frontend uses this
    
    // Validate URL formats
    expect(data.upload_url).toMatch(/^https?:\/\//);
    expect(data.url).toMatch(/^https?:\/\//);
    expect(data.public_url).toMatch(/^https?:\/\//);
    expect(data.s3_key).toContain('vehicles/');
  });

  test('POST /api/vehicles/:id/images returns expected shape', async ({ request }) => {
    const seller = await seedSeller('ImageContractSeller');
    const vehicle = await seedActiveVehicle(seller.id, {
      make: 'ImageTest',
    });
    
    const response = await request.post(`http://localhost:8080/api/vehicles/${vehicle.id}/images`, {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': String(seller.id),
      },
      data: {
        s3_key: 'vehicles/test/image.jpg',
        url: 'https://example.com/image.jpg',
        is_primary: true,
      },
    });
    
    expect(response.status()).toBe(201);
    const data = await response.json();
    
    expect(data).toHaveProperty('message');
    expect(data).toHaveProperty('image_id');
    expect(data).toHaveProperty('is_primary');
  });
});

test.describe('Auth API Contracts', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('POST /api/auth/clerk-sync returns expected shape', async ({ request }) => {
    const response = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: {
        'Content-Type': 'application/json',
      },
      data: {
        clerk_user_id: `test_clerk_${Date.now()}`,
        email: `contract-test-${Date.now()}@test.com`,
        first_name: 'Contract',
        last_name: 'Test',
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    // Frontend expects user object with these fields
    expect(data).toHaveProperty('user');
    expect(data.user).toHaveProperty('id');
    expect(data.user).toHaveProperty('email');
    expect(data.user).toHaveProperty('role');
    expect(data.user).toHaveProperty('can_bid');
    expect(data.user).toHaveProperty('is_id_verified');
    expect(data.user).toHaveProperty('has_payment_method');
  });

  test('GET /api/auth/me returns expected shape', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/auth/me', {
      headers: {
        'X-Dev-User-ID': '1',
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    expect(data).toHaveProperty('id');
    expect(data).toHaveProperty('email');
    expect(data).toHaveProperty('role');
    expect(data).toHaveProperty('can_bid');
    expect(data).toHaveProperty('is_id_verified');
    expect(data).toHaveProperty('has_payment_method');
  });
});

test.describe('Auction API Contracts', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('GET /api/auctions returns expected shape', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/auctions');
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    expect(data).toHaveProperty('auctions');
    expect(data).toHaveProperty('total');
    expect(Array.isArray(data.auctions)).toBeTruthy();
  });

  test('POST /api/auctions/:id/bid returns expected shape', async ({ request }) => {
    // Create full auction scenario
    const seller = await seedSeller('BidContractSeller');
    const vehicle = await seedActiveVehicle(seller.id, {
      make: 'BidTest',
      starting_price: 10000,
    });
    
    // Create auction via API
    const createAuctionResp = await request.post('http://localhost:8080/api/auctions', {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': String(seller.id),
      },
      data: {
        vehicle_id: vehicle.id,
        duration_hours: 24,
      },
    });
    
    if (createAuctionResp.ok()) {
      const auction = await createAuctionResp.json();
      
      // Place a bid
      const bidResponse = await request.post(`http://localhost:8080/api/auctions/${auction.auction_id}/bid`, {
        headers: {
          'Content-Type': 'application/json',
          'X-Dev-User-ID': '1', // Different user bids
        },
        data: {
          amount: 11000,
        },
      });
      
      // Bid submission returns 202 Accepted (async processing)
      expect(bidResponse.status()).toBe(202);
      const bidData = await bidResponse.json();
      
      // Frontend expects ticket_id for tracking
      expect(bidData).toHaveProperty('ticket_id');
      expect(bidData).toHaveProperty('status');
    }
  });
});

test.describe('Health Check API', () => {
  test('GET /health returns expected shape', async ({ request }) => {
    const response = await request.get('http://localhost:8080/health');
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    expect(data).toHaveProperty('status');
    expect(data.status).toBe('healthy');
    expect(data).toHaveProperty('timestamp');
    expect(data).toHaveProperty('checks');
    expect(data.checks).toHaveProperty('database');
  });
});

test.describe('Error Response Contracts', () => {
  test('401 returns standard error shape', async () => {
    // Use native fetch to ensure no extra headers are added
    const response = await fetch('http://localhost:8080/api/auth/me');
    // No auth header = 401
    
    expect(response.status).toBe(401);
    const data = await response.json();
    
    expect(data).toHaveProperty('error');
    expect(typeof data.error).toBe('string');
  });

  test('400 returns standard error shape', async ({ request }) => {
    const response = await request.post('http://localhost:8080/api/vehicles', {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': '1',
      },
      data: {
        // Missing required fields
        vin: 'short',
      },
    });
    
    expect(response.status()).toBe(400);
    const data = await response.json();
    
    expect(data).toHaveProperty('error');
    expect(typeof data.error).toBe('string');
  });

  test('404 returns standard error shape', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/vehicles/999999999', {
      headers: { 'X-Dev-User-ID': '1' },
    });
    
    expect(response.status()).toBe(404);
    const data = await response.json();
    
    expect(data).toHaveProperty('error');
  });
});

