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
