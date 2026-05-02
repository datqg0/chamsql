// Exam Submission related types
export interface ExamTimerResponse {
  participantID: number;
  examID: number;
  startedAt: string;
  timeRemainingMs: number;
  status: string;
}

export interface ProblemSubmissionRequest {
  code: string;
  databaseType: string;
}

export interface ProblemSubmissionResult {
  submissionID: number;
  examID: number;
  examProblemID: number;
  status: "accepted" | "wrong_answer" | "error";
  score: number;
  maxScore?: number;
  isCorrect: boolean;
  actualOutput?: string;
  expectedOutput?: string;
  errorMessage?: string;
  executionTimeMs?: number;
  attemptNumber: number;
  scoringMode?: string;
  feedback?: string;
}

export interface ExamProblemsResponse {
  examID: number;
  title: string;
  description?: string;
  startTime: string;
  endTime: string;
  durationMins: number;
  status: string;
  timeRemainingMs: number;
  participantStatus: string;
  problems: ExamProblemBrief[];
}

export interface ExamProblemBrief {
  examProblemID: number;
  problemID: number;
  title: string;
  difficulty: string;
  points: number;
  sortOrder: number;
}

export interface ExamProblemDetail {
  examProblemID: number;
  problemID: number;
  title: string;
  description: string;
  difficulty: string;
  points: number;
  sortOrder: number;
  initScript?: string;
  submissionStatus?: "answered" | "unanswered" | "skipped";
  submissions?: StudentSubmissionBrief[];
}

export interface StudentSubmissionBrief {
  submissionId: number;
  code: string;
  status: string;
  score: number;
  isCorrect: boolean;
  attemptNumber: number;
  executionTimeMs?: number;
  errorMessage?: string;
  submittedAt: string;
}

export type ExamFinishRequest = Record<string, never>

export interface ExamFinishResponse {
  participantID: number;
  examID: number;
  totalScore: number;
  status: string;
  submittedAt: string;
}

export interface ExamProblemSubmissionRequest {
  code: string;
  databaseType: string;
}

export interface TimeRemainingResponse {
  examID: number;
  timeRemainingMs: number;
  status: string;
  message?: string;
}
