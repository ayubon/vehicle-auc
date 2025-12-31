/**
 * WatchlistButton - Toggle auction watchlist status
 */
import { Heart } from 'lucide-react';
import { useToggleWatchlist } from '@/hooks';
import { Button } from '@/components/ui/button';

interface WatchlistButtonProps {
  auctionId: number;
  size?: 'sm' | 'md' | 'lg';
  showText?: boolean;
  className?: string;
}

export function WatchlistButton({
  auctionId,
  size = 'md',
  showText = false,
  className = '',
}: WatchlistButtonProps) {
  const { isWatching, toggle, isLoading } = useToggleWatchlist(auctionId);

  const iconSize = {
    sm: 16,
    md: 20,
    lg: 24,
  }[size];

  const handleClick = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      await toggle();
    } catch (error) {
      console.error('Failed to toggle watchlist:', error);
    }
  };

  return (
    <Button
      variant="ghost"
      size={size === 'sm' ? 'sm' : 'default'}
      onClick={handleClick}
      disabled={isLoading}
      className={`
        ${isWatching ? 'text-red-500 hover:text-red-600' : 'text-gray-400 hover:text-red-500'}
        ${className}
      `}
      title={isWatching ? 'Remove from watchlist' : 'Add to watchlist'}
    >
      <Heart
        size={iconSize}
        className={`transition-all ${isLoading ? 'animate-pulse' : ''}`}
        fill={isWatching ? 'currentColor' : 'none'}
      />
      {showText && (
        <span className="ml-2">
          {isWatching ? 'Watching' : 'Watch'}
        </span>
      )}
    </Button>
  );
}

export default WatchlistButton;

