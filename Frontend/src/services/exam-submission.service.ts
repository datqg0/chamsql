import * as types from "@/types/exam-submission.types";

import { api } from "./api/client";

const BASE_URL = "/student/exams";

const unwrapPayload = <T>(payload: unknown): T => {
  if (
    payload &&
    typeof payload === "object" &&
    "success" in payload &&
    "data" in payload
  ) {
    return payload.data as T;
  }
  return payload as T;
};

const mapExamProblemBrief = (problem: unknown): types.ExamProblemBrief => ({
  examProblemID: Number(problem.exam_problem_id ?? problem.examProblemID ?? 0),
  problemID: Number(problem.problem_id ?? problem.problemID ?? 0),
  title: problem.title ?? "",
  difficulty: problem.difficulty ?? "",
  points: Number(problem.points ?? 0),
  sortOrder: Number(problem.sort_order ?? problem.sortOrder ?? 0),
});

const mapExamProblemDetail = (problem: unknown): types.ExamProblemDetail => ({
  examProblemID: Number(problem.exam_problem_id ?? problem.examProblemID ?? 0),
  problemID: Number(problem.problem_id ?? problem.problemID ?? 0),
  title: problem.title ?? "",
  description: problem.description ?? "",
  difficulty: problem.difficulty ?? "",
  points: Number(problem.points ?? 0),
  sortOrder: Number(problem.sort_order ?? problem.sortOrder ?? 0),
  initScript: problem.init_script ?? problem.initScript,
  solutionQuery: problem.solution_query ?? problem.solutionQuery,
});

const mapSubmitResult = (payload: unknown): types.ProblemSubmissionResult => ({
  submissionID: Number(payload.submission_id ?? payload.submissionID ?? 0),
  examID: Number(payload.exam_id ?? payload.examID ?? 0),
  examProblemID: Number(payload.exam_problem_id ?? payload.examProblemID ?? 0),
  status: payload.status ?? "error",
  score: Number(payload.score ?? 0),
  isCorrect: Boolean(payload.is_correct ?? payload.isCorrect),
  actualOutput: payload.actual_output ?? payload.actualOutput,
  expectedOutput: payload.expected_output ?? payload.expectedOutput,
  errorMessage: payload.error_message ?? payload.errorMessage,
  executionTimeMs: payload.execution_time_ms ?? payload.executionTimeMs,
  attemptNumber: Number(payload.attempt_number ?? payload.attemptNumber ?? 0),
  scoringMode: payload.scoring_mode ?? payload.scoringMode,
});

export const examSubmissionService = {
  // Get exam details with all problems
  getExamWithProblems: async (examID: number): Promise<types.ExamProblemsResponse> => {
    const response = await api.get(`${BASE_URL}/${examID}`);
    const payload = unwrapPayload<unknown>(response.data);
    return {
      examID: Number(payload.exam_id ?? payload.examID ?? 0),
      title: payload.title ?? "",
      description: payload.description,
      startTime: payload.start_time ?? payload.startTime ?? "",
      endTime: payload.end_time ?? payload.endTime ?? "",
      durationMins: Number(payload.duration_minutes ?? payload.durationMins ?? 0),
      status: payload.status ?? "",
      timeRemainingMs: Number(payload.time_remaining_ms ?? payload.timeRemainingMs ?? 0),
      participantStatus: payload.participant_status ?? payload.participantStatus ?? "",
      problems: Array.isArray(payload.problems)
        ? payload.problems.map(mapExamProblemBrief)
        : [],
    };
  },

  // Join exam - explicitly register participant
  joinExam: async (examID: number): Promise<{ participantID: number; status: string }> => {
    const response = await api.post(`${BASE_URL}/join`, { exam_id: examID });
    const payload = unwrapPayload<unknown>(response.data);
    return {
      participantID: Number(payload.participant_id ?? payload.participantID ?? 0),
      status: payload.status ?? "registered",
    };
  },

  // Start exam - set started_at and get initial state
  startExam: async (examID: number): Promise<types.ExamTimerResponse> => {
    const response = await api.post(`${BASE_URL}/start`, { exam_id: examID });
    const payload = unwrapPayload<unknown>(response.data);
    return {
      participantID: Number(payload.participant_id ?? payload.participantID ?? 0),
      examID: Number(payload.exam_id ?? payload.examID ?? examID),
      startedAt: payload.started_at ?? payload.startedAt ?? "",
      timeRemainingMs: Number(payload.time_remaining_ms ?? payload.timeRemainingMs ?? 0),
      status: payload.status ?? "",
    };
  },

  // Get remaining time for exam
  getTimeRemaining: async (examID: number): Promise<types.TimeRemainingResponse> => {
    const response = await api.get(`${BASE_URL}/${examID}/time-remaining`);
    const payload = unwrapPayload<unknown>(response.data);
    return {
      examID: Number(payload.exam_id ?? payload.examID ?? examID),
      timeRemainingMs: Number(payload.time_remaining_ms ?? payload.timeRemainingMs ?? 0),
      status: payload.status ?? "",
      message: payload.message,
    };
  },

  // Get specific problem details
  getProblem: async (
    examID: number,
    problemID: number
  ): Promise<types.ExamProblemDetail> => {
    const response = await api.get(
      `${BASE_URL}/${examID}/problems/${problemID}`
    );
    const payload = unwrapPayload<unknown>(response.data);
    return mapExamProblemDetail(payload);
  },

  // Submit code for a specific problem
  submitProblemCode: async (
    examID: number,
    problemID: number,
    req: { code: string; databaseType?: string }
  ): Promise<types.ProblemSubmissionResult> => {
    const response = await api.post(
      `${BASE_URL}/${examID}/problems/${problemID}/submit`,
      {
        code: req.code,
        database_type: req.databaseType || "postgresql",
      }
    );
    return mapSubmitResult(unwrapPayload<unknown>(response.data));
  },

  // Finish/submit entire exam
  finishExam: async (examID: number): Promise<types.ExamFinishResponse> => {
    const response = await api.post(`${BASE_URL}/submit`, {
      exam_id: examID,
    });
    const payload = unwrapPayload<unknown>(response.data);
    return {
      participantID: Number(payload.participant_id ?? payload.participantID ?? 0),
      examID: Number(payload.exam_id ?? payload.examID ?? examID),
      totalScore: Number(payload.total_score ?? payload.totalScore ?? 0),
      status: payload.status ?? "",
      submittedAt: payload.submitted_at ?? payload.submittedAt ?? "",
    };
  },
};

export default examSubmissionService;
