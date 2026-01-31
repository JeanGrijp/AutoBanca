// Copyright 2024 Jean Grijp. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package handlers implements HTTP handlers for the Choice entity.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type ChoiceHandler struct {
	svc *service.ChoiceService
}

func NewChoiceHandler(svc *service.ChoiceService) *ChoiceHandler {
	return &ChoiceHandler{svc: svc}
}

// ListChoicesByQuestion returns all choices for a specific question.
// It expects "question_id" as a query parameter.
func (h *ChoiceHandler) ListChoicesByQuestion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing choices by question")

	questionID := r.URL.Query().Get("question_id")
	slog.InfoContext(r.Context(), "question_id", "value", questionID)

	if questionID == "" {
		slog.ErrorContext(r.Context(), "question_id query parameter is required")
		http.Error(w, "question_id query parameter is required", http.StatusBadRequest)
		return
	}

	questionIDUUID := pgtype.UUID{}
	if err := questionIDUUID.Scan(questionID); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning question_id UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	choices, err := h.svc.ListChoicesByQuestion(r.Context(), questionIDUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing choices", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Successfully listed choices", "count", len(choices))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(choices)
}

// CreateChoice creates a new choice.
func (h *ChoiceHandler) CreateChoice(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Creating choice")
	var body struct {
		QuestionID string `json:"question_id"`
		ChoiceText string `json:"choice_text"`
		IsCorrect  bool   `json:"is_correct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questionUUID := pgtype.UUID{}
	if err := questionUUID.Scan(body.QuestionID); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning question_id UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bodyForm := db.Choice{
		QuestionID: questionUUID,
		ChoiceText: body.ChoiceText,
		IsCorrect:  pgtype.Bool{Bool: body.IsCorrect, Valid: true},
	}

	choice, err := h.svc.CreateChoice(r.Context(), bodyForm)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating choice", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Choice created successfully", "choice_id", choice.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(choice)
}

func (h *ChoiceHandler) GetChoice(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Getting choice")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, "invalid ID format", http.StatusBadRequest)
		return
	}
	choice, err := h.svc.GetChoice(r.Context(), idUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error getting choice", "error", err)
		http.Error(w, "choice not found or invalid ID", http.StatusNotFound)
		return
	}

	slog.InfoContext(r.Context(), "Successfully retrieved choice", "id", choice.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(choice)
}

func (h *ChoiceHandler) UpdateChoice(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Updating choice")
	id := chi.URLParam(r, "id")
	var body struct {
		ID         pgtype.UUID `json:"id"`
		ChoiceText string      `json:"choice_text"`
		IsCorrect  bool        `json:"is_correct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body.ID = idUUID

	choiceForm := db.Choice{
		ID:         body.ID,
		ChoiceText: body.ChoiceText,
		IsCorrect:  pgtype.Bool{Bool: body.IsCorrect, Valid: true},
	}

	choice, err := h.svc.UpdateChoice(r.Context(), choiceForm)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error updating choice", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Successfully updated choice", "id", choice.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(choice)
}

func (h *ChoiceHandler) DeleteChoice(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Deleting choice")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteChoice(r.Context(), idUUID); err != nil {
		slog.ErrorContext(r.Context(), "Error deleting choice", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Successfully deleted choice", "id", idUUID)

	w.WriteHeader(http.StatusNoContent)
}
