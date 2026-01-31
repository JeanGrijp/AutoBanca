package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type AssuntoService struct {
	q db.Querier
}

func NewAssuntoService(q db.Querier) *AssuntoService {
	return &AssuntoService{
		q: q,
	}
}

func (s *AssuntoService) CreateAssunto(ctx context.Context, nome string) (db.Assunto, error) {
	row, err := s.q.CreateAssunto(ctx, db.CreateAssuntoParams{
		Nome: nome,
	})
	if err != nil {
		return db.Assunto{}, err
	}
	return row, nil
}

func (s *AssuntoService) ListAssuntos(ctx context.Context) ([]db.Assunto, error) {
	return s.q.ListAssuntos(ctx)
}

func (s *AssuntoService) GetAssunto(ctx context.Context, id pgtype.UUID) (db.Assunto, error) {
	return s.q.GetAssunto(ctx, id)
}

func (s *AssuntoService) UpdateAssunto(ctx context.Context, id pgtype.UUID, nome string) (db.Assunto, error) {
	arg := db.UpdateAssuntoParams{
		ID:   id,
		Nome: nome,
	}
	return s.q.UpdateAssunto(ctx, arg)
}

func (s *AssuntoService) DeleteAssunto(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteAssunto(ctx, id)
}

func (s *AssuntoService) ListAssuntosByDisciplina(ctx context.Context, disciplinaID pgtype.UUID) ([]db.Assunto, error) {
	return s.q.ListAssuntosByDisciplina(ctx, disciplinaID)
}
