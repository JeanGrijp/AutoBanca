package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type DisciplinaService struct {
	q db.Querier
}

func NewDisciplinaService(q db.Querier) *DisciplinaService {
	return &DisciplinaService{
		q: q,
	}
}

func (s *DisciplinaService) CreateDisciplina(ctx context.Context, nome string) (db.Disciplina, error) {
	row, err := s.q.CreateDisciplina(ctx, nome)
	if err != nil {
		return db.Disciplina{}, err
	}
	return row, nil
}

func (s *DisciplinaService) ListDisciplinas(ctx context.Context) ([]db.Disciplina, error) {
	return s.q.ListDisciplinas(ctx)
}

func (s *DisciplinaService) GetDisciplina(ctx context.Context, id string) (db.Disciplina, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return db.Disciplina{}, err
	}
	return s.q.GetDisciplina(ctx, uuid)
}

func (s *DisciplinaService) UpdateDisciplina(ctx context.Context, id string, nome string) (db.Disciplina, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return db.Disciplina{}, err
	}

	arg := db.UpdateDisciplinaParams{
		ID:   uuid,
		Nome: nome,
	}
	return s.q.UpdateDisciplina(ctx, arg)
}

func (s *DisciplinaService) DeleteDisciplina(ctx context.Context, id string) error {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return err
	}
	return s.q.DeleteDisciplina(ctx, uuid)
}
