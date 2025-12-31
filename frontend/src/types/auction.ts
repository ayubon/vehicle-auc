/**
 * Auction domain types - single source of truth for auction data structures.
 */

export interface Auction {
  id: number;
  vehicle_id: number;
  status: 'scheduled' | 'active' | 'ended' | 'cancelled';
  starts_at: string;
  ends_at: string;
  current_bid: number;
  current_bid_user_id?: number;
  bid_count: number;
  version: number;
  winner_id?: number;
  winning_bid?: number;
  created_at: string;
  updated_at: string;
  // Joined vehicle data
  vehicle?: AuctionVehicle;
}

export interface AuctionVehicle {
  id: number;
  vin: string;
  year: number;
  make: string;
  model: string;
  trim?: string;
  mileage?: number;
  exterior_color?: string;
  starting_price: number;
  location_city?: string;
  location_state?: string;
  status: string;
  images?: AuctionVehicleImage[];
  seller_first_name?: string;
  seller_last_name?: string;
}

export interface AuctionVehicleImage {
  id: number;
  url: string;
  is_primary: boolean;
  display_order: number;
}

export interface Bid {
  id: number;
  auction_id: number;
  user_id: number;
  amount: number;
  status: 'accepted' | 'rejected' | 'outbid';
  previous_high_bid?: number;
  created_at: string;
  // User info if joined
  user_first_name?: string;
  user_last_name?: string;
}

export interface AuctionFilters {
  status?: 'active' | 'ended' | 'all';
  ending_soon?: boolean;
  make?: string;
  min_price?: number;
  max_price?: number;
}

export interface AuctionListResponse {
  auctions: Auction[];
  total: number;
  limit: number;
  offset: number;
}

export interface BidListResponse {
  bids: Bid[];
  total: number;
}

export interface PlaceBidRequest {
  amount: number;
}

export interface PlaceBidResponse {
  ticket_id: string;
  status: 'queued' | 'accepted' | 'rejected';
  message?: string;
}

// SSE Event types
export interface BidUpdateEvent {
  type: 'bid_update';
  auction_id: number;
  current_bid: number;
  bid_count: number;
  bidder_id: number;
  timestamp: string;
}

export interface AuctionEndEvent {
  type: 'auction_ended';
  auction_id: number;
  winner_id?: number;
  winning_bid?: number;
}

export type SSEEvent = BidUpdateEvent | AuctionEndEvent;

// User's bid summary
export interface UserBid {
  auction_id: number;
  auction: Auction;
  my_highest_bid: number;
  is_winning: boolean;
  total_bids: number;
}

export interface UserBidsResponse {
  bids: UserBid[];
  winning_count: number;
  total_count: number;
}

