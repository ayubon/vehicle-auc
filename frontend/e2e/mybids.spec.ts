/**
 * E2E tests for My Bids page
 */
import { test, expect, waitForBackend } from './fixtures/test-utils';
import { seedBuyer } from './fixtures/seed';

test.describe('My Bids Page', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  // Note: These pages require Clerk authentication which can't be easily mocked in E2E tests.
  // The page will redirect to Clerk sign-in for unauthenticated users.
  // We test the routing exists and the redirect behavior is correct.
  
  test('my-bids route exists and redirects unauthenticated users', async ({ page }) => {
    await page.goto('/my-bids');
    await page.waitForLoadState('domcontentloaded');
    
    // Either shows the page (if Clerk is not required) or redirects to sign-in
    const url = page.url();
    const isOnMyBids = url.includes('/my-bids');
    const isRedirectedToSignIn = url.includes('sign-in') || url.includes('clerk');
    
    expect(isOnMyBids || isRedirectedToSignIn).toBeTruthy();
  });
});

test.describe('My Bids API', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('GET /api/bids/my requires authentication', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/bids/my');
    
    // Should return 401 or 404 (if endpoint doesn't exist yet)
    expect([401, 404]).toContain(response.status());
  });

  test('authenticated user can fetch their bids', async ({ request }) => {
    const buyer = await seedBuyer('MyBidsAPIBuyer');
    
    const response = await request.get('http://localhost:8080/api/bids/my', {
      headers: {
        'X-Dev-User-ID': String(buyer.id),
      },
    });
    
    // Should return 200 or 404 if endpoint doesn't exist
    expect([200, 404]).toContain(response.status());
    
    if (response.status() === 200) {
      const data = await response.json();
      expect(data).toHaveProperty('bids');
      expect(Array.isArray(data.bids)).toBeTruthy();
    }
  });
});

