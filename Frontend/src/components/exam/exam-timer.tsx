import { Clock, AlertCircle } from "lucide-react";
import React from "react";

import { useExamTimer } from "@/hooks/useExamTimer";

interface ExamTimerProps {
  endTimeMs: number | null;
  onTimeExpired?: () => void;
  examTitle?: string;
}

export const ExamTimer: React.FC<ExamTimerProps> = ({
  endTimeMs,
  onTimeExpired,
  examTitle,
}) => {
  const { formattedTime, isExpired, percentage } = useExamTimer(
    endTimeMs,
    onTimeExpired,
    1000
  );

  // Determine color based on percentage
  const getTimerColor = () => {
    if (isExpired) return "text-red-600";
    if (percentage <= 5) return "text-red-500";
    if (percentage <= 10) return "text-orange-500";
    if (percentage <= 25) return "text-yellow-500";
    return "text-green-600";
  };

  const getProgressColor = () => {
    if (isExpired) return "bg-red-600";
    if (percentage <= 5) return "bg-red-500";
    if (percentage <= 10) return "bg-orange-500";
    if (percentage <= 25) return "bg-yellow-500";
    return "bg-green-600";
  };

  const isWarning = percentage <= 25;
  const isCritical = percentage <= 5;

  return (
    <div className="flex items-center gap-4 p-4 bg-card border-b border-border">
      {/* Timer Display */}
      <div className="flex items-center gap-2">
        <Clock className={`w-5 h-5 ${getTimerColor()}`} />
        <div className="flex flex-col">
          <span className="text-xs text-muted-foreground">Time Remaining</span>
          <span className={`text-2xl font-mono font-bold ${getTimerColor()}`}>
            {formattedTime}
          </span>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
        <div
          className={`h-full ${getProgressColor()} transition-all duration-500`}
          style={{ width: `${percentage}%` }}
        />
      </div>

      {/* Exam Title */}
      {examTitle && (
        <div className="flex-1 text-right">
          <p className="text-sm font-medium">{examTitle}</p>
        </div>
      )}

      {/* Warning Icon */}
      {(isWarning || isExpired) && (
        <div className="flex items-center gap-2">
          <AlertCircle className={`w-5 h-5 ${isCritical ? "text-red-600 animate-pulse" : "text-orange-500"}`} />
          {isCritical && (
            <span className="text-xs font-semibold text-red-600 animate-pulse">
              HURRY UP!
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default ExamTimer;
