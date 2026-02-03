package repository

import (
	"context"

	"backend/db"
	"backend/sql/models"
)

type ITopicRepository interface {
	Create(ctx context.Context, params models.CreateTopicParams) (*models.Topic, error)
	GetByID(ctx context.Context, id int32) (*models.Topic, error)
	GetBySlug(ctx context.Context, slug string) (*models.Topic, error)
	List(ctx context.Context) ([]models.Topic, error)
	Update(ctx context.Context, id int32, params models.UpdateTopicParams) (*models.Topic, error)
	Delete(ctx context.Context, id int32) error
	CountProblemsPerTopic(ctx context.Context) ([]models.CountProblemsPerTopicRow, error)
}

type topicRepository struct {
	db      *db.Database
	queries *models.Queries
}

func NewTopicRepository(database *db.Database) ITopicRepository {
	return &topicRepository{
		db:      database,
		queries: models.New(database.GetPool()),
	}
}

func (r *topicRepository) Create(ctx context.Context, params models.CreateTopicParams) (*models.Topic, error) {
	topic, err := r.queries.CreateTopic(ctx, params)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) GetByID(ctx context.Context, id int32) (*models.Topic, error) {
	topic, err := r.queries.GetTopicByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) GetBySlug(ctx context.Context, slug string) (*models.Topic, error) {
	topic, err := r.queries.GetTopicBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) List(ctx context.Context) ([]models.Topic, error) {
	return r.queries.ListTopics(ctx)
}

func (r *topicRepository) Update(ctx context.Context, id int32, params models.UpdateTopicParams) (*models.Topic, error) {
	params.ID = id
	topic, err := r.queries.UpdateTopic(ctx, params)
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *topicRepository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteTopic(ctx, id)
}

func (r *topicRepository) CountProblemsPerTopic(ctx context.Context) ([]models.CountProblemsPerTopicRow, error) {
	return r.queries.CountProblemsPerTopic(ctx)
}
