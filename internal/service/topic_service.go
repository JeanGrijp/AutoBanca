package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type TopicService struct {
	q db.Querier
}

func NewTopicService(q db.Querier) *TopicService {
	return &TopicService{
		q: q,
	}
}

func (s *TopicService) CreateTopic(ctx context.Context, name string, subjectID pgtype.UUID) (db.Topic, error) {
	row, err := s.q.CreateTopic(ctx, db.CreateTopicParams{
		Name:      name,
		SubjectID: subjectID,
	})
	if err != nil {
		return db.Topic{}, err
	}
	return row, nil
}

func (s *TopicService) ListTopics(ctx context.Context) ([]db.Topic, error) {
	return s.q.ListTopics(ctx)
}

func (s *TopicService) GetTopic(ctx context.Context, id pgtype.UUID) (db.Topic, error) {
	return s.q.GetTopic(ctx, id)
}

func (s *TopicService) UpdateTopic(ctx context.Context, id pgtype.UUID, name string, subjectID pgtype.UUID) (db.Topic, error) {
	arg := db.UpdateTopicParams{
		ID:        id,
		Name:      name,
		SubjectID: subjectID,
	}
	return s.q.UpdateTopic(ctx, arg)
}

func (s *TopicService) DeleteTopic(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteTopic(ctx, id)
}

func (s *TopicService) ListTopicsBySubject(ctx context.Context, subjectID pgtype.UUID) ([]db.Topic, error) {
	return s.q.ListTopicsBySubject(ctx, subjectID)
}
