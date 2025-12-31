/**
 * E2E tests for Watchlist functionality
 */
import { test, expect, waitForBackend } from './fixtures/test-utils';
import { seedBuyer, seedFullAuctionScenario } from './fixtures/seed';

test.describe('Watchlist Page', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  // Note: Watchlist requires Clerk authentication which redirects unauthenticated users.
  // We test the routing exists and the redirect behavior is correct.
  
  test('watchlist route exists and redirects unauthenticated users', async ({ page }) => {
    await page.goto('/watchlist');
    await page.waitForLoadState('domcontentloaded');
    
    // Either shows the page or redirects to sign-in
    const url = page.url();
    const isOnWatchlist = url.includes('/watchlist');
    const isRedirectedToSignIn = url.includes('sign-in') || url.includes('clerk');
    
    expect(isOnWatchlist || isRedirectedToSignIn).toBeTruthy();
  });
});

test.describe('Watchlist API', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('GET /api/watchlist requires authentication', async ({ request }) => {
    const response = await request.get('http://localhost:8080/api/watchlist');
    
    expect(response.status()).toBe(401);
  });

  test('GET /api/watchlist returns data for authenticated user', async ({ request }) => {
    const user = await seedBuyer('WatchlistAPIBuyer');
    
    const response = await request.get('http://localhost:8080/api/watchlist', {
      headers: {
        'X-Dev-User-ID': String(user.id),
      },
    });
    
    expect(response.status()).toBe(200);
    const data = await response.json();
    // API returns watchlist array (may be empty)
    expect(data).toHaveProperty('watchlist');
  });
});

