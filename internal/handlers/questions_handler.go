package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type QuestionHandler struct {
	svc *service.QuestionService
}

func NewQuestionHandler(svc *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{svc: svc}
}

func (h *QuestionHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Question handler")

	var body struct {
		Enunciado    string      `json:"enunciado"`
		Ano          int32       `json:"ano"`
		AssuntoID    pgtype.UUID `json:"assunto_id"`
		Instituicao  pgtype.Text `json:"instituicao"`
		Cargo        pgtype.Text `json:"cargo"`
		Nivel        pgtype.Text `json:"nivel"`
		Dificuldade  pgtype.Text `json:"dificuldade"`
		Modalidade   pgtype.Text `json:"modalidade"`
		AreaAtuacao  pgtype.Text `json:"area_atuacao"`
		AreaFormacao pgtype.Text `json:"area_formacao"`
	}

	slog.InfoContext(r.Context(), "Decoding request body")
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question, err := h.svc.CreateQuestion(r.Context(), db.Question{
		Enunciado:    body.Enunciado,
		Ano:          body.Ano,
		AssuntoID:    body.AssuntoID,
		Instituicao:  body.Instituicao,
		Cargo:        body.Cargo,
		Nivel:        body.Nivel,
		Dificuldade:  body.Dificuldade,
		Modalidade:   body.Modalidade,
		AreaAtuacao:  body.AreaAtuacao,
		AreaFormacao: body.AreaFormacao,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating question", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Question created successfully", "question_id", question.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(question)
}

func (h *QuestionHandler) ListQuestionsByFilters(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing questions by filters")

	var body struct {
		AssuntoID    *pgtype.UUID `json:"assunto_id"`
		Instituicao  *pgtype.Text `json:"instituicao"`
		Cargo        *pgtype.Text `json:"cargo"`
		Nivel        *pgtype.Text `json:"nivel"`
		Dificuldade  *pgtype.Text `json:"dificuldade"`
		Modalidade   *pgtype.Text `json:"modalidade"`
		AreaAtuacao  *pgtype.Text `json:"area_atuacao"`
		AreaFormacao *pgtype.Text `json:"area_formacao"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filters := service.QuestionFilter{
		AssuntoID:    body.AssuntoID,
		Instituicao:  body.Instituicao,
		Cargo:        body.Cargo,
		Nivel:        body.Nivel,
		Dificuldade:  body.Dificuldade,
		Modalidade:   body.Modalidade,
		AreaAtuacao:  body.AreaAtuacao,
		AreaFormacao: body.AreaFormacao,
	}

	questions, err := h.svc.ListQuestionsByFilters(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing questions by filters", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Questions listed successfully", "count", len(questions))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)

}

func (h *QuestionHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Deleting question")

	var body struct {
		ID pgtype.UUID `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteQuestion(r.Context(), body.ID); err != nil {
		slog.ErrorContext(r.Context(), "Error deleting question", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Question deleted successfully", "question_id", body.ID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *QuestionHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Updating question")

	var body struct {
		ID           pgtype.UUID `json:"id"`
		Enunciado    string      `json:"enunciado"`
		Ano          int32       `json:"ano"`
		AssuntoID    pgtype.UUID `json:"assunto_id"`
		Instituicao  pgtype.Text `json:"instituicao"`
		Cargo        pgtype.Text `json:"cargo"`
		Nivel        pgtype.Text `json:"nivel"`
		Dificuldade  pgtype.Text `json:"dificuldade"`
		Modalidade   pgtype.Text `json:"modalidade"`
		AreaAtuacao  pgtype.Text `json:"area_atuacao"`
		AreaFormacao pgtype.Text `json:"area_formacao"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question, err := h.svc.UpdateQuestion(r.Context(), db.Question{
		ID:           body.ID,
		Enunciado:    body.Enunciado,
		Ano:          body.Ano,
		AssuntoID:    body.AssuntoID,
		Instituicao:  body.Instituicao,
		Cargo:        body.Cargo,
		Nivel:        body.Nivel,
		Dificuldade:  body.Dificuldade,
		Modalidade:   body.Modalidade,
		AreaAtuacao:  body.AreaAtuacao,
		AreaFormacao: body.AreaFormacao,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "Error updating question", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Question updated successfully", "question_id", question.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func (h *QuestionHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Getting question")

	var body struct {
		ID pgtype.UUID `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	question, err := h.svc.GetQuestion(r.Context(), body.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error getting question", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Question retrieved successfully", "question_id", question.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}
