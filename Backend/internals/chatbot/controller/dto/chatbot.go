package dto

// ChatRequest represents a student question
type ChatRequest struct {
    Message        string `json:"message" binding:"required,min=2"`
    ProblemID      *int64 `json:"problemId"`
    ProblemTitle   string `json:"problemTitle"`
    StudentSQL     string `json:"studentSql"`
    ErrorMessage   string `json:"errorMessage"`
    ConversationID string `json:"conversationId"`
    UserID         *int64 `json:"-"` // set bởi handler từ JWT, không nhận từ client
}

// ChatResponse represents chatbot reply
type ChatResponse struct {
    Reply          string   `json:"reply"`
    ConversationID string   `json:"conversationId"`
    Provider       string   `json:"provider"`
    ToolsUsed      []string `json:"toolsUsed,omitempty"`
    ResponseTimeMs int64    `json:"responseTimeMs"`
}
