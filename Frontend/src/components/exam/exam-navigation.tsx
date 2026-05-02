import { CheckCircle, Circle } from "lucide-react";
import React from "react";

import { Badge } from "@/components/ui/badge";
import * as types from "@/types/exam-submission.types";

interface ExamNavigationProps {
  problems: types.ExamProblemDetail[];
  currentProblemIndex: number;
  solvedProblems: Record<number, boolean>;
  onSelectProblem: (index: number) => void;
}

export const ExamNavigation: React.FC<ExamNavigationProps> = ({
  problems,
  currentProblemIndex,
  solvedProblems,
  onSelectProblem,
}) => {
  const getStatusIcon = (problemId: number) => {
    if (solvedProblems[problemId]) {
      return <CheckCircle className="w-4 h-4 text-green-500" />;
    }
    return <Circle className="w-4 h-4 text-muted-foreground" />;
  };

  const getStatusBadge = (problemId: number) => {
    if (solvedProblems[problemId]) {
      return (
        <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
          ✓ Answered
        </Badge>
      );
    }
    return (
      <Badge variant="outline" className="bg-gray-50 text-gray-700 border-gray-200">
        ✗ Unanswered
      </Badge>
    );
  };

  const solvedCount = Object.keys(solvedProblems).length;
  const totalProblems = problems.length;

  return (
    <div className="w-64 bg-card border-r border-border flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-border">
        <h3 className="font-semibold text-sm mb-2">Problems</h3>
        <p className="text-xs text-muted-foreground">
          {solvedCount} of {totalProblems} solved
        </p>
        <div className="mt-2 bg-muted rounded-full h-2 overflow-hidden">
          <div
            className="bg-green-500 h-full transition-all"
            style={{ width: `${(solvedCount / totalProblems) * 100}%` }}
          />
        </div>
      </div>

      {/* Problem List */}
      <div className="flex-1 overflow-y-auto">
        <div className="p-2 space-y-2">
          {problems.map((problem, idx) => (
            <button
              key={problem.examProblemID}
              onClick={() => onSelectProblem(idx)}
              className={`w-full text-left p-3 rounded-lg border-2 transition-colors ${
                currentProblemIndex === idx
                  ? "border-blue-500 bg-blue-50 dark:bg-blue-950"
                  : "border-transparent hover:bg-muted"
              }`}
            >
              <div className="flex items-start justify-between gap-2 mb-1">
                <div className="flex items-center gap-2">
                  {getStatusIcon(problem.examProblemID)}
                  <span className="text-xs font-semibold text-muted-foreground">
                    Q{idx + 1}
                  </span>
                </div>
                <span className="text-xs bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 px-2 py-1 rounded">
                  {problem.points}pts
                </span>
              </div>

              <p className="text-sm font-medium line-clamp-2 mb-2">
                {problem.title}
              </p>

              <div className="flex items-center justify-between">
                {getStatusBadge(problem.examProblemID)}
                <span className="text-xs text-muted-foreground">
                  {problem.difficulty}
                </span>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Footer Stats */}
      <div className="p-4 border-t border-border bg-muted/30 space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground">Solved:</span>
          <span className="font-semibold text-green-600">
            {solvedCount}/{totalProblems}
          </span>
        </div>
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground">Total Points:</span>
          <span className="font-semibold">
            {Object.keys(solvedProblems)
              .map(id => problems.find(p => p.examProblemID.toString() === id))
              .filter(p => p)
              .reduce((sum, p) => sum + (p?.points || 0), 0)}/{" "}
            {problems.reduce((sum, p) => sum + p.points, 0)}
          </span>
        </div>
      </div>
    </div>
  );
};

export default ExamNavigation;
