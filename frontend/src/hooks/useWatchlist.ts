/**
 * useWatchlist - hook for managing user's watchlist
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { watchlistApi } from '@/services/api';
import type { Auction } from '@/types';

interface WatchlistItem {
  id: number;
  auction_id: number;
  created_at: string;
  auction?: Auction;
}

interface WatchlistResponse {
  items: WatchlistItem[];
  total: number;
}

/**
 * Fetch user's watchlist
 */
export function useWatchlist() {
  const query = useQuery({
    queryKey: ['watchlist'],
    queryFn: async () => {
      const response = await watchlistApi.list();
      return response.data as WatchlistResponse;
    },
  });

  return {
    items: query.data?.items || [],
    total: query.data?.total || 0,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * Check if an auction is in watchlist
 */
export function useIsWatching(auctionId: number) {
  const { items } = useWatchlist();
  return items.some((item) => item.auction_id === auctionId);
}

/**
 * Add auction to watchlist
 */
export function useAddToWatchlist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (auctionId: number) => {
      const response = await watchlistApi.add(auctionId);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['watchlist'] });
    },
  });
}

/**
 * Remove auction from watchlist
 */
export function useRemoveFromWatchlist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (auctionId: number) => {
      const response = await watchlistApi.remove(auctionId);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['watchlist'] });
    },
  });
}

/**
 * Toggle watchlist status (add/remove)
 */
export function useToggleWatchlist(auctionId: number) {
  const queryClient = useQueryClient();
  const { items } = useWatchlist();
  const isWatching = items.some((item) => item.auction_id === auctionId);

  const addMutation = useAddToWatchlist();
  const removeMutation = useRemoveFromWatchlist();

  const toggle = async () => {
    if (isWatching) {
      await removeMutation.mutateAsync(auctionId);
    } else {
      await addMutation.mutateAsync(auctionId);
    }
  };

  return {
    isWatching,
    toggle,
    isLoading: addMutation.isPending || removeMutation.isPending,
    error: addMutation.error || removeMutation.error,
  };
}

