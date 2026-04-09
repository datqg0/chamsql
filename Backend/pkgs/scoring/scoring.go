package scoring

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ScoringMode string

const (
	ScoringModeAuto      ScoringMode = "auto"
	ScoringModeAnswerKey ScoringMode = "answer_key"
	ScoringModeManual    ScoringMode = "manual"
)

type GradingRequest struct {
	SubmissionID     int64       `json:"submission_id"`
	ScoringMode      ScoringMode `json:"scoring_mode"`
	ActualOutput     []byte      `json:"actual_output"`
	ExpectedOutput   []byte      `json:"expected_output"`
	StudentAnswer    *string     `json:"student_answer"`
	ReferenceAnswer  *string     `json:"reference_answer"`
	MaxPoints        float64     `json:"max_points"`
	ErrorMessage     *string     `json:"error_message"`
	SubmissionStatus string      `json:"submission_status"`
}

type GradingResult struct {
	Score             float64 `json:"score"`
	IsCorrect         bool    `json:"is_correct"`
	ComparisonDetails string  `json:"comparison_details"`
	ScoringMode       string  `json:"scoring_mode"`
}

func Score(request *GradingRequest) (*GradingResult, error) {
	if request == nil {
		return nil, fmt.Errorf("grading request cannot be nil")
	}

	switch request.ScoringMode {
	case ScoringModeAuto:
		return scoreAuto(request)
	case ScoringModeAnswerKey:
		return scoreAnswerKey(request)
	case ScoringModeManual:
		return scoreManual(request)
	default:
		return nil, fmt.Errorf("unsupported scoring mode: %s", request.ScoringMode)
	}
}

func scoreAuto(request *GradingRequest) (*GradingResult, error) {
	if request.SubmissionStatus == "error" || request.SubmissionStatus == "timeout" {
		reason := "Submission failed"
		if request.ErrorMessage != nil {
			reason = *request.ErrorMessage
		}
		return &GradingResult{
			Score:             0,
			IsCorrect:         false,
			ComparisonDetails: reason,
			ScoringMode:       string(ScoringModeAuto),
		}, nil
	}

	isCorrect, details := compareJSONOutputs(request.ActualOutput, request.ExpectedOutput)

	score := 0.0
	if isCorrect {
		score = request.MaxPoints
	}

	if !isCorrect {
		details = fmt.Sprintf("Output mismatch: %s. Score: 0 / %.2f", details, request.MaxPoints)
	} else {
		details = fmt.Sprintf("Output matches. Score: %.2f / %.2f", score, request.MaxPoints)
	}

	return &GradingResult{
		Score:             score,
		IsCorrect:         isCorrect,
		ComparisonDetails: details,
		ScoringMode:       string(ScoringModeAuto),
	}, nil
}

func scoreAnswerKey(request *GradingRequest) (*GradingResult, error) {
	if request.StudentAnswer == nil || request.ReferenceAnswer == nil {
		return nil, fmt.Errorf("student answer and reference answer required for answer-key scoring")
	}

	if request.SubmissionStatus == "error" {
		errorMsg := "Submission error"
		if request.ErrorMessage != nil {
			errorMsg = *request.ErrorMessage
		}
		return &GradingResult{
			Score:             0,
			IsCorrect:         false,
			ComparisonDetails: fmt.Sprintf("Cannot score: %s", errorMsg),
			ScoringMode:       string(ScoringModeAnswerKey),
		}, nil
	}

	isCorrect := compareAnswers(*request.StudentAnswer, *request.ReferenceAnswer)

	score := 0.0
	if isCorrect {
		score = request.MaxPoints
	}

	details := "Answer does not match"
	if isCorrect {
		details = fmt.Sprintf("Answer matches. Score: %.2f / %.2f", score, request.MaxPoints)
	} else {
		details = fmt.Sprintf("Expected: %s\nGot: %s\nScore: 0 / %.2f",
			*request.ReferenceAnswer, *request.StudentAnswer, request.MaxPoints)
	}

	return &GradingResult{
		Score:             score,
		IsCorrect:         isCorrect,
		ComparisonDetails: details,
		ScoringMode:       string(ScoringModeAnswerKey),
	}, nil
}

func scoreManual(request *GradingRequest) (*GradingResult, error) {
	details := fmt.Sprintf("Manual grading required. Status: %s", request.SubmissionStatus)
	if request.ErrorMessage != nil {
		details = fmt.Sprintf("Manual grading required. Status: %s. Error: %s",
			request.SubmissionStatus, *request.ErrorMessage)
	}

	return &GradingResult{
		Score:             0,
		IsCorrect:         false,
		ComparisonDetails: details,
		ScoringMode:       string(ScoringModeManual),
	}, nil
}

func compareJSONOutputs(actual, expected []byte) (bool, string) {
	var actualRows, expectedRows []map[string]interface{}

	if err := json.Unmarshal(actual, &actualRows); err != nil {
		return false, fmt.Sprintf("Failed to parse actual output: %v", err)
	}

	if err := json.Unmarshal(expected, &expectedRows); err != nil {
		return false, fmt.Sprintf("Failed to parse expected output: %v", err)
	}

	if len(actualRows) != len(expectedRows) {
		return false, fmt.Sprintf("Row count mismatch: got %d, expected %d", len(actualRows), len(expectedRows))
	}

	for i, expectedRow := range expectedRows {
		actualRow := actualRows[i]

		if len(actualRow) != len(expectedRow) {
			return false, fmt.Sprintf("Column count mismatch in row %d: got %d, expected %d",
				i, len(actualRow), len(expectedRow))
		}

		for key, expectedVal := range expectedRow {
			actualVal, exists := actualRow[key]
			if !exists {
				return false, fmt.Sprintf("Missing column '%s' in row %d", key, i)
			}

			expectedStr := normalizeAnswer(fmt.Sprintf("%v", expectedVal))
			actualStr := normalizeAnswer(fmt.Sprintf("%v", actualVal))

			if expectedStr != actualStr {
				return false, fmt.Sprintf("Value mismatch in row %d, column '%s': got '%s', expected '%s'",
					i, key, actualVal, expectedVal)
			}
		}
	}

	return true, "All rows and columns match"
}

func compareAnswers(student, reference string) bool {
	return normalizeAnswer(student) == normalizeAnswer(reference)
}

func normalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ToLower(answer)
	answer = strings.Join(strings.Fields(answer), " ")
	return answer
}
