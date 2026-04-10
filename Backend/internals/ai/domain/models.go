package domain

import "encoding/json"

// AIGenerationRequest is the request for AI generation
type AIGenerationRequest struct {
	ContentType string `json:"content_type"` // solution, test_case
	Description string `json:"description"`
	SchemaSQL   string `json:"schema_sql"`
	SolutionSQL string `json:"solution_sql,omitempty"`
	Mode        string `json:"mode"` // hybrid, pattern_only, llm_only
}

// AIGenerationResponse is the response from AI generation
type AIGenerationResponse struct {
	GeneratedContent string  `json:"generated_content"`
	ConfidenceScore  float64 `json:"confidence_score"` // 0-1
	AIProvider       string  `json:"ai_provider"`      // pattern, huggingface
	Error            string  `json:"error,omitempty"`
}

// SolutionGenerationInput input for solution generation
type SolutionGenerationInput struct {
	ProblemDescription string
	SchemaSQL          string
	Examples           []string // Example test cases
}

// TestCaseGenerationInput input for test case generation
type TestCaseGenerationInput struct {
	SchemaSQL           string
	SolutionSQL         string
	Description         string
	DifficultyLevel     string // easy, medium, hard
	PublicTestCaseCount int    // How many public test cases
	HiddenTestCaseCount int    // How many hidden test cases
}

// TestCaseGenerated represents generated test case
type TestCaseGenerated struct {
	TestNumber     int             `json:"test_number"`
	Description    string          `json:"description"`
	TestDataSQL    string          `json:"test_data_sql"`
	ExpectedOutput json.RawMessage `json:"expected_output"`
	IsPublic       bool            `json:"is_public"`
	Difficulty     string          `json:"difficulty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	TestNumber  int    `json:"test_number"`
	ErrorType   string `json:"error_type"` // syntax, execution, mismatch
	Message     string `json:"message"`
	ActualValue string `json:"actual_value,omitempty"`
	Expected    string `json:"expected,omitempty"`
}

// ValidationResult represents test validation result
type ValidationResult struct {
	IsValid         bool     `json:"is_valid"`
	PassedCount     int      `json:"passed_count"`
	TotalCount      int      `json:"total_count"`
	Errors          []string `json:"errors,omitempty"`
	ExecutionTimeMS int      `json:"execution_time_ms"`
}
