package service

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type SubjectService struct {
	q db.Querier
}

func NewSubjectService(q db.Querier) *SubjectService {
	return &SubjectService{
		q: q,
	}
}

func (s *SubjectService) CreateSubject(ctx context.Context, name string) (db.Subject, error) {

	slog.InfoContext(ctx, "Creating subject", "name", name)

	row, err := s.q.CreateSubject(ctx, name)
	if err != nil {
		return db.Subject{}, err
	}
	return row, nil
}

func (s *SubjectService) ListSubjects(ctx context.Context) ([]db.Subject, error) {
	slog.InfoContext(ctx, "Listing subjects")
	subjects, err := s.q.ListSubjects(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error listing subjects", "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "Subjects listed successfully", "count", len(subjects))
	return subjects, nil

}

func (s *SubjectService) GetSubject(ctx context.Context, id string) (db.Subject, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return db.Subject{}, err
	}
	return s.q.GetSubject(ctx, uuid)
}

func (s *SubjectService) UpdateSubject(ctx context.Context, id string, name string) (db.Subject, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return db.Subject{}, err
	}

	arg := db.UpdateSubjectParams{
		ID:   uuid,
		Name: name,
	}
	return s.q.UpdateSubject(ctx, arg)
}

func (s *SubjectService) DeleteSubject(ctx context.Context, id string) error {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return err
	}
	return s.q.DeleteSubject(ctx, uuid)
}

func (s *SubjectService) GetSubjectByName(ctx context.Context, name string) (db.Subject, error) {
	return s.q.GetSubjectByName(ctx, name)
}
