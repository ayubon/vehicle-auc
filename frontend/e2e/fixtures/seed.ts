import { api, generateTestVehicle } from './test-utils';

/**
 * Test data seeding utilities for E2E tests.
 * These functions create test data via the backend API.
 */

export interface SeededUser {
  id: number;
  email: string;
}

export interface SeededVehicle {
  id: number;
  vin: string;
  make: string;
  model: string;
  year: number;
}

export interface SeededAuction {
  id: number;
  vehicleId: number;
}

/**
 * Create a seller user for testing
 */
export async function seedSeller(name = 'Test Seller'): Promise<SeededUser> {
  const email = `seller-${Date.now()}@test.com`;
  const result = await api.createUser(email, name, 'User');
  return {
    id: result.id,
    email,
  };
}

/**
 * Create a buyer user for testing
 */
export async function seedBuyer(name = 'Test Buyer'): Promise<SeededUser> {
  const email = `buyer-${Date.now()}@test.com`;
  const result = await api.createUser(email, name, 'User');
  return {
    id: result.id,
    email,
  };
}

/**
 * Create a vehicle with default test data
 */
export async function seedVehicle(
  sellerId: number,
  overrides?: Partial<{
    vin: string;
    year: number;
    make: string;
    model: string;
    starting_price: number;
  }>
): Promise<SeededVehicle> {
  const data = generateTestVehicle(overrides);
  const result = await api.createVehicle(data, sellerId);
  
  return {
    id: result.id,
    vin: data.vin,
    make: data.make,
    model: data.model,
    year: data.year,
  };
}

/**
 * Create and submit a vehicle (makes it active/visible)
 */
export async function seedActiveVehicle(
  sellerId: number,
  overrides?: Partial<{
    vin: string;
    year: number;
    make: string;
    model: string;
    starting_price: number;
  }>
): Promise<SeededVehicle> {
  const vehicle = await seedVehicle(sellerId, overrides);
  await api.submitVehicle(vehicle.id, sellerId);
  return vehicle;
}

/**
 * Create an auction for a vehicle
 */
export async function seedAuction(
  vehicleId: number,
  sellerId: number,
  durationHours = 24
): Promise<SeededAuction> {
  const result = await api.createAuction({
    vehicle_id: vehicleId,
    duration_hours: durationHours,
  }, sellerId);
  
  return {
    id: result.id,
    vehicleId,
  };
}

/**
 * Create a complete test scenario: seller + vehicle + auction
 */
export async function seedFullAuctionScenario(): Promise<{
  seller: SeededUser;
  vehicle: SeededVehicle;
  auction: SeededAuction;
}> {
  const seller = await seedSeller();
  const vehicle = await seedActiveVehicle(seller.id);
  const auction = await seedAuction(vehicle.id, seller.id);
  
  return { seller, vehicle, auction };
}

/**
 * Create multiple vehicles for listing tests
 */
export async function seedMultipleVehicles(
  sellerId: number,
  count: number,
  makes = ['Honda', 'Toyota', 'Ford', 'Chevrolet', 'BMW']
): Promise<SeededVehicle[]> {
  const vehicles: SeededVehicle[] = [];
  
  for (let i = 0; i < count; i++) {
    const make = makes[i % makes.length];
    const vehicle = await seedActiveVehicle(sellerId, {
      make,
      model: `Model${i + 1}`,
      year: 2020 + (i % 5),
      starting_price: 10000 + (i * 5000),
    });
    vehicles.push(vehicle);
  }
  
  return vehicles;
}

