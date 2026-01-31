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

type TopicHandler struct {
	svc *service.TopicService
}

func NewTopicHandler(ctx context.Context, svc *service.TopicService) *TopicHandler {
	if svc == nil {
		slog.ErrorContext(ctx, "TopicService is nil")
	}
	slog.InfoContext(ctx, "TopicService is valid")
	return &TopicHandler{svc: svc}
}

func (h *TopicHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing topics")
	topics, err := h.svc.ListTopics(r.Context())
	slog.InfoContext(r.Context(), "Topics listed")
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing topics", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Topics listed successfully", "count", len(topics))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
}

func (h *TopicHandler) CreateTopic(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Creating topic")
	var body struct {
		Name      string      `json:"name"`
		SubjectID pgtype.UUID `json:"subject_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic, err := h.svc.CreateTopic(r.Context(), body.Name, body.SubjectID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error creating topic", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Topic created successfully", "topic_id", topic.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(topic)
}

func (h *TopicHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Getting topic")
	id := chi.URLParam(r, "id")

	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic, err := h.svc.GetTopic(r.Context(), idUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error getting topic", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topic)
}

func (h *TopicHandler) UpdateTopic(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Updating topic")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body struct {
		Name      string      `json:"name"`
		SubjectID pgtype.UUID `json:"subject_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topic, err := h.svc.UpdateTopic(r.Context(), idUUID, body.Name, body.SubjectID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error updating topic", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Topic updated", "topic_id", topic.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topic)
}

func (h *TopicHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Deleting topic")
	id := chi.URLParam(r, "id")
	idUUID := pgtype.UUID{}
	if err := idUUID.Scan(id); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteTopic(r.Context(), idUUID); err != nil {
		slog.ErrorContext(r.Context(), "Error deleting topic", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.InfoContext(r.Context(), "Topic deleted", "topic_id", idUUID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TopicHandler) ListTopicsBySubject(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing topics by subject")
	subjectID := chi.URLParam(r, "subject_id")
	subjectUUID := pgtype.UUID{}
	if err := subjectUUID.Scan(subjectID); err != nil {
		slog.ErrorContext(r.Context(), "Error scanning UUID", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	topics, err := h.svc.ListTopicsBySubject(r.Context(), subjectUUID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing topics by subject", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Topics listed by subject", "subject_id", subjectUUID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
}
