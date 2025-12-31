import { test, expect, api, generateTestVehicle, waitForBackend } from './fixtures/test-utils';
import { seedSeller } from './fixtures/seed';

test.describe('Vehicle Creation Flow', () => {
  test.beforeAll(async () => {
    // Ensure backend is ready
    await waitForBackend();
  });

  test('can navigate to vehicle creation page', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/vehicles');
    
    // Click "List a Vehicle" button
    await authenticatedPage.click('text=List a Vehicle');
    
    // Should be on create page
    await expect(authenticatedPage).toHaveURL('/vehicles/new');
    await expect(authenticatedPage.locator('text=List a Vehicle')).toBeVisible();
  });

  test('shows validation errors for empty form', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/vehicles/new');
    
    // Try to submit without filling required fields
    await authenticatedPage.click('button[type="submit"]');
    
    // Should show validation errors (form shouldn't submit)
    await expect(authenticatedPage).toHaveURL('/vehicles/new');
  });

  test('can fill out vehicle form with valid data', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/vehicles/new');
    
    // Fill VIN
    await authenticatedPage.fill('input[name="vin"]', 'JH4KA8260MC000001');
    
    // Fill year
    await authenticatedPage.fill('input[name="year"]', '2021');
    
    // Fill make
    await authenticatedPage.fill('input[name="make"]', 'Honda');
    
    // Fill model
    await authenticatedPage.fill('input[name="model"]', 'Accord');
    
    // Fill starting price
    await authenticatedPage.fill('input[name="starting_price"]', '15000');
    
    // All fields should have values
    await expect(authenticatedPage.locator('input[name="vin"]')).toHaveValue('JH4KA8260MC000001');
    await expect(authenticatedPage.locator('input[name="year"]')).toHaveValue('2021');
    await expect(authenticatedPage.locator('input[name="make"]')).toHaveValue('Honda');
    await expect(authenticatedPage.locator('input[name="model"]')).toHaveValue('Accord');
    await expect(authenticatedPage.locator('input[name="starting_price"]')).toHaveValue('15000');
  });

  test('seller can create vehicle and see photos step', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/vehicles/new');
    
    // Fill required fields
    const testVin = `TEST${Date.now()}`.slice(0, 17).padEnd(17, '0');
    await authenticatedPage.fill('input[name="vin"]', testVin);
    await authenticatedPage.fill('input[name="year"]', '2021');
    await authenticatedPage.fill('input[name="make"]', 'Honda');
    await authenticatedPage.fill('input[name="model"]', 'Accord');
    await authenticatedPage.fill('input[name="starting_price"]', '15000');
    
    // Submit form
    await authenticatedPage.click('button:has-text("Create Vehicle")');
    
    // Wait for photos step to appear - use more specific locator
    await expect(authenticatedPage.getByRole('heading', { name: /Photos|Add Photos/i })).toBeVisible({ timeout: 10000 });
    
    // Should show image upload area
    await expect(authenticatedPage.locator('text=Drag & drop images here')).toBeVisible();
  });

  test('can see submit button after photos step', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/vehicles/new');
    
    // Fill and submit form
    const testVin = `TEST${Date.now()}`.slice(0, 17).padEnd(17, '0');
    await authenticatedPage.fill('input[name="vin"]', testVin);
    await authenticatedPage.fill('input[name="year"]', '2021');
    await authenticatedPage.fill('input[name="make"]', 'Toyota');
    await authenticatedPage.fill('input[name="model"]', 'Camry');
    await authenticatedPage.fill('input[name="starting_price"]', '20000');
    
    await authenticatedPage.click('button:has-text("Create Vehicle")');
    
    // Wait for photos step - use more specific locator
    await expect(authenticatedPage.getByRole('heading', { name: /Photos|Add Photos/i })).toBeVisible({ timeout: 10000 });
    
    // Should see Submit for Review button (may be disabled without images)
    await expect(authenticatedPage.locator('button:has-text("Submit for Review")')).toBeVisible();
  });
});

test.describe('Vehicle Creation API Integration', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('API creates vehicle with correct data', async ({ request }) => {
    const vehicleData = generateTestVehicle({
      make: 'APITest',
      model: 'TestModel',
    });
    
    const response = await request.post('http://localhost:8080/api/vehicles', {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': '1',
      },
      data: vehicleData,
    });
    
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.vehicle_id).toBeDefined();
    expect(data.message).toContain('created');
  });

  test('API returns error for invalid VIN', async ({ request }) => {
    const response = await request.post('http://localhost:8080/api/vehicles', {
      headers: {
        'Content-Type': 'application/json',
        'X-Dev-User-ID': '1',
      },
      data: {
        vin: 'short', // Invalid VIN - should be 17 chars
        year: 2021,
        make: 'Test',
        model: 'Model',
        starting_price: 15000,
      },
    });
    
    // Should return validation error
    expect(response.status()).toBe(400);
  });

  test('API rejects vehicle creation without auth', async ({ request }) => {
    const vehicleData = generateTestVehicle();
    
    // Make direct fetch call to avoid Playwright adding any default headers
    const response = await fetch('http://localhost:8080/api/vehicles', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // No X-Dev-User-ID header
      },
      body: JSON.stringify(vehicleData),
    });
    
    expect(response.status).toBe(401);
  });
});

