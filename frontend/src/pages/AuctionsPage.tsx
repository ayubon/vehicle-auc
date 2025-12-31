/**
 * AuctionsPage - Browse all auctions with filters
 */
import { useState } from 'react';
import { useAuctions } from '@/hooks';
import { AuctionCard, AuctionCardSkeleton } from '@/components/AuctionCard';
import { Button } from '@/components/ui/button';
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from '@/components/ui/select';
import { Gavel, AlertCircle, Search } from 'lucide-react';
import { Input } from '@/components/ui/input';

type AuctionStatus = 'all' | 'active' | 'scheduled' | 'ended';

export default function AuctionsPage() {
  const [statusFilter, setStatusFilter] = useState<AuctionStatus>('active');
  const [searchQuery, setSearchQuery] = useState('');
  
  const { auctions, isLoading, error, refetch } = useAuctions(
    statusFilter === 'all' ? undefined : statusFilter
  );

  // Filter auctions by search query
  const filteredAuctions = auctions.filter((auction) => {
    if (!searchQuery) return true;
    
    const vehicle = auction.vehicle;
    const searchLower = searchQuery.toLowerCase();
    
    return (
      vehicle?.make?.toLowerCase().includes(searchLower) ||
      vehicle?.model?.toLowerCase().includes(searchLower) ||
      vehicle?.year?.toString().includes(searchQuery)
    );
  });

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="container mx-auto px-4 py-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div>
              <h1 className="text-3xl font-bold flex items-center gap-2">
                <Gavel className="h-8 w-8" />
                Auctions
              </h1>
              <p className="text-gray-500 mt-1">
                Browse {statusFilter === 'all' ? 'all' : statusFilter} auctions
              </p>
            </div>
            
            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-3">
              {/* Search */}
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                <Input
                  type="text"
                  placeholder="Search make, model, year..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 w-full sm:w-64"
                />
              </div>
              
              {/* Status filter */}
              <Select
                value={statusFilter}
                onValueChange={(v) => setStatusFilter(v as AuctionStatus)}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="scheduled">Upcoming</SelectItem>
                  <SelectItem value="ended">Ended</SelectItem>
                </SelectContent>
              </Select>
            </div>
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
              {searchQuery 
                ? 'No auctions match your search'
                : `No ${statusFilter === 'all' ? '' : statusFilter} auctions`
              }
            </h2>
            <p className="text-gray-500 mb-6 max-w-md mx-auto">
              {searchQuery
                ? 'Try adjusting your search terms.'
                : 'Check back soon for new listings!'}
            </p>
            {(searchQuery || statusFilter !== 'active') && (
              <div className="flex gap-3 justify-center">
                {searchQuery && (
                  <Button variant="outline" onClick={() => setSearchQuery('')}>
                    Clear Search
                  </Button>
                )}
                {statusFilter !== 'active' && (
                  <Button onClick={() => setStatusFilter('active')}>
                    View Active Auctions
                  </Button>
                )}
              </div>
            )}
          </div>
        )}

        {/* Auction Grid */}
        {!isLoading && filteredAuctions.length > 0 && (
          <>
            <p className="text-sm text-gray-500 mb-4">
              Showing {filteredAuctions.length} auction{filteredAuctions.length !== 1 ? 's' : ''}
            </p>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredAuctions.map((auction) => (
                <AuctionCard 
                  key={auction.id} 
                  auction={auction}
                  showBidForm={auction.status === 'active'}
                />
              ))}
            </div>
          </>
        )}
      </main>
    </div>
  );
}
