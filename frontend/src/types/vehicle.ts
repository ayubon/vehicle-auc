/**
 * Vehicle domain types - single source of truth for vehicle data structures.
 */

export interface Vehicle {
  id: number;
  vin: string;
  year: number;
  make: string;
  model: string;
  trim?: string;
  mileage?: number;
  condition?: string;
  title_type?: string;
  starting_price?: number;
  buy_now_price?: number;
  location_city?: string;
  location_state?: string;
  primary_image_url?: string;
}

export interface VehicleDetail extends Vehicle {
  body_type?: string;
  engine?: string;
  transmission?: string;
  drivetrain?: string;
  exterior_color?: string;
  interior_color?: string;
  title_state?: string;
  has_keys?: boolean;
  description?: string;
  reserve_price?: number;
  location?: {
    address?: string;
    city?: string;
    state?: string;
    zip?: string;
  };
  images?: VehicleImage[];
  auction?: AuctionSummary;
}

export interface VehicleImage {
  url: string;
  is_primary: boolean;
}

export interface AuctionSummary {
  id: number;
  status: string;
  current_bid: number;
  bid_count: number;
  ends_at?: string;
  time_remaining: number;
}

export interface VehicleFilters {
  make: string;
  year_min: string;
  year_max: string;
  price_max: string;
}

export interface VehicleListResponse {
  vehicles: Vehicle[];
  total: number;
  pages: number;
  current_page: number;
}
