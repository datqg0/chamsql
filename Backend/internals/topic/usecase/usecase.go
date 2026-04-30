package usecase

import (
	"context"
	"errors"

	"backend/internals/topic/controller/dto"
	"backend/internals/topic/repository"
	"backend/sql/models"
)

var (
	ErrTopicNotFound = errors.New("topic not found")
	ErrSlugExists    = errors.New("topic slug already exists")
)

type ITopicUseCase interface {
	Create(ctx context.Context, req *dto.CreateTopicRequest) (*dto.TopicResponse, error)
	GetByID(ctx context.Context, id int32) (*dto.TopicResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.TopicResponse, error)
	List(ctx context.Context) (*dto.TopicListResponse, error)
	Update(ctx context.Context, id int32, req *dto.UpdateTopicRequest) (*dto.TopicResponse, error)
	Delete(ctx context.Context, id int32) error
	GetTree(ctx context.Context) (*dto.TopicTreeResponse, error)
}

type topicUseCase struct {
	repo repository.ITopicRepository
}

func NewTopicUseCase(repo repository.ITopicRepository) ITopicUseCase {
	return &topicUseCase{repo: repo}
}

func (u *topicUseCase) Create(ctx context.Context, req *dto.CreateTopicRequest) (*dto.TopicResponse, error) {
	// Check if slug exists
	_, err := u.repo.GetBySlug(ctx, req.Slug)
	if err == nil {
		return nil, ErrSlugExists
	}

	topic, err := u.repo.Create(ctx, models.CreateTopicParams{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: strPtr(req.Description),
		Icon:        strPtr(req.Icon),
		SortOrder:   int32Ptr(req.SortOrder),
		ParentID:    req.ParentID,
		Level:       int32Ptr(req.Level),
	})
	if err != nil {
		return nil, err
	}

	return toTopicResponse(topic), nil
}

func (u *topicUseCase) GetByID(ctx context.Context, id int32) (*dto.TopicResponse, error) {
	topic, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTopicNotFound
	}
	return toTopicResponse(topic), nil
}

func (u *topicUseCase) GetBySlug(ctx context.Context, slug string) (*dto.TopicResponse, error) {
	topic, err := u.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrTopicNotFound
	}
	return toTopicResponse(topic), nil
}

func (u *topicUseCase) List(ctx context.Context) (*dto.TopicListResponse, error) {
	// Get topics with problem count
	topicsWithCount, err := u.repo.CountProblemsPerTopic(ctx)
	if err != nil {
		return nil, err
	}

	topics := make([]dto.TopicResponse, len(topicsWithCount))
	for i, t := range topicsWithCount {
		topics[i] = dto.TopicResponse{
			ID:           t.ID,
			Name:         t.Name,
			Slug:         t.Slug,
			ProblemCount: t.ProblemCount,
		}
	}

	return &dto.TopicListResponse{
		Topics: topics,
		Total:  int64(len(topics)),
	}, nil
}

func (u *topicUseCase) Update(ctx context.Context, id int32, req *dto.UpdateTopicRequest) (*dto.TopicResponse, error) {
	// Check if topic exists
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTopicNotFound
	}

	params := models.UpdateTopicParams{ID: id}
	if req.Name != nil {
		params.Name = req.Name
	}
	if req.Description != nil {
		params.Description = req.Description
	}
	if req.Icon != nil {
		params.Icon = req.Icon
	}
	if req.SortOrder != nil {
		sortOrder := int32(*req.SortOrder)
		params.SortOrder = &sortOrder
	}
	if req.ParentID != nil {
		params.ParentID = req.ParentID
	}
	if req.Level != nil {
		level := int32(*req.Level)
		params.Level = &level
	}

	topic, err := u.repo.Update(ctx, id, params)
	if err != nil {
		return nil, err
	}

	return toTopicResponse(topic), nil
}

func (u *topicUseCase) Delete(ctx context.Context, id int32) error {
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTopicNotFound
	}
	return u.repo.Delete(ctx, id)
}

func (u *topicUseCase) GetTree(ctx context.Context) (*dto.TopicTreeResponse, error) {
	rows, err := u.repo.GetTopicTree(ctx)
	if err != nil {
		return nil, err
	}

	// Map to store nodes by ID for quick lookup (pointers)
	nodeMap := make(map[int32]*dto.TopicTreeNode)
	
	// First pass: create all node pointers and store them in the map
	for _, row := range rows {
		node := &dto.TopicTreeNode{
			ID:           row.ID,
			Name:         row.Name,
			Slug:         row.Slug,
			Description:  ptrToStr(row.Description),
			Icon:         ptrToStr(row.Icon),
			SortOrder:    int(ptrToInt32(row.SortOrder)),
			Level:        int(ptrToInt32(row.Level)),
			ProblemCount: row.ProblemCount,
			Children:     []*dto.TopicTreeNode{},
		}
		nodeMap[row.ID] = node
	}

	// Second pass: link children to their parents
	var roots []dto.TopicTreeNode
	for _, row := range rows {
		node := nodeMap[row.ID]
		if row.ParentID == nil || *row.ParentID == 0 {
			// It's a root node, add its value to the roots slice at the end
			// We can't add it yet if we want it to have all children populated
		} else {
			// Find the parent and add this node pointer as a child
			if parent, ok := nodeMap[*row.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// Third pass: collect roots
	for _, row := range rows {
		if row.ParentID == nil || *row.ParentID == 0 {
			if node, ok := nodeMap[row.ID]; ok {
				roots = append(roots, *node)
			}
		}
	}

	return &dto.TopicTreeResponse{
		Tree: roots,
	}, nil
}

// Helper functions
func toTopicResponse(t *models.Topic) *dto.TopicResponse {
	return &dto.TopicResponse{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Description: ptrToStr(t.Description),
		Icon:        ptrToStr(t.Icon),
		SortOrder:   int(ptrToInt32(t.SortOrder)),
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int32Ptr(i int) *int32 {
	v := int32(i)
	return &v
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
