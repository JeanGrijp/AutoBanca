package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type AssuntoHandler struct {
	svc *service.AssuntoService
}

func NewAssuntoHandler(ctx context.Context, svc *service.AssuntoService) *AssuntoHandler {
	if svc == nil {
		slog.ErrorContext(ctx, "AssuntoService is nil")
	}
	slog.InfoContext(ctx, "AssuntoService is valid")
	return &AssuntoHandler{svc: svc}
}

func (h *AssuntoHandler) ListAssuntos(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing assuntos")
	assuntos, err := h.svc.ListAssuntos(r.Context())
	slog.InfoContext(r.Context(), "Assuntos listed")
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing assuntos", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Assuntos listed successfully", "count", len(assuntos))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assuntos)
}

func (h *AssuntoHandler) CreateAssunto(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Creating assunto")
	var body struct {
		Nome string `json:"nome"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	assunto, err := h.svc.CreateAssunto(r.Context(), body.Nome)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating assunto", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Assunto created successfully", "assunto_id", assunto.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assunto)
}

func (h *AssuntoHandler) GetAssunto(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Getting assunto")
	id := chi.URLParam(r, "id")

	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	assunto, err := h.svc.GetAssunto(r.Context(), idUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error getting assunto", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assunto)
}

func (h *AssuntoHandler) UpdateAssunto(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Updating assunto")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body struct {
		Nome string `json:"nome"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	assunto, err := h.svc.UpdateAssunto(r.Context(), idUUID, body.Nome)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error updating assunto", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Assunto updated", "assunto_id", assunto.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assunto)
}

func (h *AssuntoHandler) DeleteAssunto(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Deleting assunto")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteAssunto(r.Context(), idUUID); err != nil {
		slog.ErrorContext(r.Context(), "Error deleting assunto", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.InfoContext(r.Context(), "Assunto deleted", "assunto_id", idUUID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AssuntoHandler) ListAssuntosByDisciplina(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing assuntos by disciplina")
	disciplinaID := chi.URLParam(r, "disciplina_id")
	disciplinaUUID := pgtype.UUID{}
	if err := disciplinaUUID.Scan(disciplinaID); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	assuntos, err := h.svc.ListAssuntosByDisciplina(r.Context(), disciplinaUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing assuntos by disciplina", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Assuntos listed by disciplina", "disciplina_id", disciplinaUUID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assuntos)
}
