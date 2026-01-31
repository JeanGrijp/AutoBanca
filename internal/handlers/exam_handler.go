package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type ExamHandler struct {
	svc *service.ExamService
}

func NewExamHandler(svc *service.ExamService) *ExamHandler {
	return &ExamHandler{svc: svc}
}

func (h *ExamHandler) GenerateExam(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "--------------------------------------------------------------------------------")
	slog.InfoContext(r.Context(), "Generating exam - request received")
	slog.InfoContext(r.Context(), "--------------------------------------------------------------------------------")

	var body struct {
		Subjects []struct {
			Name          string   `json:"name"`
			QuestionCount int32    `json:"question_count"`
			Topics        []string `json:"topics,omitempty"`
		} `json:"subjects"`
		Difficulty   *string `json:"difficulty"`
		Level        *string `json:"level"`
		Modality     *string `json:"modality"`
		Position     *string `json:"position"`
		FieldOfStudy *string `json:"field_of_study"`
		MinYear      *int32  `json:"min_year"`
		MaxYear      *int32  `json:"max_year"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "Request body decoded successfully",
		"subjects_count", len(body.Subjects),
		"difficulty", body.Difficulty,
		"modality", body.Modality,
		"field_of_study", body.FieldOfStudy,
	)

	// Convert body subjects to service SubjectFilter
	subjects := make([]service.SubjectFilter, len(body.Subjects))
	for i, s := range body.Subjects {
		subjects[i] = service.SubjectFilter{
			Name:          s.Name,
			QuestionCount: s.QuestionCount,
			Topics:        s.Topics,
		}
		slog.InfoContext(r.Context(), "Subject parsed", "index", i, "name", s.Name, "question_count", s.QuestionCount, "topics", s.Topics)
	}

	filters := service.GenerateExamFilters{
		Subjects:     subjects,
		Difficulty:   stringToPgText(body.Difficulty),
		Level:        stringToPgText(body.Level),
		Modality:     stringToPgText(body.Modality),
		Position:     stringToPgText(body.Position),
		FieldOfStudy: stringToPgText(body.FieldOfStudy),
		MinYear:      int32ToPgInt4(body.MinYear),
		MaxYear:      int32ToPgInt4(body.MaxYear),
	}

	slog.InfoContext(r.Context(), "Filters created, calling service to generate exam")

	pdfBytes, err := h.svc.GenerateExam(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error generating exam", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "PDF generated", "size_bytes", len(pdfBytes))

	timeStamp := time.Now()

	examName := fmt.Sprintf("%s_exam_%s_%s_%s.pdf", "AutoBanca", timeStamp.Format("2006-01-02"), timeStamp.Format("15-04-05"), timeStamp.Format("000"))

	slog.InfoContext(r.Context(), "Exam generated successfully", "exam_name", examName)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", examName))
	w.Write(pdfBytes)
}

func stringToPgText(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func int32ToPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}
