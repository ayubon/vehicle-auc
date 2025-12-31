/**
 * AuctionCard - Display auction in a card format with real-time updates
 */
import { Link } from 'react-router-dom';
import { useAuctionSSE } from '@/hooks';
import { CountdownTimer } from './CountdownTimer';
import { WatchlistButton } from './WatchlistButton';
import { BidForm } from './BidForm';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { Auction } from '@/types';

interface AuctionCardProps {
  auction: Auction;
  showBidForm?: boolean;
  className?: string;
}

export function AuctionCard({
  auction,
  showBidForm = true,
  className = '',
}: AuctionCardProps) {
  // Subscribe to real-time updates
  const { currentBid, bidCount, isConnected } = useAuctionSSE(
    auction.id,
    auction.current_bid,
    auction.bid_count
  );

  const vehicle = auction.vehicle;
  const primaryImage = vehicle?.images?.find((img) => img.is_primary)?.url 
    || vehicle?.images?.[0]?.url
    || 'https://images.unsplash.com/photo-1494976388531-d1058494cdd8?w=400';

  const title = vehicle 
    ? `${vehicle.year} ${vehicle.make} ${vehicle.model}`
    : `Auction #${auction.id}`;

  return (
    <Card className={`overflow-hidden hover:shadow-lg transition-shadow ${className}`}>
      {/* Image */}
      <Link to={`/auctions/${auction.id}`} className="block relative">
        <div className="aspect-[16/10] overflow-hidden bg-gray-100">
          <img
            src={primaryImage}
            alt={title}
            className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
          />
        </div>
        
        {/* Status badge */}
        <Badge 
          className="absolute top-2 left-2"
          variant={auction.status === 'active' ? 'default' : 'secondary'}
        >
          {auction.status === 'active' ? 'Live' : auction.status}
        </Badge>

        {/* SSE connection indicator */}
        {isConnected && (
          <div className="absolute top-2 right-12 w-2 h-2 bg-green-500 rounded-full animate-pulse" 
               title="Live updates active" />
        )}

        {/* Watchlist button */}
        <div className="absolute top-2 right-2">
          <WatchlistButton auctionId={auction.id} size="sm" />
        </div>
      </Link>

      <CardContent className="p-4">
        {/* Title */}
        <Link to={`/auctions/${auction.id}`}>
          <h3 className="font-semibold text-lg hover:text-blue-600 transition-colors line-clamp-1">
            {title}
          </h3>
        </Link>

        {/* Vehicle details */}
        {vehicle && (
          <p className="text-sm text-gray-500 mt-1">
            {vehicle.mileage?.toLocaleString()} miles
            {vehicle.location_city && ` • ${vehicle.location_city}, ${vehicle.location_state}`}
          </p>
        )}

        {/* Bid info */}
        <div className="mt-4 flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500">Current Bid</p>
            <p className="text-2xl font-bold text-green-600">
              ${currentBid.toLocaleString()}
            </p>
            <p className="text-xs text-gray-400">{bidCount} bids</p>
          </div>
          
          <div className="text-right">
            <p className="text-sm text-gray-500">Ends In</p>
            <CountdownTimer endsAt={auction.ends_at} compact />
          </div>
        </div>

        {/* Quick bid form */}
        {showBidForm && auction.status === 'active' && (
          <div className="mt-4 pt-4 border-t">
            <BidForm
              auctionId={auction.id}
              currentBid={currentBid}
              compact
            />
          </div>
        )}

        {/* View details link */}
        <Link
          to={`/auctions/${auction.id}`}
          className="mt-4 block text-center text-sm text-blue-600 hover:text-blue-800"
        >
          View Details →
        </Link>
      </CardContent>
    </Card>
  );
}

/**
 * AuctionCardSkeleton - Loading placeholder for AuctionCard
 */
export function AuctionCardSkeleton() {
  return (
    <Card className="overflow-hidden">
      <div className="aspect-[16/10] bg-gray-200 animate-pulse" />
      <CardContent className="p-4">
        <div className="h-6 bg-gray-200 rounded animate-pulse mb-2" />
        <div className="h-4 bg-gray-200 rounded animate-pulse w-2/3 mb-4" />
        <div className="flex justify-between">
          <div className="space-y-2">
            <div className="h-4 bg-gray-200 rounded animate-pulse w-20" />
            <div className="h-8 bg-gray-200 rounded animate-pulse w-28" />
          </div>
          <div className="space-y-2">
            <div className="h-4 bg-gray-200 rounded animate-pulse w-16" />
            <div className="h-6 bg-gray-200 rounded animate-pulse w-24" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default AuctionCard;

