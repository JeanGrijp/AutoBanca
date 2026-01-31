// Copyright 2024 Jean Grijp. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package handlers implements HTTP handlers for the Alternativa entity.
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

type AlternativaHandler struct {
	svc *service.AlternativaService
}

func NewAlternativaHandler(svc *service.AlternativaService) *AlternativaHandler {
	return &AlternativaHandler{svc: svc}
}

// ListAlternativasByQuestion returns all alternatives for a specific question.
// It expects "question_id" as a query parameter.
func (h *AlternativaHandler) ListAlternativasByQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParamFromCtx(r.Context(), "question_id")
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

	alternativas, err := h.svc.ListAlternativasByQuestion(r.Context(), questionIDUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing alternativas", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alternativas)
}

// CreateAlternativa creates a new alternativa.
func (h *AlternativaHandler) CreateAlternativa(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Creating alternativa")
	var body struct {
		TextoAlternativa string      `json:"texto_alternativa"`
		Correta          pgtype.Bool `json:"correta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bodyForm := db.Alternativa{
		TextoAlternativa: body.TextoAlternativa,
		Correta:          body.Correta,
	}

	alternativa, err := h.svc.CreateAlternativa(r.Context(), bodyForm)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating alternativa", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Alternativa created successfully", "alternativa_id", alternativa.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alternativa)
}

func (h *AlternativaHandler) GetAlternativa(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Getting alternativa")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}
	alternativa, err := h.svc.GetAlternativa(r.Context(), idUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error getting alternativa", "error", err)
		http.Error(w, "Alternativa not found or invalid ID", http.StatusNotFound)
		return
	}

	slog.InfoContext(r.Context(), "Successfully retrieved alternativa", "id", alternativa.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alternativa)
}

func (h *AlternativaHandler) UpdateAlternativa(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Updating alternativa")
	id := chi.URLParam(r, "id")
	var body struct {
		ID               pgtype.UUID `json:"id"`
		TextoAlternativa string      `json:"texto_alternativa"`
		Correta          pgtype.Bool `json:"correta"`
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

	alternativaForm := db.Alternativa{
		ID:               body.ID,
		TextoAlternativa: body.TextoAlternativa,
		Correta:          body.Correta,
	}

	alternativa, err := h.svc.UpdateAlternativa(r.Context(), alternativaForm)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error updating alternativa", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Successfully updated alternativa", "id", alternativa.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alternativa)
}

func (h *AlternativaHandler) DeleteAlternativa(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Deleting alternativa")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteAlternativa(r.Context(), idUUID); err != nil {
		slog.ErrorContext(r.Context(), "Error deleting alternativa", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Successfully deleted alternativa", "id", idUUID)

	w.WriteHeader(http.StatusNoContent)
}
