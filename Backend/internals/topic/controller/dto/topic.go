package dto

type CreateTopicRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Slug        string `json:"slug" binding:"required,min=2,max=100,alphanum"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Icon        string `json:"icon" binding:"omitempty,max=50"`
	SortOrder   int    `json:"sortOrder" binding:"omitempty,min=0"`
}

type UpdateTopicRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Icon        *string `json:"icon" binding:"omitempty,max=50"`
	SortOrder   *int    `json:"sortOrder" binding:"omitempty,min=0"`
}

type TopicResponse struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description,omitempty"`
	Icon         string `json:"icon,omitempty"`
	SortOrder    int    `json:"sortOrder"`
	ProblemCount int64  `json:"problemCount,omitempty"`
}

type TopicListResponse struct {
	Topics []TopicResponse `json:"topics"`
	Total  int64           `json:"total"`
}
