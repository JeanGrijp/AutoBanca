// Copyright 2024 Jean Grijp. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package handlers implements HTTP handlers for the Disciplina entity.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type DisciplinaHandler struct {
	svc *service.DisciplinaService
}

func NewDisciplinaHandler(svc *service.DisciplinaService) *DisciplinaHandler {
	return &DisciplinaHandler{svc: svc}
}

func (h *DisciplinaHandler) ListDisciplinas(w http.ResponseWriter, r *http.Request) {
	disciplinas, err := h.svc.ListDisciplinas(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disciplinas)
}

func (h *DisciplinaHandler) CreateDisciplina(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nome string `json:"nome"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	disciplina, err := h.svc.CreateDisciplina(r.Context(), body.Nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(disciplina)
}

func (h *DisciplinaHandler) GetDisciplina(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	disciplina, err := h.svc.GetDisciplina(r.Context(), id)
	if err != nil {
		http.Error(w, "Disciplina not found or invalid ID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disciplina)
}

func (h *DisciplinaHandler) UpdateDisciplina(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Nome string `json:"nome"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	disciplina, err := h.svc.UpdateDisciplina(r.Context(), id, body.Nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disciplina)
}

func (h *DisciplinaHandler) DeleteDisciplina(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteDisciplina(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
