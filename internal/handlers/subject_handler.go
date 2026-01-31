// Copyright 2024 Jean Grijp. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package handlers implements HTTP handlers for the Subject entity.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type SubjectHandler struct {
	svc *service.SubjectService
}

func NewSubjectHandler(svc *service.SubjectService) *SubjectHandler {
	return &SubjectHandler{svc: svc}
}

func (h *SubjectHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing subjects")

	subjects, err := h.svc.ListSubjects(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing subjects", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Subjects listed successfully", "count", len(subjects))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subjects)
}

func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Creating subject")

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "Subject data received", "name", body.Name)

	subj, err := h.svc.GetSubjectByName(r.Context(), body.Name)
	if err == nil {
		slog.InfoContext(r.Context(), "Subject already exists", "subject_id", subj.ID, "name", subj.Name)
		http.Error(w, "subject already exists", http.StatusNoContent)
		return
	}

	subject, err := h.svc.CreateSubject(r.Context(), body.Name)
	if err != nil {
		// Check for duplicate key error (race condition case)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			slog.WarnContext(r.Context(), "Subject already exists (race condition)", "name", body.Name)
			http.Error(w, "subject already exists", http.StatusConflict)
			return
		}
		slog.ErrorContext(r.Context(), "Error creating subject", "error", err, "name", body.Name)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Subject created successfully", "subject_id", subject.ID, "name", subject.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subject)
}

func (h *SubjectHandler) GetSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	slog.InfoContext(r.Context(), "Getting subject", "id", id)

	subject, err := h.svc.GetSubject(r.Context(), id)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error getting subject", "error", err, "id", id)
		http.Error(w, "subject not found or invalid ID", http.StatusNotFound)
		return
	}

	slog.InfoContext(r.Context(), "Subject retrieved successfully", "subject_id", subject.ID, "name", subject.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subject)
}

func (h *SubjectHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	slog.InfoContext(r.Context(), "Updating subject", "id", id)

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "Subject update data received", "id", id, "new_name", body.Name)

	subject, err := h.svc.UpdateSubject(r.Context(), id, body.Name)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error updating subject", "error", err, "id", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Subject updated successfully", "subject_id", subject.ID, "name", subject.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subject)
}

func (h *SubjectHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	slog.InfoContext(r.Context(), "Deleting subject", "id", id)

	if err := h.svc.DeleteSubject(r.Context(), id); err != nil {
		slog.ErrorContext(r.Context(), "Error deleting subject", "error", err, "id", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Subject deleted successfully", "id", id)

	w.WriteHeader(http.StatusNoContent)
}
