/**
 * useAuctions - hook for fetching and filtering auction lists.
 */
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { auctionsApi } from '@/services/api';
import type { Auction, AuctionFilters, PlaceBidRequest } from '@/types';

/**
 * Fetch list of auctions with optional filters
 */
export function useAuctions(filters?: AuctionFilters) {
  const query = useQuery({
    queryKey: ['auctions', filters],
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (filters?.status && filters.status !== 'all') {
        params.status = filters.status;
      }
      if (filters?.ending_soon) {
        params.ending_soon = 'true';
      }
      const response = await auctionsApi.list(params);
      return response.data;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  // Client-side filtering for instant response
  const auctions = useMemo(() => {
    const allAuctions: Auction[] = query.data?.auctions || [];
    return allAuctions.filter((a) => {
      if (filters?.make && a.vehicle?.make) {
        if (!a.vehicle.make.toLowerCase().includes(filters.make.toLowerCase())) {
          return false;
        }
      }
      if (filters?.min_price && a.current_bid < filters.min_price) {
        return false;
      }
      if (filters?.max_price && a.current_bid > filters.max_price) {
        return false;
      }
      return true;
    });
  }, [query.data, filters]);

  // Sort by ending soon first
  const sortedAuctions = useMemo(() => {
    return [...auctions].sort((a, b) => {
      const endA = new Date(a.ends_at).getTime();
      const endB = new Date(b.ends_at).getTime();
      return endA - endB;
    });
  }, [auctions]);

  return {
    auctions: sortedAuctions,
    total: query.data?.total || 0,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * Fetch active auctions only
 */
export function useActiveAuctions() {
  return useAuctions({ status: 'active' });
}

/**
 * Fetch auctions ending soon (within 24 hours)
 */
export function useEndingSoonAuctions() {
  const { auctions, ...rest } = useAuctions({ status: 'active' });
  
  const endingSoon = useMemo(() => {
    const now = Date.now();
    const twentyFourHours = 24 * 60 * 60 * 1000;
    return auctions.filter((a) => {
      const endsAt = new Date(a.ends_at).getTime();
      return endsAt - now < twentyFourHours && endsAt > now;
    });
  }, [auctions]);

  return {
    auctions: endingSoon,
    ...rest,
  };
}

/**
 * Fetch a single auction by ID
 */
export function useAuction(id: number | undefined) {
  const query = useQuery({
    queryKey: ['auction', id],
    queryFn: async () => {
      if (!id) throw new Error('No auction ID');
      const response = await auctionsApi.get(id);
      return response.data;
    },
    enabled: !!id,
    refetchInterval: 10000, // Refresh every 10 seconds for active auction
  });

  return {
    auction: query.data?.auction as Auction | undefined,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * Fetch bid history for an auction
 */
export function useAuctionBids(auctionId: number | undefined) {
  const query = useQuery({
    queryKey: ['auction-bids', auctionId],
    queryFn: async () => {
      if (!auctionId) throw new Error('No auction ID');
      const response = await auctionsApi.getBids(auctionId);
      return response.data;
    },
    enabled: !!auctionId,
    refetchInterval: 5000, // Refresh every 5 seconds
  });

  return {
    bids: query.data?.bids || [],
    total: query.data?.total || 0,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * Place a bid on an auction
 */
export function usePlaceBid(auctionId: number) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (data: PlaceBidRequest) => {
      const response = await auctionsApi.placeBid(auctionId, data.amount);
      return response.data;
    },
    onSuccess: () => {
      // Invalidate auction queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['auction', auctionId] });
      queryClient.invalidateQueries({ queryKey: ['auction-bids', auctionId] });
      queryClient.invalidateQueries({ queryKey: ['auctions'] });
    },
  });

  return {
    placeBid: mutation.mutate,
    placeBidAsync: mutation.mutateAsync,
    isLoading: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
    data: mutation.data,
    reset: mutation.reset,
  };
}

/**
 * Get time remaining for an auction
 */
export function getTimeRemaining(endsAt: string): {
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
  total: number;
  isEnded: boolean;
  isUrgent: boolean;
} {
  const now = Date.now();
  const end = new Date(endsAt).getTime();
  const total = end - now;

  if (total <= 0) {
    return {
      days: 0,
      hours: 0,
      minutes: 0,
      seconds: 0,
      total: 0,
      isEnded: true,
      isUrgent: false,
    };
  }

  const seconds = Math.floor((total / 1000) % 60);
  const minutes = Math.floor((total / 1000 / 60) % 60);
  const hours = Math.floor((total / (1000 * 60 * 60)) % 24);
  const days = Math.floor(total / (1000 * 60 * 60 * 24));

  return {
    days,
    hours,
    minutes,
    seconds,
    total,
    isEnded: false,
    isUrgent: total < 60 * 60 * 1000, // Less than 1 hour
  };
}

