package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type AlternativaService struct {
	q db.Querier
}

func NewAlternativaService(q db.Querier) *AlternativaService {
	return &AlternativaService{q: q}
}

func (s *AlternativaService) CreateAlternativa(ctx context.Context, alt db.Alternativa) (db.Alternativa, error) {
	row, err := s.q.CreateAlternativa(ctx, db.CreateAlternativaParams{
		TextoAlternativa: alt.TextoAlternativa,
		Correta:          alt.Correta,
	})

	if err != nil {
		return db.Alternativa{}, err
	}
	return row, nil
}

func (s *AlternativaService) ListAlternativasByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.Alternativa, error) {
	return s.q.ListAlternativasByQuestion(ctx, questionID)
}

func (s *AlternativaService) DeleteAlternativa(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteAlternativa(ctx, id)
}

func (s *AlternativaService) GetAlternativa(ctx context.Context, id pgtype.UUID) (db.Alternativa, error) {
	return s.q.GetAlternativa(ctx, id)
}

func (s *AlternativaService) UpdateAlternativa(ctx context.Context, alt db.Alternativa) (db.Alternativa, error) {
	row, err := s.q.UpdateAlternativa(ctx, db.UpdateAlternativaParams{
		ID:               alt.ID,
		TextoAlternativa: alt.TextoAlternativa,
		Correta:          alt.Correta,
	})

	if err != nil {
		return db.Alternativa{}, err
	}
	return row, nil
}
