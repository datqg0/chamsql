import { ChevronLeft, ChevronRight } from "lucide-react";
import React from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import * as types from "@/types/exam-submission.types";

interface ProblemHeaderProps {
  problem?: types.ExamProblemDetail | null;
  problemIndex: number;
  totalProblems: number;
  points?: number;
  submissionResult?: types.ProblemSubmissionResult | null;
  onPrevious: () => void;
  onNext: () => void;
  canGoBack: boolean;
  canGoNext: boolean;
}

export const ProblemHeader: React.FC<ProblemHeaderProps> = ({
  problem,
  problemIndex,
  totalProblems,
  points = 0,
  submissionResult,
  onPrevious,
  onNext,
  canGoBack,
  canGoNext,
}) => {
  const getDifficultyColor = (difficulty?: string) => {
    const diff = difficulty?.toLowerCase() || "medium";
    switch (diff) {
      case "easy":
        return "bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-300";
      case "medium":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-950 dark:text-yellow-300";
      case "hard":
        return "bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "accepted":
        return "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300 border-green-200";
      case "wrong_answer":
        return "bg-orange-50 text-orange-700 dark:bg-orange-950 dark:text-orange-300 border-orange-200";
      case "error":
        return "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300 border-red-200";
      default:
        return "bg-gray-50 text-gray-700 dark:bg-gray-800 dark:text-gray-300 border-gray-200";
    }
  };

  return (
    <div className="border-b border-border bg-card p-4">
      {/* Top Row: Navigation & Problem Count */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={onPrevious}
            disabled={!canGoBack}
            className="p-2"
            title="Previous Problem"
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>

          <span className="text-sm font-medium text-muted-foreground px-3">
            Problem {problemIndex} of {totalProblems}
          </span>

          <Button
            variant="ghost"
            size="sm"
            onClick={onNext}
            disabled={!canGoNext}
            className="p-2"
            title="Next Problem"
          >
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>

        {/* Points Badge */}
        <Badge variant="outline" className="bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300 border-blue-200">
          {points || (problem?.points ?? 0)} points
        </Badge>
      </div>

      {/* Middle Row: Title & Difficulty */}
      {problem && (
        <div className="mb-4">
          <h2 className="text-xl font-bold mb-2">{problem.title}</h2>
          <div className="flex items-center gap-2">
            <Badge className={getDifficultyColor(problem.difficulty)}>
              {problem.difficulty}
            </Badge>
          </div>
        </div>
      )}

      {/* Bottom Row: Submission Status (if available) */}
      {submissionResult && (
        <div className={`p-3 rounded-lg border-2 ${getStatusColor(submissionResult.status)} flex items-center justify-between`}>
          <div>
            {submissionResult.status === "accepted" && (
              <p className="font-semibold">✓ Correct!</p>
            )}
            {submissionResult.status === "wrong_answer" && (
              <p className="font-semibold">✗ Wrong Answer</p>
            )}
            {submissionResult.status === "error" && (
              <p className="font-semibold">⚠ Execution Error</p>
            )}
            {submissionResult.status === "accepted" && (
              <p className="text-sm">You earned {submissionResult.score} points</p>
            )}
          </div>
          <div className="text-right space-y-1">
            <p className="text-xs text-muted-foreground">
              Attempt {submissionResult.attemptNumber}
            </p>
            {submissionResult.executionTimeMs !== undefined && (
              <p className="text-xs font-mono font-medium">
                ⏱ {submissionResult.executionTimeMs}ms
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ProblemHeader;
