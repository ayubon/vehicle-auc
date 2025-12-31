/**
 * AuctionDetailPage - Full auction view with bidding
 */
import { useParams, Link } from 'react-router-dom';
import { useAuction, useAuctionBids, useAuctionSSE, useAuth } from '@/hooks';
import { CountdownTimer } from '@/components/CountdownTimer';
import { WatchlistButton } from '@/components/WatchlistButton';
import { BidForm } from '@/components/BidForm';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { ArrowLeft, Calendar, MapPin, Gauge, AlertCircle, CheckCircle, Users } from 'lucide-react';

export default function AuctionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const auctionId = parseInt(id || '0');
  const { user } = useAuth();
  
  const { auction, vehicle, isLoading, error } = useAuction(auctionId);
  const { bids } = useAuctionBids(auctionId);
  const { currentBid, bidCount, isConnected } = useAuctionSSE(
    auctionId,
    auction?.current_bid || 0,
    auction?.bid_count || 0
  );

  if (isLoading) {
    return <AuctionDetailSkeleton />;
  }

  if (error || !auction) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <AlertCircle className="mx-auto h-12 w-12 text-red-500 mb-4" />
        <h1 className="text-2xl font-bold mb-2">Auction Not Found</h1>
        <p className="text-gray-500 mb-6">
          This auction may have ended or doesn't exist.
        </p>
        <Link to="/">
          <Button>Back to Auctions</Button>
        </Link>
      </div>
    );
  }

  const isActive = auction.status === 'active';
  const isWinning = auction.current_bid_user_id === user?.id;
  const primaryImage = vehicle?.images?.find(img => img.is_primary)?.url
    || vehicle?.images?.[0]?.url
    || 'https://images.unsplash.com/photo-1494976388531-d1058494cdd8?w=800';
  const images = vehicle?.images || [];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Breadcrumb */}
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-3">
          <Link
            to="/"
            className="inline-flex items-center text-sm text-gray-500 hover:text-gray-800"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Auctions
          </Link>
        </div>
      </div>

      <main className="container mx-auto px-4 py-8">
        <div className="grid lg:grid-cols-3 gap-8">
          {/* Left Column - Images & Details */}
          <div className="lg:col-span-2 space-y-6">
            {/* Main Image */}
            <Card className="overflow-hidden">
              <div className="aspect-[16/10] relative">
                <img
                  src={primaryImage}
                  alt={`${vehicle?.year} ${vehicle?.make} ${vehicle?.model}`}
                  className="w-full h-full object-cover"
                />
                
                {/* SSE indicator */}
                {isConnected && (
                  <div className="absolute top-4 left-4 flex items-center gap-2 bg-black/50 text-white px-3 py-1 rounded-full text-sm">
                    <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
                    Live
                  </div>
                )}

                {/* Status badge */}
                <Badge
                  className="absolute top-4 right-4"
                  variant={isActive ? 'default' : 'secondary'}
                >
                  {auction.status}
                </Badge>
              </div>

              {/* Thumbnail gallery */}
              {images.length > 1 && (
                <div className="p-4 flex gap-2 overflow-x-auto">
                  {images.map((img, i) => (
                    <img
                      key={img.id || i}
                      src={img.url}
                      alt={`View ${i + 1}`}
                      className="w-20 h-20 object-cover rounded cursor-pointer hover:opacity-80 transition-opacity"
                    />
                  ))}
                </div>
              )}
            </Card>

            {/* Vehicle Details */}
            <Card>
              <CardHeader>
                <CardTitle>Vehicle Details</CardTitle>
              </CardHeader>
              <CardContent className="grid sm:grid-cols-2 gap-4">
                <DetailRow label="VIN" value={vehicle?.vin} />
                <DetailRow label="Year" value={vehicle?.year} />
                <DetailRow label="Make" value={vehicle?.make} />
                <DetailRow label="Model" value={vehicle?.model} />
                <DetailRow label="Trim" value={vehicle?.trim || 'N/A'} />
                <DetailRow 
                  label="Mileage" 
                  value={vehicle?.mileage ? `${vehicle.mileage.toLocaleString()} miles` : 'N/A'} 
                  icon={<Gauge className="h-4 w-4" />}
                />
                <DetailRow 
                  label="Location" 
                  value={vehicle?.location_city && vehicle?.location_state 
                    ? `${vehicle.location_city}, ${vehicle.location_state}` 
                    : 'N/A'
                  }
                  icon={<MapPin className="h-4 w-4" />}
                />
                <DetailRow 
                  label="Listed" 
                  value={new Date(auction.created_at).toLocaleDateString()}
                  icon={<Calendar className="h-4 w-4" />}
                />
              </CardContent>
            </Card>

            {/* Description */}
            {vehicle?.description && (
              <Card>
                <CardHeader>
                  <CardTitle>Description</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-gray-700 whitespace-pre-wrap">{vehicle.description}</p>
                </CardContent>
              </Card>
            )}

            {/* Bid History */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-5 w-5" />
                  Bid History
                </CardTitle>
                <span className="text-sm text-gray-500">{bidCount} bids</span>
              </CardHeader>
              <CardContent>
                {bids.length === 0 ? (
                  <p className="text-gray-500 text-center py-4">No bids yet. Be the first!</p>
                ) : (
                  <div className="space-y-2 max-h-64 overflow-y-auto">
                    {bids.slice(0, 10).map((bid, i) => (
                      <div
                        key={bid.id || i}
                        className={`flex justify-between items-center p-3 rounded-lg ${
                          i === 0 ? 'bg-green-50 border border-green-200' : 'bg-gray-50'
                        }`}
                      >
                        <div className="flex items-center gap-2">
                          {i === 0 && <CheckCircle className="h-4 w-4 text-green-600" />}
                          <span className="font-medium">
                            ${bid.amount.toLocaleString()}
                          </span>
                          {bid.user_id === user?.id && (
                            <Badge variant="outline" className="text-xs">You</Badge>
                          )}
                        </div>
                        <span className="text-sm text-gray-500">
                          {new Date(bid.created_at).toLocaleTimeString()}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Right Column - Bidding Panel */}
          <div className="space-y-6">
            {/* Sticky bid panel */}
            <div className="lg:sticky lg:top-4 space-y-6">
              {/* Auction Summary */}
              <Card>
                <CardContent className="pt-6">
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h1 className="text-2xl font-bold">
                        {vehicle?.year} {vehicle?.make} {vehicle?.model}
                      </h1>
                      {vehicle?.trim && (
                        <p className="text-gray-500">{vehicle.trim}</p>
                      )}
                    </div>
                    <WatchlistButton auctionId={auctionId} showText />
                  </div>

                  <Separator className="my-4" />

                  {/* Time remaining */}
                  <div className="text-center mb-6">
                    <p className="text-sm text-gray-500 mb-1">
                      {isActive ? 'Time Remaining' : 'Auction Ended'}
                    </p>
                    <CountdownTimer endsAt={auction.ends_at} />
                  </div>

                  {/* Current bid */}
                  <div className="text-center mb-6">
                    <p className="text-sm text-gray-500">Current Bid</p>
                    <p className="text-4xl font-bold text-green-600">
                      ${currentBid.toLocaleString()}
                    </p>
                    <p className="text-sm text-gray-500">{bidCount} bids</p>
                  </div>

                  {/* Winning indicator */}
                  {isWinning && (
                    <div className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4 text-center">
                      <CheckCircle className="inline h-5 w-5 text-green-600 mr-2" />
                      <span className="text-green-800 font-medium">You're winning!</span>
                    </div>
                  )}

                  {/* Starting price */}
                  <div className="flex justify-between text-sm text-gray-500">
                    <span>Starting Price</span>
                    <span>${vehicle?.starting_price?.toLocaleString()}</span>
                  </div>
                </CardContent>
              </Card>

              {/* Bid Form */}
              {isActive && (
                <Card>
                  <CardHeader>
                    <CardTitle>Place Your Bid</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {user ? (
                      <BidForm
                        auctionId={auctionId}
                        currentBid={currentBid}
                        onSuccess={() => {
                          // Could show a success toast or update local state
                        }}
                      />
                    ) : (
                      <div className="text-center py-4">
                        <p className="text-gray-500 mb-4">Sign in to place a bid</p>
                        <Link to="/sign-in">
                          <Button className="w-full">Sign In</Button>
                        </Link>
                      </div>
                    )}
                  </CardContent>
                </Card>
              )}

              {/* Auction ended */}
              {!isActive && (
                <Card className="bg-gray-50">
                  <CardContent className="py-8 text-center">
                    <h3 className="text-xl font-semibold mb-2">Auction Ended</h3>
                    <p className="text-gray-500">
                      Final bid: ${currentBid.toLocaleString()}
                    </p>
                  </CardContent>
                </Card>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

interface DetailRowProps {
  label: string;
  value: string | number | undefined;
  icon?: React.ReactNode;
}

function DetailRow({ label, value, icon }: DetailRowProps) {
  return (
    <div className="flex items-center gap-2">
      {icon && <span className="text-gray-400">{icon}</span>}
      <div>
        <p className="text-sm text-gray-500">{label}</p>
        <p className="font-medium">{value}</p>
      </div>
    </div>
  );
}

function AuctionDetailSkeleton() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-3">
          <div className="h-5 w-32 bg-gray-200 rounded animate-pulse" />
        </div>
      </div>
      <main className="container mx-auto px-4 py-8">
        <div className="grid lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            <Card>
              <div className="aspect-[16/10] bg-gray-200 animate-pulse" />
            </Card>
            <Card>
              <CardContent className="pt-6">
                <div className="h-8 w-48 bg-gray-200 rounded animate-pulse mb-4" />
                <div className="grid sm:grid-cols-2 gap-4">
                  {[...Array(8)].map((_, i) => (
                    <div key={i} className="h-12 bg-gray-100 rounded animate-pulse" />
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
          <div>
            <Card>
              <CardContent className="pt-6 space-y-4">
                <div className="h-8 bg-gray-200 rounded animate-pulse" />
                <div className="h-16 bg-gray-200 rounded animate-pulse" />
                <div className="h-32 bg-gray-200 rounded animate-pulse" />
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
}
