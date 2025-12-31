/**
 * CountdownTimer - Displays time remaining until auction ends
 */
import { useState, useEffect } from 'react';
import { getTimeRemaining } from '@/hooks';

interface CountdownTimerProps {
  endsAt: string;
  onEnd?: () => void;
  showDays?: boolean;
  compact?: boolean;
  className?: string;
}

export function CountdownTimer({
  endsAt,
  onEnd,
  showDays = true,
  compact = false,
  className = '',
}: CountdownTimerProps) {
  const [timeLeft, setTimeLeft] = useState(() => getTimeRemaining(endsAt));

  useEffect(() => {
    const timer = setInterval(() => {
      const remaining = getTimeRemaining(endsAt);
      setTimeLeft(remaining);

      if (remaining.isEnded) {
        clearInterval(timer);
        onEnd?.();
      }
    }, 1000);

    return () => clearInterval(timer);
  }, [endsAt, onEnd]);

  if (timeLeft.isEnded) {
    return (
      <div className={`text-red-600 font-semibold ${className}`}>
        Auction Ended
      </div>
    );
  }

  if (compact) {
    // Compact format: "2d 5h" or "5h 30m" or "30m 15s"
    let display = '';
    if (timeLeft.days > 0) {
      display = `${timeLeft.days}d ${timeLeft.hours}h`;
    } else if (timeLeft.hours > 0) {
      display = `${timeLeft.hours}h ${timeLeft.minutes}m`;
    } else {
      display = `${timeLeft.minutes}m ${timeLeft.seconds}s`;
    }

    return (
      <span className={`font-mono ${timeLeft.isUrgent ? 'text-red-600 font-bold animate-pulse' : 'text-gray-700'} ${className}`}>
        {display}
      </span>
    );
  }

  // Full format with boxes
  return (
    <div className={`flex gap-2 ${className}`}>
      {showDays && timeLeft.days > 0 && (
        <TimeBox value={timeLeft.days} label="Days" />
      )}
      <TimeBox 
        value={timeLeft.hours} 
        label="Hours" 
        urgent={timeLeft.isUrgent}
      />
      <TimeBox 
        value={timeLeft.minutes} 
        label="Min" 
        urgent={timeLeft.isUrgent}
      />
      <TimeBox 
        value={timeLeft.seconds} 
        label="Sec" 
        urgent={timeLeft.isUrgent}
      />
    </div>
  );
}

interface TimeBoxProps {
  value: number;
  label: string;
  urgent?: boolean;
}

function TimeBox({ value, label, urgent }: TimeBoxProps) {
  return (
    <div className={`flex flex-col items-center ${urgent ? 'text-red-600' : ''}`}>
      <div className={`
        w-12 h-12 flex items-center justify-center 
        rounded-lg font-mono text-xl font-bold
        ${urgent 
          ? 'bg-red-100 border-2 border-red-500 animate-pulse' 
          : 'bg-gray-100 border border-gray-300'
        }
      `}>
        {String(value).padStart(2, '0')}
      </div>
      <span className="text-xs text-gray-500 mt-1">{label}</span>
    </div>
  );
}

export default CountdownTimer;

