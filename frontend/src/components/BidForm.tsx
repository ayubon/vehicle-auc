/**
 * BidForm - Form for placing bids on auctions
 */
import { useState } from 'react';
import { usePlaceBid } from '@/hooks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface BidFormProps {
  auctionId: number;
  currentBid: number;
  minIncrement?: number;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
  compact?: boolean;
  className?: string;
}

export function BidForm({
  auctionId,
  currentBid,
  minIncrement = 100,
  onSuccess,
  onError,
  compact = false,
  className = '',
}: BidFormProps) {
  const minBid = currentBid + minIncrement;
  const [amount, setAmount] = useState(minBid.toString());
  const [error, setError] = useState<string | null>(null);
  
  const { placeBid, isLoading, isSuccess, reset } = usePlaceBid(auctionId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const bidAmount = parseFloat(amount);
    
    if (isNaN(bidAmount)) {
      setError('Please enter a valid amount');
      return;
    }

    if (bidAmount < minBid) {
      setError(`Minimum bid is $${minBid.toLocaleString()}`);
      return;
    }

    try {
      await placeBid({ amount: bidAmount }, {
        onSuccess: () => {
          onSuccess?.();
          // Update to next minimum bid
          setAmount((bidAmount + minIncrement).toString());
        },
        onError: (err) => {
          const message = err instanceof Error ? err.message : 'Failed to place bid';
          setError(message);
          onError?.(err instanceof Error ? err : new Error(message));
        },
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to place bid';
      setError(message);
    }
  };

  const handleAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/[^0-9.]/g, '');
    setAmount(value);
    setError(null);
    if (isSuccess) reset();
  };

  const quickBids = [
    minBid,
    minBid + minIncrement,
    minBid + minIncrement * 2,
    minBid + minIncrement * 5,
  ];

  if (compact) {
    return (
      <form onSubmit={handleSubmit} className={`flex gap-2 ${className}`}>
        <div className="relative flex-1">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
          <Input
            type="text"
            value={amount}
            onChange={handleAmountChange}
            className="pl-7"
            placeholder={minBid.toString()}
            disabled={isLoading}
          />
        </div>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Placing...' : 'Bid'}
        </Button>
      </form>
    );
  }

  return (
    <form onSubmit={handleSubmit} className={`space-y-4 ${className}`}>
      {/* Current bid display */}
      <div className="text-center">
        <p className="text-sm text-gray-500">Current Bid</p>
        <p className="text-3xl font-bold text-green-600">
          ${currentBid.toLocaleString()}
        </p>
      </div>

      {/* Quick bid buttons */}
      <div className="grid grid-cols-4 gap-2">
        {quickBids.map((qb) => (
          <Button
            key={qb}
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setAmount(qb.toString())}
            className="text-xs"
          >
            ${qb.toLocaleString()}
          </Button>
        ))}
      </div>

      {/* Custom amount input */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Your Bid
        </label>
        <div className="relative">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 text-lg">
            $
          </span>
          <Input
            type="text"
            value={amount}
            onChange={handleAmountChange}
            className="pl-8 text-lg font-semibold"
            placeholder={minBid.toString()}
            disabled={isLoading}
          />
        </div>
        <p className="text-xs text-gray-500 mt-1">
          Minimum bid: ${minBid.toLocaleString()}
        </p>
      </div>

      {/* Error message */}
      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      {/* Success message */}
      {isSuccess && (
        <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
          <p className="text-sm text-green-600">Bid placed successfully!</p>
        </div>
      )}

      {/* Submit button */}
      <Button
        type="submit"
        className="w-full"
        size="lg"
        disabled={isLoading}
      >
        {isLoading ? (
          <>
            <span className="animate-spin mr-2">⏳</span>
            Placing Bid...
          </>
        ) : (
          `Place Bid - $${parseFloat(amount || '0').toLocaleString()}`
        )}
      </Button>

      <p className="text-xs text-center text-gray-400">
        By placing a bid, you agree to our terms of service
      </p>
    </form>
  );
}

export default BidForm;

