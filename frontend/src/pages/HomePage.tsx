/**
 * HomePage - Live auction grid with real-time updates
 */
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useActiveAuctions } from '@/hooks';
import { AuctionCard, AuctionCardSkeleton } from '@/components/AuctionCard';
import { Button } from '@/components/ui/button';
import { Gavel, Clock, Sparkles, AlertCircle } from 'lucide-react';

type FilterTab = 'all' | 'ending_soon' | 'newly_listed';

export default function HomePage() {
  const [activeTab, setActiveTab] = useState<FilterTab>('all');
  const { auctions, isLoading, error, refetch } = useActiveAuctions();

  // Filter auctions based on active tab
  const filteredAuctions = auctions.filter((auction) => {
    const now = Date.now();
    const endsAt = new Date(auction.ends_at).getTime();
    const createdAt = new Date(auction.created_at).getTime();
    const twentyFourHours = 24 * 60 * 60 * 1000;

    switch (activeTab) {
      case 'ending_soon':
        return endsAt - now < twentyFourHours && endsAt > now;
      case 'newly_listed':
        return now - createdAt < twentyFourHours;
      default:
        return true;
    }
  });

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero Section - Compact */}
      <section className="bg-gradient-to-r from-slate-900 to-slate-700 text-white py-12">
        <div className="container mx-auto px-4">
          <div className="flex flex-col md:flex-row items-center justify-between gap-6">
            <div>
              <h1 className="text-3xl md:text-4xl font-bold mb-2">
                Live Vehicle Auctions
              </h1>
              <p className="text-gray-300 text-lg">
                Bid on quality vehicles. Real-time updates. No dealer markup.
              </p>
            </div>
            <div className="flex gap-3">
              <Link to="/vehicles/new">
                <Button variant="secondary" size="lg">
                  Sell Your Car
                </Button>
              </Link>
              <Link to="/dashboard">
                <Button variant="outline" size="lg" className="text-white border-white hover:bg-white/10">
                  My Dashboard
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Filter Tabs */}
      <div className="bg-white border-b sticky top-0 z-40">
        <div className="container mx-auto px-4">
          <div className="flex gap-1 py-2">
            <TabButton
              active={activeTab === 'all'}
              onClick={() => setActiveTab('all')}
              icon={<Gavel size={16} />}
            >
              All Auctions
              <span className="ml-2 px-2 py-0.5 bg-gray-100 rounded-full text-xs">
                {auctions.length}
              </span>
            </TabButton>
            <TabButton
              active={activeTab === 'ending_soon'}
              onClick={() => setActiveTab('ending_soon')}
              icon={<Clock size={16} />}
            >
              Ending Soon
            </TabButton>
            <TabButton
              active={activeTab === 'newly_listed'}
              onClick={() => setActiveTab('newly_listed')}
              icon={<Sparkles size={16} />}
            >
              Newly Listed
            </TabButton>
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
              <h3 className="font-semibold text-red-800">Failed to load auctions</h3>
              <p className="text-red-600 text-sm">
                {error instanceof Error ? error.message : 'Something went wrong'}
              </p>
            </div>
            <Button variant="outline" onClick={() => refetch()} className="ml-auto">
              Retry
            </Button>
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
        {!isLoading && !error && filteredAuctions.length === 0 && (
          <div className="text-center py-16">
            <div className="w-20 h-20 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <Gavel className="text-gray-400" size={40} />
            </div>
            <h2 className="text-2xl font-semibold text-gray-700 mb-2">
              {activeTab === 'all' 
                ? 'No Active Auctions' 
                : activeTab === 'ending_soon'
                ? 'No Auctions Ending Soon'
                : 'No New Listings'
              }
            </h2>
            <p className="text-gray-500 mb-6 max-w-md mx-auto">
              {activeTab === 'all'
                ? 'Check back soon for new vehicles or list your own!'
                : 'Try viewing all auctions to see what\'s available.'
              }
            </p>
            {activeTab !== 'all' && (
              <Button onClick={() => setActiveTab('all')}>
                View All Auctions
              </Button>
            )}
          </div>
        )}

        {/* Auction Grid */}
        {!isLoading && filteredAuctions.length > 0 && (
          <>
            {activeTab === 'ending_soon' && filteredAuctions.length > 0 && (
              <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-6 flex items-center gap-3">
                <Clock className="text-amber-600" size={20} />
                <p className="text-amber-800">
                  <strong>{filteredAuctions.length}</strong> auction{filteredAuctions.length !== 1 ? 's' : ''} ending in the next 24 hours!
                </p>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredAuctions.map((auction) => (
                <AuctionCard key={auction.id} auction={auction} />
              ))}
            </div>
          </>
        )}
      </main>

      {/* Stats Footer */}
      <section className="bg-white border-t py-8">
        <div className="container mx-auto px-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6 text-center">
            <div>
              <p className="text-3xl font-bold text-primary">{auctions.length}</p>
              <p className="text-sm text-gray-500">Active Auctions</p>
            </div>
            <div>
              <p className="text-3xl font-bold text-primary">
                {auctions.reduce((sum, a) => sum + a.bid_count, 0)}
              </p>
              <p className="text-sm text-gray-500">Total Bids</p>
            </div>
            <div>
              <p className="text-3xl font-bold text-primary">
                ${(auctions.reduce((sum, a) => sum + a.current_bid, 0) / 1000).toFixed(0)}K+
              </p>
              <p className="text-sm text-gray-500">In Bids</p>
            </div>
            <div>
              <p className="text-3xl font-bold text-primary">24/7</p>
              <p className="text-sm text-gray-500">Live Support</p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}

interface TabButtonProps {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  children: React.ReactNode;
}

function TabButton({ active, onClick, icon, children }: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`
        flex items-center gap-2 px-4 py-2 rounded-lg font-medium text-sm transition-colors
        ${active 
          ? 'bg-primary text-primary-foreground' 
          : 'text-gray-600 hover:bg-gray-100'
        }
      `}
    >
      {icon}
      {children}
    </button>
  );
}
