/**
 * useMyBids - hook for fetching user's bid history
 */
import { useQuery } from '@tanstack/react-query';
import { auctionsApi } from '@/services/api';
import { useAuth } from './useAuth';
import type { Auction, Bid } from '@/types';

interface MyBid extends Bid {
  auction?: Auction;
}

/**
 * Fetch current user's bids
 */
export function useMyBids() {
  const { user } = useAuth();
  
  const query = useQuery({
    queryKey: ['my-bids', user?.id],
    queryFn: async () => {
      const response = await auctionsApi.getMyBids();
      return response.data;
    },
    enabled: !!user?.id,
    refetchInterval: 30000,
  });

  const bids: MyBid[] = query.data?.bids || [];

  // Categorize bids by status
  const activeBids = bids.filter(
    (b) => b.status === 'accepted' && b.auction?.status === 'active'
  );
  
  const wonBids = bids.filter(
    (b) => b.status === 'accepted' && b.auction?.status === 'ended'
  );
  
  const outbidBids = bids.filter(
    (b) => b.status === 'outbid'
  );

  return {
    activeBids,
    wonBids,
    outbidBids,
    allBids: bids,
    isLoading: query.isLoading,
    error: query.error,
  };
}

