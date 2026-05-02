import { JsonValue } from "./api.types"

export interface PDFUploadResponse {
  id: number
  status: string
  file_name: string
  created_at: string
  message?: string
}

export interface PDFStatusResponse {
  id: number
  status: string
  file_name: string
  extraction_result?: JsonValue
  error_message?: string
  created_at: string
  updated_at: string
}

export interface ExtractedProblem {
  id: number
  problem_number: number
  title: string
  description: string
  difficulty: 'easy' | 'medium' | 'hard'
  solution?: string
  test_case_count: number
  status: 'pending' | 'approved' | 'rejected' | 'editing'
  created_at: string
  updated_at: string
}

export interface ProblemSolution {
  solution_query: string
  db_type: 'postgresql' | 'mysql' | 'sqlserver' | 'sqlite'
}
