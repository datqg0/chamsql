import type * as raw from "@/types/api-raw.types";
import type * as types from "@/types/exam-submission.types";

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

const mapExamProblemBrief = (p: raw.RawExamProblemBrief): types.ExamProblemBrief => ({
  examProblemID: Number(p.exam_problem_id ?? p.examProblemID ?? 0),
  problemID: Number(p.problem_id ?? p.problemID ?? 0),
  title: p.title ?? "",
  difficulty: p.difficulty ?? "",
  points: Number(p.points ?? 0),
  sortOrder: Number(p.sort_order ?? p.sortOrder ?? 0),
});

const mapExamProblemDetail = (p: raw.RawExamProblemDetail): types.ExamProblemDetail => ({
  examProblemID: Number(p.exam_problem_id ?? p.examProblemID ?? 0),
  problemID: Number(p.problem_id ?? p.problemID ?? 0),
  title: p.title ?? "",
  description: p.description ?? "",
  difficulty: p.difficulty ?? "",
  points: Number(p.points ?? 0),
  sortOrder: Number(p.sort_order ?? p.sortOrder ?? 0),
  initScript: p.init_script ?? p.initScript,
});

const mapSubmitResult = (p: raw.RawSubmitResult): types.ProblemSubmissionResult => ({
  submissionID: Number(p.submission_id ?? p.submissionID ?? 0),
  examID: Number(p.exam_id ?? p.examID ?? 0),
  examProblemID: Number(p.exam_problem_id ?? p.examProblemID ?? 0),
  status: (p.status ?? "error") as types.ProblemSubmissionResult["status"],
  score: Number(p.score ?? 0),
  isCorrect: Boolean(p.is_correct ?? p.isCorrect),
  actualOutput: p.actual_output ?? p.actualOutput,
  expectedOutput: p.expected_output ?? p.expectedOutput,
  errorMessage: p.error_message ?? p.errorMessage ?? undefined,
  executionTimeMs: p.execution_time_ms ?? p.executionTimeMs ?? undefined,
  attemptNumber: Number(p.attempt_number ?? p.attemptNumber ?? 0),
  scoringMode: p.scoring_mode ?? p.scoringMode,
});

export const examSubmissionService = {
  // Get exam details with all problems
  getExamWithProblems: async (examID: number): Promise<types.ExamProblemsResponse> => {
    const response = await api.get(`${BASE_URL}/${examID}`);
    const payload = unwrapPayload<raw.RawExamResponse>(response.data);
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
    const payload = unwrapPayload<raw.RawJoinExamResponse>(response.data);
    return {
      participantID: Number(payload.participant_id ?? payload.participantID ?? 0),
      status: payload.status ?? "registered",
    };
  },

  // Start exam - set started_at and get initial state
  startExam: async (examID: number): Promise<types.ExamTimerResponse> => {
    const response = await api.post(`${BASE_URL}/start`, { exam_id: examID });
    const payload = unwrapPayload<raw.RawStartExamResponse>(response.data);
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
    const payload = unwrapPayload<raw.RawTimeRemainingResponse>(response.data);
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
    const payload = unwrapPayload<raw.RawExamProblemDetail>(response.data);
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
    return mapSubmitResult(unwrapPayload<raw.RawSubmitResult>(response.data));
  },

  // Finish/submit entire exam
  finishExam: async (examID: number): Promise<types.ExamFinishResponse> => {
    const response = await api.post(`${BASE_URL}/submit`, {
      exam_id: examID,
    });
    const payload = unwrapPayload<raw.RawSubmitExamResponse>(response.data);
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
