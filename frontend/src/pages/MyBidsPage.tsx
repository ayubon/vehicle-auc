/**
 * MyBidsPage - View user's bidding history
 */
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useUser, RedirectToSignIn } from '@clerk/clerk-react';
import { useMyBids } from '@/hooks';
import { AuctionCard, AuctionCardSkeleton } from '@/components/AuctionCard';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Gavel, 
  Trophy, 
  AlertCircle, 
  ArrowLeft,
  Clock,
  CheckCircle,
  XCircle
} from 'lucide-react';

type BidTab = 'active' | 'won' | 'outbid' | 'all';

export default function MyBidsPage() {
  const { isLoaded, isSignedIn } = useUser();
  const [activeTab, setActiveTab] = useState<BidTab>('active');
  const { activeBids, wonBids, outbidBids, allBids, isLoading, error } = useMyBids();

  if (!isLoaded) {
    return <MyBidsPageSkeleton />;
  }

  if (!isSignedIn) {
    return <RedirectToSignIn />;
  }

  const getDisplayBids = () => {
    switch (activeTab) {
      case 'active':
        return activeBids;
      case 'won':
        return wonBids;
      case 'outbid':
        return outbidBids;
      default:
        return allBids;
    }
  };

  const displayBids = getDisplayBids();

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
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Gavel className="h-8 w-8" />
            My Bids
          </h1>
          <p className="text-gray-500 mt-1">
            Track your bidding activity
          </p>
        </div>
      </div>

      {/* Stats */}
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-4">
          <div className="flex gap-6 overflow-x-auto">
            <StatCard
              icon={<Clock className="h-5 w-5 text-blue-600" />}
              label="Active"
              count={activeBids.length}
              active={activeTab === 'active'}
              onClick={() => setActiveTab('active')}
            />
            <StatCard
              icon={<Trophy className="h-5 w-5 text-green-600" />}
              label="Won"
              count={wonBids.length}
              active={activeTab === 'won'}
              onClick={() => setActiveTab('won')}
            />
            <StatCard
              icon={<XCircle className="h-5 w-5 text-red-600" />}
              label="Outbid"
              count={outbidBids.length}
              active={activeTab === 'outbid'}
              onClick={() => setActiveTab('outbid')}
            />
            <StatCard
              icon={<Gavel className="h-5 w-5 text-gray-600" />}
              label="All"
              count={allBids.length}
              active={activeTab === 'all'}
              onClick={() => setActiveTab('all')}
            />
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
              <h3 className="font-semibold text-red-800">Failed to load bids</h3>
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
        {!isLoading && !error && displayBids.length === 0 && (
          <div className="text-center py-16">
            <div className="w-20 h-20 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              {activeTab === 'won' ? (
                <Trophy className="text-gray-400" size={40} />
              ) : activeTab === 'outbid' ? (
                <XCircle className="text-gray-400" size={40} />
              ) : (
                <Gavel className="text-gray-400" size={40} />
              )}
            </div>
            <h2 className="text-2xl font-semibold text-gray-700 mb-2">
              {activeTab === 'active' && 'No Active Bids'}
              {activeTab === 'won' && 'No Auctions Won Yet'}
              {activeTab === 'outbid' && 'No Outbid Auctions'}
              {activeTab === 'all' && 'No Bid History'}
            </h2>
            <p className="text-gray-500 mb-6 max-w-md mx-auto">
              {activeTab === 'active' 
                ? "You haven't placed any active bids. Browse auctions to find your next vehicle!"
                : activeTab === 'won'
                ? "Keep bidding! Your winning auction will show up here."
                : activeTab === 'outbid'
                ? "Great news - none of your bids have been outbid!"
                : "Start bidding on auctions to build your history."
              }
            </p>
            <Link to="/">
              <Button>Browse Auctions</Button>
            </Link>
          </div>
        )}

        {/* Bids Grid */}
        {!isLoading && displayBids.length > 0 && (
          <div className="space-y-6">
            {displayBids.map((bid) => (
              <BidCard key={bid.id} bid={bid} />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
}

function StatCard({ icon, label, count, active, onClick }: StatCardProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-3 px-4 py-2 rounded-lg transition-colors ${
        active 
          ? 'bg-primary text-primary-foreground' 
          : 'hover:bg-gray-100'
      }`}
    >
      {icon}
      <div className="text-left">
        <p className="text-2xl font-bold">{count}</p>
        <p className="text-sm">{label}</p>
      </div>
    </button>
  );
}

interface BidCardProps {
  bid: {
    id: number;
    amount: number;
    status: string;
    created_at: string;
    auction?: {
      id: number;
      status: string;
      current_bid: number;
      ends_at: string;
      vehicle?: {
        year: number;
        make: string;
        model: string;
        images?: { url: string; is_primary: boolean }[];
      };
    };
  };
}

function BidCard({ bid }: BidCardProps) {
  const auction = bid.auction;
  const vehicle = auction?.vehicle;
  const isWinning = bid.status === 'accepted' && auction?.current_bid === bid.amount;

  return (
    <Card className="overflow-hidden">
      <div className="flex flex-col md:flex-row">
        {/* Image */}
        <Link 
          to={`/auctions/${auction?.id}`}
          className="md:w-48 h-32 md:h-auto bg-gray-100"
        >
          <img
            src={vehicle?.images?.find(i => i.is_primary)?.url || vehicle?.images?.[0]?.url || 'https://images.unsplash.com/photo-1494976388531-d1058494cdd8?w=200'}
            alt={`${vehicle?.year} ${vehicle?.make} ${vehicle?.model}`}
            className="w-full h-full object-cover"
          />
        </Link>
        
        {/* Details */}
        <CardContent className="flex-1 p-4">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <Link 
                to={`/auctions/${auction?.id}`}
                className="font-semibold text-lg hover:text-blue-600"
              >
                {vehicle?.year} {vehicle?.make} {vehicle?.model}
              </Link>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant={bid.status === 'accepted' ? 'default' : 'secondary'}>
                  {bid.status === 'accepted' && isWinning && (
                    <CheckCircle className="h-3 w-3 mr-1" />
                  )}
                  {bid.status}
                </Badge>
                <span className="text-sm text-gray-500">
                  {new Date(bid.created_at).toLocaleString()}
                </span>
              </div>
            </div>
            
            <div className="text-right">
              <p className="text-sm text-gray-500">Your Bid</p>
              <p className="text-xl font-bold">${bid.amount.toLocaleString()}</p>
              {auction && (
                <p className="text-sm text-gray-500">
                  Current: ${auction.current_bid.toLocaleString()}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </div>
    </Card>
  );
}

function MyBidsPageSkeleton() {
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

