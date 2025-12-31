/**
 * NotificationBell - Notification icon with dropdown
 */
import { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Bell } from 'lucide-react';
import { useNotifications, useMarkNotificationRead, getNotificationIcon, getTimeAgo } from '@/hooks';
import { Button } from '@/components/ui/button';
import type { Notification } from '@/hooks';

interface NotificationBellProps {
  className?: string;
}

export function NotificationBell({ className = '' }: NotificationBellProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { notifications, unreadCount, isLoading } = useNotifications();
  const markRead = useMarkNotificationRead();

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleNotificationClick = (notification: Notification) => {
    if (!notification.read_at) {
      markRead.mutate(notification.id);
    }
    setIsOpen(false);
  };

  const recentNotifications = notifications.slice(0, 5);

  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      {/* Bell button */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2"
      >
        <Bell size={20} />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center font-bold">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </Button>

      {/* Dropdown */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border z-50">
          <div className="p-3 border-b flex justify-between items-center">
            <h3 className="font-semibold">Notifications</h3>
            {unreadCount > 0 && (
              <span className="text-xs text-gray-500">{unreadCount} unread</span>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {isLoading ? (
              <div className="p-4 text-center text-gray-500">Loading...</div>
            ) : recentNotifications.length === 0 ? (
              <div className="p-4 text-center text-gray-500">
                No notifications yet
              </div>
            ) : (
              recentNotifications.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  onClick={() => handleNotificationClick(notification)}
                />
              ))
            )}
          </div>

          {notifications.length > 5 && (
            <div className="p-2 border-t">
              <Link
                to="/notifications"
                className="block text-center text-sm text-blue-600 hover:text-blue-800"
                onClick={() => setIsOpen(false)}
              >
                View all notifications
              </Link>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

interface NotificationItemProps {
  notification: Notification;
  onClick: () => void;
}

function NotificationItem({ notification, onClick }: NotificationItemProps) {
  const isUnread = !notification.read_at;
  const icon = getNotificationIcon(notification.type);
  const timeAgo = getTimeAgo(notification.created_at);

  // Get link based on notification type
  const getLink = () => {
    const data = notification.data as Record<string, unknown> | undefined;
    if (data?.auction_id) {
      return `/auctions/${data.auction_id}`;
    }
    return '#';
  };

  return (
    <Link
      to={getLink()}
      onClick={onClick}
      className={`block p-3 hover:bg-gray-50 border-b last:border-b-0 ${
        isUnread ? 'bg-blue-50' : ''
      }`}
    >
      <div className="flex gap-3">
        <span className="text-xl">{icon}</span>
        <div className="flex-1 min-w-0">
          <p className={`text-sm ${isUnread ? 'font-semibold' : ''} line-clamp-1`}>
            {notification.title}
          </p>
          <p className="text-xs text-gray-500 line-clamp-2 mt-0.5">
            {notification.message}
          </p>
          <p className="text-xs text-gray-400 mt-1">{timeAgo}</p>
        </div>
        {isUnread && (
          <div className="w-2 h-2 bg-blue-500 rounded-full mt-2" />
        )}
      </div>
    </Link>
  );
}

export default NotificationBell;

