import { api } from "./api/client";
import * as types from "@/types/exam-submission.types";

const BASE_URL = "/student/exams";

export const examSubmissionService = {
  // Get exam details with all problems
  getExamWithProblems: async (examID: number): Promise<types.ExamProblemsResponse> => {
    const response = await api.get(`${BASE_URL}/${examID}`);
    return response.data;
  },

  // Start exam - set started_at and get initial state
  startExam: async (examID: number): Promise<types.ExamTimerResponse> => {
    const response = await api.post(`${BASE_URL}/${examID}/start`, {});
    return response.data;
  },

  // Get remaining time for exam
  getTimeRemaining: async (examID: number): Promise<types.TimeRemainingResponse> => {
    const response = await api.get(`${BASE_URL}/${examID}/time-remaining`);
    return response.data;
  },

  // Get specific problem details
  getProblem: async (
    examID: number,
    problemID: number
  ): Promise<types.ExamProblemDetail> => {
    const response = await api.get(
      `${BASE_URL}/${examID}/problems/${problemID}`
    );
    return response.data;
  },

  // Submit code for a specific problem
  submitProblemCode: async (
    examID: number,
    problemID: number,
    req: { code: string; language?: string }
  ): Promise<types.ProblemSubmissionResult> => {
    const response = await api.post(
      `${BASE_URL}/${examID}/problems/${problemID}/submit`,
      {
        code: req.code,
        databaseType: req.language || 'postgresql',
      }
    );
    return response.data;
  },

  // Finish/submit entire exam
  finishExam: async (examID: number): Promise<types.ExamFinishResponse> => {
    const response = await api.post(`${BASE_URL}/submit`, {
      examID,
    });
    return response.data;
  },
};

export default examSubmissionService;
