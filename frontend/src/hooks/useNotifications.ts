/**
 * useNotifications - hook for managing user notifications
 */
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationsApi } from '@/services/api';

export interface Notification {
  id: number;
  user_id: number;
  type: 'outbid' | 'auction_ending' | 'bid_accepted' | 'auction_won' | 'auction_lost' | 'payment_reminder';
  title: string;
  message: string;
  data?: Record<string, unknown>;
  read_at: string | null;
  created_at: string;
}

interface NotificationsResponse {
  notifications: Notification[];
  total: number;
  unread_count: number;
}

/**
 * Fetch user's notifications
 */
export function useNotifications() {
  const query = useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const response = await notificationsApi.list();
      return response.data as NotificationsResponse;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  return {
    notifications: query.data?.notifications || [],
    total: query.data?.total || 0,
    unreadCount: query.data?.unread_count || 0,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

/**
 * Get just the unread count (lightweight)
 */
export function useUnreadCount() {
  const query = useQuery({
    queryKey: ['notifications-unread-count'],
    queryFn: async () => {
      const response = await notificationsApi.unreadCount();
      return response.data as { count: number };
    },
    refetchInterval: 15000, // Refresh every 15 seconds
  });

  return {
    count: query.data?.count || 0,
    isLoading: query.isLoading,
  };
}

/**
 * Mark notification as read
 */
export function useMarkNotificationRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (notificationId: number) => {
      const response = await notificationsApi.markRead(notificationId);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] });
    },
  });
}

/**
 * Mark all notifications as read
 */
export function useMarkAllNotificationsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await notificationsApi.markAllRead();
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] });
    },
  });
}

/**
 * Get notification icon based on type
 */
export function getNotificationIcon(type: Notification['type']): string {
  switch (type) {
    case 'outbid':
      return '⚠️';
    case 'auction_ending':
      return '⏰';
    case 'bid_accepted':
      return '✅';
    case 'auction_won':
      return '🎉';
    case 'auction_lost':
      return '😔';
    case 'payment_reminder':
      return '💳';
    default:
      return '📢';
  }
}

/**
 * Get time ago string
 */
export function getTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
  return date.toLocaleDateString();
}

