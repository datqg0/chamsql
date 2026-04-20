import { useEffect, useState, useCallback } from "react";

interface TimerState {
  remainingMs: number;
  formattedTime: string;
  isExpired: boolean;
  percentage: number; // 0-100 for progress
}

/**
 * Hook to manage exam timer countdown
 * @param endTimeMs - End time in milliseconds (from now)
 * @param onTimeExpired - Callback when timer reaches 0
 * @param pollIntervalMs - How often to update (default 1000ms)
 */
export const useExamTimer = (
  endTimeMs: number | null,
  onTimeExpired?: () => void,
  pollIntervalMs: number = 1000
): TimerState => {
  const [remainingMs, setRemainingMs] = useState<number>(endTimeMs || 0);
  const [isExpired, setIsExpired] = useState(false);

  const formatTime = useCallback((ms: number): string => {
    if (ms <= 0) return "00:00";

    const totalSeconds = Math.floor(ms / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;

    return `${minutes.toString().padStart(2, "0")}:${seconds
      .toString()
      .padStart(2, "0")}`;
  }, []);

  const formattedTime = formatTime(remainingMs);

  // Calculate percentage (0-100)
  const percentage = endTimeMs
    ? Math.max(0, Math.min(100, (remainingMs / endTimeMs) * 100))
    : 100;

  useEffect(() => {
    if (!endTimeMs || isExpired) return;

    const timer = setInterval(() => {
      setRemainingMs((prev) => {
        const newRemaining = prev - pollIntervalMs;

        if (newRemaining <= 0) {
          setIsExpired(true);
          onTimeExpired?.();
          clearInterval(timer);
          return 0;
        }

        return newRemaining;
      });
    }, pollIntervalMs);

    return () => clearInterval(timer);
  }, [endTimeMs, isExpired, onTimeExpired, pollIntervalMs]);

  return {
    remainingMs: Math.max(0, remainingMs),
    formattedTime,
    isExpired,
    percentage,
  };
};
