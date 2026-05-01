package dto

import "encoding/json"

type CreateProblemRequest struct {
	Title              string          `json:"title" binding:"required,min=3,max=255"`
	Slug               string          `json:"slug" binding:"required,min=3,max=255"`
	Description        string          `json:"description" binding:"required,min=10"`
	Difficulty         string          `json:"difficulty" binding:"required,oneof=easy medium hard"`
	TopicID            *int32          `json:"topicId" binding:"omitempty"`
	InitScript         string          `json:"initScript" binding:"omitempty"`
	SolutionQuery      string          `json:"solutionQuery" binding:"omitempty"`
	SupportedDatabases []string        `json:"supportedDatabases" binding:"required,min=1"`
	OrderMatters       bool            `json:"orderMatters"`
	Hints              json.RawMessage `json:"hints" binding:"omitempty"`
	SampleOutput       json.RawMessage `json:"sampleOutput" binding:"omitempty"`
	IsPublic           bool            `json:"isPublic"`
	TestCases          []TestCaseRequest `json:"testCases" binding:"omitempty"`
}

type TestCaseRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	InitScript    string `json:"initScript" binding:"required"`
	SolutionQuery string `json:"solutionQuery" binding:"required"`
	Weight        int32  `json:"weight" binding:"required,min=1"`
	IsHidden      bool   `json:"isHidden"`
}

type UpdateProblemRequest struct {
	Title         *string         `json:"title" binding:"omitempty,min=3,max=255"`
	Description   *string         `json:"description" binding:"omitempty,min=10"`
	Difficulty    *string         `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	TopicID       *int32          `json:"topicId" binding:"omitempty"`
	InitScript    *string         `json:"initScript" binding:"omitempty"`
	SolutionQuery *string         `json:"solutionQuery" binding:"omitempty"`
	OrderMatters  *bool           `json:"orderMatters" binding:"omitempty"`
	Hints         json.RawMessage `json:"hints" binding:"omitempty"`
	SampleOutput  json.RawMessage `json:"sampleOutput" binding:"omitempty"`
	IsPublic      *bool           `json:"isPublic" binding:"omitempty"`
	TestCases     []TestCaseRequest `json:"testCases" binding:"omitempty"`
}

type ProblemResponse struct {
	ID                 int64              `json:"id"`
	Title              string             `json:"title"`
	Slug               string             `json:"slug"`
	Description        string             `json:"description"`
	Difficulty         string             `json:"difficulty"`
	TopicID            *int32             `json:"topicId,omitempty"`
	TopicName          string             `json:"topicName,omitempty"`
	TopicSlug          string             `json:"topicSlug,omitempty"`
	InitScript         string             `json:"initScript,omitempty"`
	SolutionQuery      string             `json:"solutionQuery,omitempty"`
	SupportedDatabases []string           `json:"supportedDatabases"`
	OrderMatters       bool               `json:"orderMatters"`
	Hints              json.RawMessage    `json:"hints,omitempty"`
	SampleOutput       json.RawMessage    `json:"sampleOutput,omitempty"`
	IsPublic           bool               `json:"isPublic"`
	CreatedBy          *int64             `json:"createdBy,omitempty"`
	SourcePdfUrl       *string            `json:"sourcePdfUrl,omitempty"`
	CreatedAt          string             `json:"createdAt,omitempty"`
	TestCases          []TestCaseResponse `json:"testCases,omitempty"`
	// User progress (if authenticated)
	IsSolved   *bool `json:"isSolved,omitempty"`
	Attempts   *int  `json:"attempts,omitempty"`
	BestTimeMs *int  `json:"bestTimeMs,omitempty"`
}

type TestCaseResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	InitScript    string `json:"initScript"`
	SolutionQuery string `json:"solutionQuery"`
	Weight        int32  `json:"weight"`
	IsHidden      bool   `json:"isHidden"`
}

type ProblemListResponse struct {
	Problems interface{} `json:"problems"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

type ProblemListQuery struct {
	TopicID    *int32  `form:"topicId"`
	Difficulty *string `form:"difficulty"`
	Search     *string `form:"search"`
	Page       int     `form:"page,default=1"`
	PageSize   int     `form:"pageSize,default=20"`
}

type ProblemSummary struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Difficulty string `json:"difficulty"`
	IsPublic   bool   `json:"isPublic"`
}
