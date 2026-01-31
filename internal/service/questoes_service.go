package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type QuestionService struct {
	svc db.Queries
}

type QuestionFilter struct {
	AssuntoID    *pgtype.UUID
	Instituicao  *pgtype.Text
	Cargo        *pgtype.Text
	Nivel        *pgtype.Text
	Dificuldade  *pgtype.Text
	Modalidade   *pgtype.Text
	AreaAtuacao  *pgtype.Text
	AreaFormacao *pgtype.Text
}

func NewQuestionService(svc db.Queries) *QuestionService {
	return &QuestionService{svc: svc}
}

func (s *QuestionService) CreateQuestion(ctx context.Context, question db.Question) (db.Question, error) {
	row, err := s.svc.CreateQuestion(ctx, db.CreateQuestionParams{
		Enunciado:    question.Enunciado,
		Ano:          question.Ano,
		AssuntoID:    question.AssuntoID,
		Instituicao:  question.Instituicao,
		Cargo:        question.Cargo,
		Nivel:        question.Nivel,
		Dificuldade:  question.Dificuldade,
		Modalidade:   question.Modalidade,
		AreaAtuacao:  question.AreaAtuacao,
		AreaFormacao: question.AreaFormacao,
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
		Enunciado:    question.Enunciado,
		Ano:          question.Ano,
		AssuntoID:    question.AssuntoID,
		Instituicao:  question.Instituicao,
		Cargo:        question.Cargo,
		Nivel:        question.Nivel,
		Dificuldade:  question.Dificuldade,
		Modalidade:   question.Modalidade,
		AreaAtuacao:  question.AreaAtuacao,
		AreaFormacao: question.AreaFormacao,
	}
	return s.svc.UpdateQuestion(ctx, arg)
}

func (s *QuestionService) DeleteQuestion(ctx context.Context, id pgtype.UUID) error {
	return s.svc.DeleteQuestion(ctx, id)
}

func (s *QuestionService) ListQuestionsByAssunto(ctx context.Context, assuntoID pgtype.UUID) ([]db.Question, error) {
	return s.svc.ListQuestionsByAssunto(ctx, assuntoID)
}

func (s *QuestionService) ListQuestionsByFilters(ctx context.Context, filters QuestionFilter) ([]db.Question, error) {
	// Converter string para pgtype.UUID aqui (dentro do service)
	var assuntoID *pgtype.UUID
	if filters.AssuntoID != nil {
		uuid := pgtype.UUID{}
		if err := uuid.Scan(*filters.AssuntoID); err != nil {
			return []db.Question{}, err
		}
		assuntoID = &uuid
	}

	row, err := s.svc.ListQuestionsByFilters(ctx, db.ListQuestionsByFiltersParams{
		AssuntoID:    *assuntoID,
		Instituicao:  *filters.Instituicao,
		Cargo:        *filters.Cargo,
		Nivel:        *filters.Nivel,
		Dificuldade:  *filters.Dificuldade,
		Modalidade:   *filters.Modalidade,
		AreaAtuacao:  *filters.AreaAtuacao,
		AreaFormacao: *filters.AreaFormacao,
	})
	if err != nil {
		return []db.Question{}, err
	}
	return row, nil
}
