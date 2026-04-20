// Exam Submission related types
export interface ExamTimerResponse {
  examID: number;
  userID: number;
  startedAt: string;
  endTime: string;
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
  maxScore: number;
  isCorrect: boolean;
  actualOutput?: string;
  expectedOutput?: string;
  errorMessage?: string;
  executionTimeMs: number;
  attemptNumber: number;
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
  problems: ExamProblemDetail[];
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
  solutionQuery?: string;
  submissionStatus?: "answered" | "unanswered" | "skipped";
}

export interface ExamFinishRequest {
  // no body needed
}

export interface ExamFinishResponse {
  examID: number;
  userID: number;
  totalScore: number;
  maxScore: number;
  status: string;
  submittedAt: string;
}

export interface ExamProblemSubmissionRequest {
  code: string;
  databaseType: string;
}

export interface TimeRemainingResponse {
  examID: number;
  userID: number;
  startedAt: string;
  endTime: string;
  timeRemainingMs: number;
  isExpired: boolean;
}
