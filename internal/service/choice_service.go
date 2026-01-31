package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type ChoiceService struct {
	q db.Querier
}

func NewChoiceService(q db.Querier) *ChoiceService {
	return &ChoiceService{q: q}
}

func (s *ChoiceService) CreateChoice(ctx context.Context, choice db.Choice) (db.Choice, error) {
	row, err := s.q.CreateChoice(ctx, db.CreateChoiceParams{
		QuestionID: choice.QuestionID,
		ChoiceText: choice.ChoiceText,
		IsCorrect:  choice.IsCorrect,
	})

	if err != nil {
		return db.Choice{}, err
	}
	return row, nil
}

func (s *ChoiceService) ListChoicesByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.Choice, error) {
	return s.q.ListChoicesByQuestion(ctx, questionID)
}

func (s *ChoiceService) DeleteChoice(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteChoice(ctx, id)
}

func (s *ChoiceService) GetChoice(ctx context.Context, id pgtype.UUID) (db.Choice, error) {
	return s.q.GetChoice(ctx, id)
}

func (s *ChoiceService) UpdateChoice(ctx context.Context, choice db.Choice) (db.Choice, error) {
	row, err := s.q.UpdateChoice(ctx, db.UpdateChoiceParams{
		ID:         choice.ID,
		ChoiceText: choice.ChoiceText,
		IsCorrect:  choice.IsCorrect,
	})

	if err != nil {
		return db.Choice{}, err
	}
	return row, nil
}
