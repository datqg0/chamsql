import { useEffect, useState, useCallback } from "react";

interface TimerState {
  remainingMs: number;
  formattedTime: string;
  isExpired: boolean;
  percentage: number; // 0-100 for progress
}

/**
 * Hook to manage exam timer countdown with optional server synchronization
 * @param endTimeMs - Initial end time in milliseconds (from now)
 * @param onTimeExpired - Callback when timer reaches 0
 * @param pollIntervalMs - How often to update UI (default 1000ms)
 * @param syncIntervalMs - How often to sync with server (default 30000ms)
 * @param syncFn - Optional function to fetch exact remaining time from server
 */
export const useExamTimer = (
  endTimeMs: number | null,
  onTimeExpired?: () => void,
  pollIntervalMs: number = 1000,
  syncIntervalMs: number = 30000,
  syncFn?: () => Promise<number>
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

  // UI Tick
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

  // Server Sync
  useEffect(() => {
    if (!syncFn || !endTimeMs || isExpired) return;

    const syncTimer = setInterval(async () => {
      try {
        const serverRemaining = await syncFn();
        if (serverRemaining !== undefined) {
          setRemainingMs(serverRemaining);
          if (serverRemaining <= 0) {
            setIsExpired(true);
            onTimeExpired?.();
          }
        }
      } catch (error) {
        console.error("Failed to sync exam timer:", error);
      }
    }, syncIntervalMs);

    return () => clearInterval(syncTimer);
  }, [syncFn, endTimeMs, isExpired, onTimeExpired, syncIntervalMs]);

  return {
    remainingMs: Math.max(0, remainingMs),
    formattedTime,
    isExpired,
    percentage,
  };
};
