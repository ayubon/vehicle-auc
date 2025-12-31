/**
 * useAuctionSSE - Real-time auction updates via Server-Sent Events
 */
import { useState, useEffect, useCallback, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { BidUpdateEvent, AuctionEndEvent, SSEEvent } from '@/types';

interface SSEState {
  currentBid: number;
  bidCount: number;
  lastBidderId?: number;
  isConnected: boolean;
  isEnded: boolean;
  error: string | null;
}

interface UseAuctionSSEOptions {
  onBidUpdate?: (event: BidUpdateEvent) => void;
  onAuctionEnd?: (event: AuctionEndEvent) => void;
  onError?: (error: Error) => void;
  enabled?: boolean;
}

/**
 * Hook to subscribe to real-time auction updates via SSE
 */
export function useAuctionSSE(
  auctionId: number | undefined,
  initialBid: number = 0,
  initialBidCount: number = 0,
  options: UseAuctionSSEOptions = {}
) {
  const { onBidUpdate, onAuctionEnd, onError, enabled = true } = options;
  const queryClient = useQueryClient();
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;
  const baseReconnectDelay = 1000;

  const [state, setState] = useState<SSEState>({
    currentBid: initialBid,
    bidCount: initialBidCount,
    lastBidderId: undefined,
    isConnected: false,
    isEnded: false,
    error: null,
  });

  // Update initial values when props change
  useEffect(() => {
    setState(prev => ({
      ...prev,
      currentBid: initialBid || prev.currentBid,
      bidCount: initialBidCount || prev.bidCount,
    }));
  }, [initialBid, initialBidCount]);

  const connect = useCallback(() => {
    if (!auctionId || !enabled) return;

    // Close existing connection
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const url = `/api/sse/auctions/${auctionId}`;
    console.log('[SSE] Connecting to:', url);
    
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      console.log('[SSE] Connected to auction', auctionId);
      reconnectAttempts.current = 0;
      setState(prev => ({
        ...prev,
        isConnected: true,
        error: null,
      }));
    };

    eventSource.onmessage = (event) => {
      try {
        const data: SSEEvent = JSON.parse(event.data);
        console.log('[SSE] Received event:', data);

        if (data.type === 'bid_update') {
          const bidEvent = data as BidUpdateEvent;
          setState(prev => ({
            ...prev,
            currentBid: bidEvent.current_bid,
            bidCount: bidEvent.bid_count,
            lastBidderId: bidEvent.bidder_id,
          }));
          
          // Invalidate queries to refresh data
          queryClient.invalidateQueries({ queryKey: ['auction', auctionId] });
          queryClient.invalidateQueries({ queryKey: ['auction-bids', auctionId] });
          
          onBidUpdate?.(bidEvent);
        } else if (data.type === 'auction_ended') {
          const endEvent = data as AuctionEndEvent;
          setState(prev => ({
            ...prev,
            isEnded: true,
          }));
          
          // Invalidate all auction queries
          queryClient.invalidateQueries({ queryKey: ['auctions'] });
          queryClient.invalidateQueries({ queryKey: ['auction', auctionId] });
          
          onAuctionEnd?.(endEvent);
        }
      } catch (err) {
        console.error('[SSE] Failed to parse event:', err);
      }
    };

    eventSource.onerror = (error) => {
      console.error('[SSE] Connection error:', error);
      setState(prev => ({
        ...prev,
        isConnected: false,
        error: 'Connection lost',
      }));

      eventSource.close();
      eventSourceRef.current = null;

      // Attempt reconnection with exponential backoff
      if (reconnectAttempts.current < maxReconnectAttempts) {
        const delay = baseReconnectDelay * Math.pow(2, reconnectAttempts.current);
        console.log(`[SSE] Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current + 1})`);
        
        reconnectTimeoutRef.current = setTimeout(() => {
          reconnectAttempts.current++;
          connect();
        }, delay);
      } else {
        const err = new Error('Max reconnection attempts reached');
        setState(prev => ({
          ...prev,
          error: err.message,
        }));
        onError?.(err);
      }
    };
  }, [auctionId, enabled, queryClient, onBidUpdate, onAuctionEnd, onError]);

  // Connect on mount and when auctionId changes
  useEffect(() => {
    if (enabled && auctionId) {
      connect();
    }

    return () => {
      if (eventSourceRef.current) {
        console.log('[SSE] Closing connection');
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };
  }, [auctionId, enabled, connect]);

  // Manual reconnect function
  const reconnect = useCallback(() => {
    reconnectAttempts.current = 0;
    connect();
  }, [connect]);

  return {
    ...state,
    reconnect,
  };
}

/**
 * Hook to subscribe to multiple auctions at once (for homepage grid)
 */
export function useMultiAuctionSSE(auctionIds: number[]) {
  const [updates, setUpdates] = useState<Map<number, { currentBid: number; bidCount: number }>>(new Map());
  const eventSourcesRef = useRef<Map<number, EventSource>>(new Map());

  useEffect(() => {
    // Clean up old connections
    const currentIds = new Set(auctionIds);
    eventSourcesRef.current.forEach((es, id) => {
      if (!currentIds.has(id)) {
        es.close();
        eventSourcesRef.current.delete(id);
      }
    });

    // Create new connections
    auctionIds.forEach((id) => {
      if (!eventSourcesRef.current.has(id)) {
        const es = new EventSource(`/api/sse/auctions/${id}`);
        
        es.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'bid_update') {
              setUpdates(prev => new Map(prev).set(id, {
                currentBid: data.current_bid,
                bidCount: data.bid_count,
              }));
            }
          } catch {
            // Ignore parse errors
          }
        };

        es.onerror = () => {
          es.close();
          eventSourcesRef.current.delete(id);
        };

        eventSourcesRef.current.set(id, es);
      }
    });

    return () => {
      eventSourcesRef.current.forEach((es) => es.close());
      eventSourcesRef.current.clear();
    };
  }, [auctionIds.join(',')]); // eslint-disable-line react-hooks/exhaustive-deps

  return updates;
}

