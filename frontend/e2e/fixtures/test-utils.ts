import { Page, BrowserContext, test as baseTest, expect } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api';

/**
 * Test fixture that provides an authenticated page with X-Dev-User-ID header
 * injected into all API requests.
 */
export const test = baseTest.extend<{
  authenticatedPage: Page;
  testUserId: number;
}>({
  testUserId: 1, // Default test user ID
  
  authenticatedPage: async ({ page, testUserId }, use) => {
    // Intercept all API requests and add the dev auth header
    await page.route('**/api/**', async (route) => {
      const headers = {
        ...route.request().headers(),
        'X-Dev-User-ID': String(testUserId),
      };
      await route.continue({ headers });
    });
    
    await use(page);
  },
});

export { expect };

/**
 * API client for test data seeding and verification
 */
export class TestAPI {
  private baseUrl: string;
  private defaultUserId: number;
  
  constructor(baseUrl = API_BASE, defaultUserId = 1) {
    this.baseUrl = baseUrl;
    this.defaultUserId = defaultUserId;
  }
  
  private async request(
    method: string,
    path: string,
    data?: unknown,
    userId?: number
  ): Promise<Response> {
    const url = `${this.baseUrl}${path}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-Dev-User-ID': String(userId ?? this.defaultUserId),
    };
    
    const options: RequestInit = {
      method,
      headers,
    };
    
    if (data) {
      options.body = JSON.stringify(data);
    }
    
    return fetch(url, options);
  }
  
  // User operations
  async createUser(email: string, firstName: string, lastName: string): Promise<{ id: number }> {
    const response = await this.request('POST', '/auth/clerk-sync', {
      clerk_user_id: `test_${Date.now()}`,
      email,
      first_name: firstName,
      last_name: lastName,
    });
    const data = await response.json();
    return { id: data.user?.id };
  }
  
  // Vehicle operations
  async createVehicle(data: {
    vin: string;
    year: number;
    make: string;
    model: string;
    starting_price: number;
    trim?: string;
    mileage?: number;
  }, userId?: number): Promise<{ id: number }> {
    const response = await this.request('POST', '/vehicles', data, userId);
    const result = await response.json();
    return { id: result.vehicle_id };
  }
  
  async submitVehicle(vehicleId: number, userId?: number): Promise<void> {
    await this.request('POST', `/vehicles/${vehicleId}/submit`, null, userId);
  }
  
  async getVehicle(vehicleId: number): Promise<unknown> {
    const response = await this.request('GET', `/vehicles/${vehicleId}`);
    return response.json();
  }
  
  async listVehicles(params?: Record<string, string>): Promise<{ vehicles: unknown[]; total: number }> {
    const queryString = params ? '?' + new URLSearchParams(params).toString() : '';
    const response = await this.request('GET', `/vehicles${queryString}`);
    return response.json();
  }
  
  // Auction operations
  async createAuction(data: {
    vehicle_id: number;
    duration_hours?: number;
  }, userId?: number): Promise<{ id: number }> {
    // Calculate starts_at and ends_at from duration
    const now = new Date();
    const startsAt = now.toISOString();
    const endsAt = new Date(now.getTime() + (data.duration_hours || 24) * 60 * 60 * 1000).toISOString();
    
    const response = await this.request('POST', '/auctions', {
      vehicle_id: data.vehicle_id,
      starts_at: startsAt,
      ends_at: endsAt,
    }, userId);
    const result = await response.json();
    return { id: result.auction_id };
  }
  
  async getAuction(auctionId: number): Promise<unknown> {
    const response = await this.request('GET', `/auctions/${auctionId}`);
    return response.json();
  }
  
  async placeBid(auctionId: number, amount: number, userId?: number): Promise<unknown> {
    const response = await this.request('POST', `/auctions/${auctionId}/bid`, { amount }, userId);
    return response.json();
  }
  
  // Image operations
  async getUploadUrl(vehicleId: number, filename: string, contentType: string, userId?: number): Promise<{
    upload_url: string;
    s3_key: string;
    url: string;
    public_url: string;
  }> {
    const response = await this.request('POST', `/vehicles/${vehicleId}/upload-url`, {
      filename,
      content_type: contentType,
    }, userId);
    return response.json();
  }
  
  async addImage(vehicleId: number, s3Key: string, url: string, isPrimary: boolean, userId?: number): Promise<void> {
    await this.request('POST', `/vehicles/${vehicleId}/images`, {
      s3_key: s3Key,
      url,
      is_primary: isPrimary,
    }, userId);
  }
  
  // Health check
  async healthCheck(): Promise<{ status: string }> {
    const response = await fetch(`${this.baseUrl.replace('/api', '')}/health`);
    return response.json();
  }
}

/**
 * Global test API instance
 */
export const api = new TestAPI();

/**
 * Helper to generate unique test data
 */
export function generateTestVehicle(overrides?: Partial<{
  vin: string;
  year: number;
  make: string;
  model: string;
  starting_price: number;
}>) {
  const timestamp = Date.now();
  return {
    vin: `TEST${timestamp}`.slice(0, 17).padEnd(17, '0'),
    year: 2021,
    make: 'TestMake',
    model: 'TestModel',
    starting_price: 15000,
    ...overrides,
  };
}

/**
 * Wait for backend to be ready
 */
export async function waitForBackend(maxAttempts = 30, delayMs = 1000): Promise<void> {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      const health = await api.healthCheck();
      if (health.status === 'healthy') {
        return;
      }
    } catch {
      // Backend not ready yet
    }
    await new Promise(resolve => setTimeout(resolve, delayMs));
  }
  throw new Error('Backend did not become ready in time');
}

