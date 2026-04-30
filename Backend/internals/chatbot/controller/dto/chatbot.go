package dto

import "encoding/json"

// ChatRequest represents a student asking for help
type ChatRequest struct {
	Message       string          `json:"message" binding:"required,min=3"`
	ProblemID     *int64          `json:"problemId"`
	ProblemTitle  string          `json:"problemTitle"`
	ProblemDesc   string          `json:"problemDescription"`
	StudentSQL    string          `json:"studentSql"`
	ErrorMessage  string          `json:"errorMessage"`
	Context       json.RawMessage `json:"context"`        // Additional context from frontend
	ConversationID string        `json:"conversationId"`  // For multi-turn conversation
}

// ChatResponse represents the chatbot's reply
type ChatResponse struct {
	Reply          string   `json:"reply"`
	Suggestions    []string `json:"suggestions,omitempty"`
	Hints          []string `json:"hints,omitempty"`
	ConversationID string   `json:"conversationId"`
	Provider       string   `json:"provider"` // huggingface, pattern
	ResponseTimeMs int64    `json:"responseTimeMs"`
}
