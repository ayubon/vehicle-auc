import { test, expect, waitForBackend } from './fixtures/test-utils';

/**
 * Auth Flow Tests
 * 
 * Tests for Clerk authentication sync with backend.
 * These tests validate the auth flow using the dev bypass header
 * since we can't automate actual Clerk sign-in in E2E tests.
 */

test.describe('Auth Sync Flow', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('clerk-sync creates new user', async ({ request }) => {
    const uniqueId = Date.now();
    const response = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: {
        'Content-Type': 'application/json',
      },
      data: {
        clerk_user_id: `clerk_test_${uniqueId}`,
        email: `newuser-${uniqueId}@test.com`,
        first_name: 'New',
        last_name: 'User',
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    // Should return user data
    expect(data.user).toBeDefined();
    expect(data.user.email).toBe(`newuser-${uniqueId}@test.com`);
    expect(data.user.first_name).toBe('New');
    expect(data.user.role).toBe('buyer');
    expect(data.user.can_bid).toBe(false); // New users can't bid yet
  });

  test('clerk-sync updates existing user', async ({ request }) => {
    const uniqueId = Date.now();
    const email = `update-test-${uniqueId}@test.com`;
    
    // First sync - create user
    const createResponse = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        clerk_user_id: `clerk_update_${uniqueId}`,
        email,
        first_name: 'Original',
        last_name: 'Name',
      },
    });
    expect(createResponse.ok()).toBeTruthy();
    
    // Second sync - update user with new name
    const updateResponse = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        clerk_user_id: `clerk_update_${uniqueId}`,
        email,
        first_name: 'Updated',
        last_name: 'Person',
      },
    });
    
    expect(updateResponse.ok()).toBeTruthy();
    const data = await updateResponse.json();
    
    // Name should be updated
    expect(data.user.first_name).toBe('Updated');
    expect(data.user.last_name).toBe('Person');
  });

  test('clerk-sync rejects missing email', async ({ request }) => {
    const response = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        clerk_user_id: `clerk_invalid_${Date.now()}`,
        // Missing email
        first_name: 'Test',
      },
    });
    
    expect(response.status()).toBe(400);
  });

  test('clerk-sync rejects missing clerk_user_id', async ({ request }) => {
    const response = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        // Missing clerk_user_id
        email: 'test@test.com',
        first_name: 'Test',
      },
    });
    
    expect(response.status()).toBe(400);
  });
});

test.describe('Auth Me Endpoint', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('GET /auth/me returns current user with dev bypass', async ({ request }) => {
    // First ensure user exists
    await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        clerk_user_id: 'clerk_me_test',
        email: 'me-test@test.com',
        first_name: 'Me',
        last_name: 'Test',
      },
    });
    
    const response = await request.get('http://localhost:8080/api/auth/me', {
      headers: {
        'X-Dev-User-ID': '1', // Use dev bypass
      },
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    expect(data.id).toBeDefined();
    expect(data.email).toBeDefined();
    expect(data.role).toBeDefined();
  });

  test('GET /auth/me returns 401 without auth', async () => {
    // Use native fetch to ensure no extra headers are added
    const response = await fetch('http://localhost:8080/api/auth/me');
    // No auth header
    
    expect(response.status).toBe(401);
  });
});

test.describe('Profile Update', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('can update user profile', async ({ request }) => {
    // Create user first
    const uniqueId = Date.now();
    const syncResponse = await request.post('http://localhost:8080/api/auth/clerk-sync', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        clerk_user_id: `clerk_profile_${uniqueId}`,
        email: `profile-${uniqueId}@test.com`,
        first_name: 'Profile',
        last_name: 'Test',
      },
    });
    const userData = await syncResponse.json();
    const userId = userData.user.id;
    
    // Update profile via PUT /auth/me
    const updateResponse = await request.put('http://localhost:8080/api/auth/me', {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': String(userId),
      },
      data: {
        first_name: 'UpdatedFirst',
        last_name: 'UpdatedLast',
        phone: '555-1234',
      },
    });
    
    expect(updateResponse.ok()).toBeTruthy();
    
    // Verify update via GET /auth/me
    const meResponse = await request.get('http://localhost:8080/api/auth/me', {
      headers: { 'X-Dev-User-ID': String(userId) },
    });
    const meData = await meResponse.json();
    
    expect(meData.first_name).toBe('UpdatedFirst');
    expect(meData.last_name).toBe('UpdatedLast');
  });

  test('profile update requires auth', async () => {
    // Use native fetch to ensure no extra headers are added
    const response = await fetch('http://localhost:8080/api/auth/me', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        first_name: 'Test',
      }),
    });
    
    expect(response.status).toBe(401);
  });
});

test.describe('Dev Auth Bypass', () => {
  test('X-Dev-User-ID header bypasses auth in development', async ({ request }) => {
    // This verifies the dev bypass is working
    const response = await request.get('http://localhost:8080/api/auth/me', {
      headers: {
        'X-Dev-User-ID': '1',
      },
    });
    
    // Should succeed with dev bypass
    expect(response.ok()).toBeTruthy();
  });

  test('invalid X-Dev-User-ID is rejected', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/auth/me', {
      headers: {
        'X-Dev-User-ID': 'invalid', // Not a number
      },
    });
    
    // Should fall back to normal auth (and fail since no token)
    expect(response.status()).toBe(401);
  });

  test('X-Dev-User-ID with zero is rejected', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/auth/me', {
      headers: {
        'X-Dev-User-ID': '0',
      },
    });
    
    // Zero is not a valid user ID
    expect(response.status()).toBe(401);
  });
});

