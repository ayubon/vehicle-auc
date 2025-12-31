/**
 * WatchlistPage - View user's watched auctions
 */
import { Link } from 'react-router-dom';
import { useUser, RedirectToSignIn } from '@clerk/clerk-react';
import { useWatchlist } from '@/hooks';
import { AuctionCard, AuctionCardSkeleton } from '@/components/AuctionCard';
import { Button } from '@/components/ui/button';
import { Heart, AlertCircle, ArrowLeft, Gavel } from 'lucide-react';

export default function WatchlistPage() {
  const { isLoaded, isSignedIn } = useUser();
  const { watchlist, isLoading, error } = useWatchlist();

  if (!isLoaded) {
    return <WatchlistPageSkeleton />;
  }

  if (!isSignedIn) {
    return <RedirectToSignIn />;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-6">
          <Link
            to="/dashboard"
            className="inline-flex items-center text-sm text-gray-500 hover:text-gray-800 mb-4"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Dashboard
          </Link>
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold flex items-center gap-2">
                <Heart className="h-8 w-8 text-red-500" />
                My Watchlist
              </h1>
              <p className="text-gray-500 mt-1">
                {watchlist.length} auction{watchlist.length !== 1 ? 's' : ''} you're watching
              </p>
            </div>
            <Link to="/">
              <Button>
                <Gavel className="h-4 w-4 mr-2" />
                Browse More
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {/* Error State */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-6 mb-8 flex items-center gap-4">
            <AlertCircle className="text-red-500" size={24} />
            <div>
              <h3 className="font-semibold text-red-800">Failed to load watchlist</h3>
              <p className="text-red-600 text-sm">
                {error instanceof Error ? error.message : 'Something went wrong'}
              </p>
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <AuctionCardSkeleton key={i} />
            ))}
          </div>
        )}

        {/* Empty State */}
        {!isLoading && !error && watchlist.length === 0 && (
          <div className="text-center py-16">
            <div className="w-20 h-20 bg-red-50 rounded-full flex items-center justify-center mx-auto mb-4">
              <Heart className="text-red-300" size={40} />
            </div>
            <h2 className="text-2xl font-semibold text-gray-700 mb-2">
              Your Watchlist is Empty
            </h2>
            <p className="text-gray-500 mb-6 max-w-md mx-auto">
              Start watching auctions to get notified about updates and never miss a great deal!
            </p>
            <Link to="/">
              <Button size="lg">
                <Gavel className="h-5 w-5 mr-2" />
                Browse Auctions
              </Button>
            </Link>
          </div>
        )}

        {/* Watchlist Grid */}
        {!isLoading && watchlist.length > 0 && (
          <>
            <div className="flex items-center justify-between mb-6">
              <p className="text-sm text-gray-500">
                Showing {watchlist.length} watched auction{watchlist.length !== 1 ? 's' : ''}
              </p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {watchlist.map((auction) => (
                <AuctionCard key={auction.id} auction={auction} />
              ))}
            </div>
          </>
        )}
      </main>
    </div>
  );
}

function WatchlistPageSkeleton() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-6">
          <div className="h-9 w-48 bg-gray-200 rounded animate-pulse mb-2" />
          <div className="h-5 w-64 bg-gray-200 rounded animate-pulse" />
        </div>
      </div>
      <main className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <AuctionCardSkeleton key={i} />
          ))}
        </div>
      </main>
    </div>
  );
}

