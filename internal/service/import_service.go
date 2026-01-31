// Package service implements business logic for importing questions with choices.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

// ImportService handles importing questions with choices in a transaction.
type ImportService struct {
	pool *pgxpool.Pool
}

// NewImportService creates a new ImportService.
func NewImportService(pool *pgxpool.Pool) *ImportService {
	return &ImportService{pool: pool}
}

// ChoiceInput represents a choice to be created.
type ChoiceInput struct {
	Text      string
	IsCorrect bool
}

// QuestionWithChoicesInput represents a question with its choices.
type QuestionWithChoicesInput struct {
	Question db.Question
	Choices  []ChoiceInput
}

// ImportResult represents the result of importing a single question.
type ImportResult struct {
	Success  bool
	Question db.Question
	Choices  []db.Choice
	Error    error
}

// ErrQuestionAlreadyExists is returned when a question with the same statement already exists.
var ErrQuestionAlreadyExists = errors.New("questão já existe no banco de dados")

// CreateQuestionWithChoices creates a question and its choices in a single transaction.
// If any operation fails, the entire transaction is rolled back.
// Returns ErrQuestionAlreadyExists if a question with the same statement already exists.
func (s *ImportService) CreateQuestionWithChoices(ctx context.Context, input QuestionWithChoicesInput) (db.Question, []db.Choice, error) {
	// Validate: must have exactly 5 choices for multiple choice
	if len(input.Choices) != 5 {
		return db.Question{}, nil, errors.New("questão de múltipla escolha deve ter exatamente 5 alternativas (A-E)")
	}

	// Validate: exactly one correct answer
	correctCount := 0
	for _, c := range input.Choices {
		if c.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return db.Question{}, nil, fmt.Errorf("deve haver exatamente 1 alternativa correta, encontrado: %d", correctCount)
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx) // Will be no-op if committed

	// Create queries with transaction
	qtx := db.New(tx)

	// Check if question already exists by statement
	exists, err := qtx.QuestionExistsByStatement(ctx, input.Question.Statement)
	if err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao verificar duplicidade: %w", err)
	}
	if exists {
		return db.Question{}, nil, ErrQuestionAlreadyExists
	}

	// Create question
	question, err := qtx.CreateQuestion(ctx, db.CreateQuestionParams{
		Statement:    input.Question.Statement,
		Year:         input.Question.Year,
		TopicID:      input.Question.TopicID,
		Position:     input.Question.Position,
		Level:        input.Question.Level,
		Difficulty:   input.Question.Difficulty,
		Modality:     input.Question.Modality,
		PracticeArea: input.Question.PracticeArea,
		FieldOfStudy: input.Question.FieldOfStudy,
	})
	if err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao criar questão: %w", err)
	}

	// Create choices
	choices := make([]db.Choice, 0, len(input.Choices))
	for i, c := range input.Choices {
		choice, err := qtx.CreateChoice(ctx, db.CreateChoiceParams{
			QuestionID: question.ID,
			ChoiceText: c.Text,
			IsCorrect:  pgtype.Bool{Bool: c.IsCorrect, Valid: true},
		})
		if err != nil {
			return db.Question{}, nil, fmt.Errorf("erro ao criar alternativa %d: %w", i+1, err)
		}
		choices = append(choices, choice)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return question, choices, nil
}

// CreateTrueFalseQuestion creates a true/false question with exactly 2 choices.
func (s *ImportService) CreateTrueFalseQuestion(ctx context.Context, input QuestionWithChoicesInput) (db.Question, []db.Choice, error) {
	// Validate: must have exactly 2 choices for true/false
	if len(input.Choices) != 2 {
		return db.Question{}, nil, errors.New("questão de verdadeiro/falso deve ter exatamente 2 alternativas")
	}

	// Validate: exactly one correct answer
	correctCount := 0
	for _, c := range input.Choices {
		if c.IsCorrect {
			correctCount++
		}
	}
	if correctCount != 1 {
		return db.Question{}, nil, fmt.Errorf("deve haver exatamente 1 alternativa correta, encontrado: %d", correctCount)
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := db.New(tx)

	// Create question
	question, err := qtx.CreateQuestion(ctx, db.CreateQuestionParams{
		Statement:    input.Question.Statement,
		Year:         input.Question.Year,
		TopicID:      input.Question.TopicID,
		Position:     input.Question.Position,
		Level:        input.Question.Level,
		Difficulty:   input.Question.Difficulty,
		Modality:     input.Question.Modality,
		PracticeArea: input.Question.PracticeArea,
		FieldOfStudy: input.Question.FieldOfStudy,
	})
	if err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao criar questão: %w", err)
	}

	// Create choices
	choices := make([]db.Choice, 0, len(input.Choices))
	for i, c := range input.Choices {
		choice, err := qtx.CreateChoice(ctx, db.CreateChoiceParams{
			QuestionID: question.ID,
			ChoiceText: c.Text,
			IsCorrect:  pgtype.Bool{Bool: c.IsCorrect, Valid: true},
		})
		if err != nil {
			return db.Question{}, nil, fmt.Errorf("erro ao criar alternativa %d: %w", i+1, err)
		}
		choices = append(choices, choice)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Question{}, nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return question, choices, nil
}

// Pool returns the underlying pool for cases where direct access is needed.
func (s *ImportService) Pool() *pgxpool.Pool {
	return s.pool
}

// BeginTx starts a new transaction. Caller is responsible for committing or rolling back.
func (s *ImportService) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.pool.Begin(ctx)
}
