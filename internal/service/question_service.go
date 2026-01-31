package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type QuestionService struct {
	svc db.Querier
}

type QuestionFilter struct {
	TopicID      *pgtype.UUID
	Position     *pgtype.Text
	Level        *pgtype.Text
	Difficulty   *pgtype.Text
	Modality     *pgtype.Text
	PracticeArea *pgtype.Text
	FieldOfStudy *pgtype.Text
}

func NewQuestionService(svc db.Querier) *QuestionService {
	return &QuestionService{svc: svc}
}

func (s *QuestionService) CreateQuestion(ctx context.Context, question db.Question) (db.Question, error) {
	row, err := s.svc.CreateQuestion(ctx, db.CreateQuestionParams{
		Statement:    question.Statement,
		Year:         question.Year,
		TopicID:      question.TopicID,
		Position:     question.Position,
		Level:        question.Level,
		Difficulty:   question.Difficulty,
		Modality:     question.Modality,
		PracticeArea: question.PracticeArea,
		FieldOfStudy: question.FieldOfStudy,
	})
	if err != nil {
		return db.Question{}, err
	}
	return row, nil
}

func (s *QuestionService) ListQuestions(ctx context.Context) ([]db.Question, error) {
	return s.svc.ListQuestions(ctx)
}

func (s *QuestionService) GetQuestion(ctx context.Context, id pgtype.UUID) (db.Question, error) {
	return s.svc.GetQuestion(ctx, id)
}

func (s *QuestionService) UpdateQuestion(ctx context.Context, question db.Question) (db.Question, error) {
	arg := db.UpdateQuestionParams{
		ID:           question.ID,
		Statement:    question.Statement,
		Year:         question.Year,
		TopicID:      question.TopicID,
		Position:     question.Position,
		Level:        question.Level,
		Difficulty:   question.Difficulty,
		Modality:     question.Modality,
		PracticeArea: question.PracticeArea,
		FieldOfStudy: question.FieldOfStudy,
	}
	return s.svc.UpdateQuestion(ctx, arg)
}

func (s *QuestionService) DeleteQuestion(ctx context.Context, id pgtype.UUID) error {
	return s.svc.DeleteQuestion(ctx, id)
}

func (s *QuestionService) ListQuestionsByTopic(ctx context.Context, topicID pgtype.UUID) ([]db.Question, error) {
	return s.svc.ListQuestionsByTopic(ctx, topicID)
}

func (s *QuestionService) ListQuestionsByFilters(ctx context.Context, filters QuestionFilter) ([]db.Question, error) {
	params := db.ListQuestionsByFiltersParams{}

	// Convert pointer fields to values, using empty/invalid values when nil
	if filters.TopicID != nil {
		params.TopicID = *filters.TopicID
	}
	if filters.Position != nil {
		params.Position = *filters.Position
	}
	if filters.Level != nil {
		params.Level = *filters.Level
	}
	if filters.Difficulty != nil {
		params.Difficulty = *filters.Difficulty
	}
	if filters.Modality != nil {
		params.Modality = *filters.Modality
	}
	if filters.PracticeArea != nil {
		params.PracticeArea = *filters.PracticeArea
	}
	if filters.FieldOfStudy != nil {
		params.FieldOfStudy = *filters.FieldOfStudy
	}

	row, err := s.svc.ListQuestionsByFilters(ctx, params)
	if err != nil {
		return []db.Question{}, err
	}
	return row, nil
}

func (s *QuestionService) ListQuestionsByFiltersWithChoices(ctx context.Context, filters QuestionFilter) ([]db.ListQuestionsByFiltersWithChoicesRow, error) {
	params := db.ListQuestionsByFiltersWithChoicesParams{}

	// Convert pointer fields to values, using empty/invalid values when nil
	if filters.TopicID != nil {
		params.TopicID = *filters.TopicID
	}
	if filters.Position != nil {
		params.Position = *filters.Position
	}
	if filters.Level != nil {
		params.Level = *filters.Level
	}
	if filters.Difficulty != nil {
		params.Difficulty = *filters.Difficulty
	}
	if filters.Modality != nil {
		params.Modality = *filters.Modality
	}
	if filters.PracticeArea != nil {
		params.PracticeArea = *filters.PracticeArea
	}
	if filters.FieldOfStudy != nil {
		params.FieldOfStudy = *filters.FieldOfStudy
	}

	row, err := s.svc.ListQuestionsByFiltersWithChoices(ctx, params)
	if err != nil {
		return []db.ListQuestionsByFiltersWithChoicesRow{}, err
	}
	return row, nil
}
