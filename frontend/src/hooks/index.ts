/**
 * Re-export all hooks.
 */
export { useVehicles } from './useVehicles';
export { useVehicle } from './useVehicle';
export { useAuth } from './useAuth';
export { 
  useAuctions, 
  useActiveAuctions, 
  useEndingSoonAuctions, 
  useAuction, 
  useAuctionBids, 
  usePlaceBid,
  getTimeRemaining 
} from './useAuctions';
export { useMyBids } from './useMyBids';
export { useAuctionSSE, useMultiAuctionSSE } from './useAuctionSSE';
export { 
  useWatchlist, 
  useIsWatching, 
  useAddToWatchlist, 
  useRemoveFromWatchlist, 
  useToggleWatchlist 
} from './useWatchlist';
export { 
  useNotifications, 
  useUnreadCount, 
  useMarkNotificationRead, 
  useMarkAllNotificationsRead,
  getNotificationIcon,
  getTimeAgo
} from './useNotifications';
export type { Notification } from './useNotifications';
