/**
 * DashboardPage - User dashboard with real data
 */
import { Link } from 'react-router-dom';
import { useUser, RedirectToSignIn } from '@clerk/clerk-react';
import { useAuth, useMyBids, useWatchlist, useNotifications } from '@/hooks';
import { AuctionCard, AuctionCardSkeleton } from '@/components/AuctionCard';
import { NotificationBell } from '@/components/NotificationBell';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { 
  Shield, 
  CreditCard, 
  Gavel, 
  Heart, 
  Car, 
  Bell,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  AlertTriangle
} from 'lucide-react';

export default function DashboardPage() {
  const { isLoaded, isSignedIn, user: clerkUser } = useUser();
  const { user: backendUser, isLoading: isLoadingUser } = useAuth();
  const { activeBids, wonBids, isLoading: isLoadingBids } = useMyBids();
  const { watchlist, isLoading: isLoadingWatchlist } = useWatchlist();
  const { unreadCount } = useNotifications();

  if (!isLoaded) {
    return <DashboardSkeleton />;
  }

  if (!isSignedIn) {
    return <RedirectToSignIn />;
  }

  const isVerified = backendUser?.is_id_verified;
  const hasPayment = backendUser?.has_payment_method;
  const canBid = isVerified && hasPayment;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold">
                Welcome back, {clerkUser?.firstName || 'User'}
              </h1>
              <p className="text-gray-500">
                Manage your auctions, bids, and account settings
              </p>
            </div>
            <NotificationBell />
          </div>
        </div>
      </div>

      <main className="container mx-auto px-4 py-8">
        {/* Status Cards */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {/* Active Bids */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
                  <Gavel className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{activeBids?.length || 0}</p>
                  <p className="text-sm text-gray-500">Active Bids</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Watchlist */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center">
                  <Heart className="h-6 w-6 text-red-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{watchlist?.length || 0}</p>
                  <p className="text-sm text-gray-500">Watching</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Auctions Won */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-full bg-green-100 flex items-center justify-center">
                  <TrendingUp className="h-6 w-6 text-green-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{wonBids?.length || 0}</p>
                  <p className="text-sm text-gray-500">Auctions Won</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Notifications */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-full bg-yellow-100 flex items-center justify-center">
                  <Bell className="h-6 w-6 text-yellow-600" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{unreadCount}</p>
                  <p className="text-sm text-gray-500">Unread</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid lg:grid-cols-3 gap-8">
          {/* Left Column - Account & Verification */}
          <div className="space-y-6">
            {/* Bidding Eligibility */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Gavel className="h-5 w-5" />
                  Bidding Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                {canBid ? (
                  <div className="flex items-center gap-2 text-green-600 mb-4">
                    <CheckCircle className="h-5 w-5" />
                    <span className="font-medium">Ready to Bid</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-2 text-yellow-600 mb-4">
                    <AlertTriangle className="h-5 w-5" />
                    <span className="font-medium">Complete setup to bid</span>
                  </div>
                )}
                
                {/* ID Verification */}
                <div className="flex items-center justify-between py-3 border-t">
                  <div className="flex items-center gap-2">
                    <Shield className={`h-5 w-5 ${isVerified ? 'text-green-500' : 'text-gray-400'}`} />
                    <span>ID Verification</span>
                  </div>
                  {isVerified ? (
                    <Badge variant="outline" className="text-green-600 border-green-600">
                      <CheckCircle className="h-3 w-3 mr-1" /> Verified
                    </Badge>
                  ) : (
                    <Button size="sm" variant="outline">Verify</Button>
                  )}
                </div>

                {/* Payment Method */}
                <div className="flex items-center justify-between py-3 border-t">
                  <div className="flex items-center gap-2">
                    <CreditCard className={`h-5 w-5 ${hasPayment ? 'text-green-500' : 'text-gray-400'}`} />
                    <span>Payment Method</span>
                  </div>
                  {hasPayment ? (
                    <Badge variant="outline" className="text-green-600 border-green-600">
                      <CheckCircle className="h-3 w-3 mr-1" /> Added
                    </Badge>
                  ) : (
                    <Button size="sm" variant="outline">Add Card</Button>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Account Info */}
            <Card>
              <CardHeader>
                <CardTitle>Account Info</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-500">Email</span>
                  <span className="font-medium">{clerkUser?.emailAddresses[0]?.emailAddress}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Name</span>
                  <span className="font-medium">{clerkUser?.fullName || 'Not set'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Role</span>
                  <Badge variant="outline">{backendUser?.role || 'buyer'}</Badge>
                </div>
              </CardContent>
            </Card>

            {/* Quick Actions */}
            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <Link to="/vehicles/new" className="block">
                  <Button variant="outline" className="w-full justify-start">
                    <Car className="h-4 w-4 mr-2" />
                    List a Vehicle
                  </Button>
                </Link>
                <Link to="/" className="block">
                  <Button variant="outline" className="w-full justify-start">
                    <Gavel className="h-4 w-4 mr-2" />
                    Browse Auctions
                  </Button>
                </Link>
                <Link to="/watchlist" className="block">
                  <Button variant="outline" className="w-full justify-start">
                    <Heart className="h-4 w-4 mr-2" />
                    My Watchlist
                  </Button>
                </Link>
              </CardContent>
            </Card>
          </div>

          {/* Right Column - Active Bids & Watchlist */}
          <div className="lg:col-span-2 space-y-6">
            {/* Active Bids */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  Your Active Bids
                </CardTitle>
                <Link to="/my-bids">
                  <Button variant="ghost" size="sm">View All →</Button>
                </Link>
              </CardHeader>
              <CardContent>
                {isLoadingBids ? (
                  <div className="grid md:grid-cols-2 gap-4">
                    <AuctionCardSkeleton />
                    <AuctionCardSkeleton />
                  </div>
                ) : activeBids && activeBids.length > 0 ? (
                  <div className="grid md:grid-cols-2 gap-4">
                    {activeBids.slice(0, 4).map((bid) => (
                      <AuctionCard 
                        key={bid.auction?.id} 
                        auction={bid.auction!} 
                        showBidForm={false}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    <Gavel className="h-12 w-12 mx-auto mb-3 text-gray-300" />
                    <p>No active bids</p>
                    <Link to="/" className="text-blue-600 hover:underline text-sm">
                      Browse auctions →
                    </Link>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Watchlist Preview */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Heart className="h-5 w-5" />
                  Watchlist
                </CardTitle>
                <Link to="/watchlist">
                  <Button variant="ghost" size="sm">View All →</Button>
                </Link>
              </CardHeader>
              <CardContent>
                {isLoadingWatchlist ? (
                  <div className="grid md:grid-cols-2 gap-4">
                    <AuctionCardSkeleton />
                    <AuctionCardSkeleton />
                  </div>
                ) : watchlist && watchlist.length > 0 ? (
                  <div className="grid md:grid-cols-2 gap-4">
                    {watchlist.slice(0, 4).map((auction) => (
                      <AuctionCard 
                        key={auction.id} 
                        auction={auction}
                        showBidForm={false}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    <Heart className="h-12 w-12 mx-auto mb-3 text-gray-300" />
                    <p>No items in watchlist</p>
                    <Link to="/" className="text-blue-600 hover:underline text-sm">
                      Browse auctions →
                    </Link>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-6">
          <Skeleton className="h-9 w-64 mb-2" />
          <Skeleton className="h-5 w-48" />
        </div>
      </div>
      <main className="container mx-auto px-4 py-8">
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <Skeleton className="w-12 h-12 rounded-full" />
                  <div>
                    <Skeleton className="h-8 w-12 mb-1" />
                    <Skeleton className="h-4 w-20" />
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </main>
    </div>
  );
}
