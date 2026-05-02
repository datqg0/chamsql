/** Raw snake_case shape from backend GET /student/exams/:id */
export interface RawExamResponse {
  exam_id?: number
  examID?: number
  title?: string
  description?: string
  start_time?: string
  startTime?: string
  end_time?: string
  endTime?: string
  duration_minutes?: number
  durationMins?: number
  status?: string
  time_remaining_ms?: number
  timeRemainingMs?: number
  participant_status?: string
  participantStatus?: string
  problems?: RawExamProblemBrief[]
}

export interface RawExamProblemBrief {
  exam_problem_id?: number
  examProblemID?: number
  problem_id?: number
  problemID?: number
  title?: string
  difficulty?: string
  points?: number | null
  sort_order?: number | null
  sortOrder?: number | null
}

export interface RawExamProblemDetail extends RawExamProblemBrief {
  description?: string
  init_script?: string
  initScript?: string
  // NO solution_query for students
  attempt_number?: number
  attemptNumber?: number
  submissions?: RawStudentSubmission[]
}

export interface RawStudentSubmission {
  submission_id?: number
  submissionID?: number
  code?: string
  status?: string
  score?: number
  is_correct?: boolean
  isCorrect?: boolean
  attempt_number?: number
  attemptNumber?: number
  execution_time_ms?: number | null
  executionTimeMs?: number | null
  error_message?: string | null
  errorMessage?: string | null
  submitted_at?: string
  submittedAt?: string
}

export interface RawSubmitResult {
  submission_id?: number
  submissionID?: number
  exam_id?: number
  examID?: number
  exam_problem_id?: number
  examProblemID?: number
  status?: string
  score?: number
  is_correct?: boolean
  isCorrect?: boolean
  actual_output?: string
  actualOutput?: string
  expected_output?: string
  expectedOutput?: string
  error_message?: string | null
  errorMessage?: string | null
  execution_time_ms?: number | null
  executionTimeMs?: number | null
  attempt_number?: number
  attemptNumber?: number
  scoring_mode?: string
  scoringMode?: string
}

export interface RawJoinExamResponse {
  participant_id?: number
  participantID?: number
  status?: string
}

export interface RawStartExamResponse {
  participant_id?: number
  participantID?: number
  exam_id?: number
  examID?: number
  started_at?: string
  startedAt?: string
  time_remaining_ms?: number
  timeRemainingMs?: number
  status?: string
}

export interface RawTimeRemainingResponse {
  exam_id?: number
  examID?: number
  time_remaining_ms?: number
  timeRemainingMs?: number
  status?: string
  message?: string
}

export interface RawSubmitExamResponse {
  participant_id?: number
  participantID?: number
  exam_id?: number
  examID?: number
  total_score?: number
  totalScore?: number
  status?: string
  submitted_at?: string
  submittedAt?: string
}
