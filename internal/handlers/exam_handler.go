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
	slog.InfoContext(r.Context(), "Generating exam - request received")

	var body struct {
		Subjects []struct {
			Name          string `json:"name"`
			QuestionCount int32  `json:"question_count"`
		} `json:"subjects"`
		Topics        *[]string `json:"topics"`
		Institution   *string   `json:"institution"`
		Position      *string   `json:"position"`
		Level         *string   `json:"level"`
		Difficulty    *string   `json:"difficulty"`
		Modality      *string   `json:"modality"`
		PracticeArea  *string   `json:"practice_area"`
		FieldOfStudy  *string   `json:"field_of_study"`
		QuestionCount int32     `json:"question_count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "Request body decoded successfully",
		"subjects_count", len(body.Subjects),
		"question_count", body.QuestionCount,
		"difficulty", body.Difficulty,
		"modality", body.Modality,
		"field_of_study", body.FieldOfStudy,
	)

	// Convert body subjects to service SubjectInfo
	subjects := make([]service.SubjectInfo, len(body.Subjects))
	for i, s := range body.Subjects {
		subjects[i] = service.SubjectInfo{
			Name:          s.Name,
			QuestionCount: s.QuestionCount,
		}
		slog.InfoContext(r.Context(), "Subject parsed", "index", i, "name", s.Name, "question_count", s.QuestionCount)
	}

	filters := service.GenerateExamFilters{
		Subjects:      subjects,
		Topics:        body.Topics,
		Difficulty:    stringToPgText(body.Difficulty),
		Modality:      derefString(body.Modality),
		PracticeArea:  derefString(body.PracticeArea),
		FieldOfStudy:  derefString(body.FieldOfStudy),
		Level:         stringToPgTextPtr(body.Level),
		QuestionCount: body.QuestionCount,
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

func stringToPgTextPtr(s *string) *pgtype.Text {
	if s == nil || *s == "" {
		return nil
	}
	return &pgtype.Text{String: *s, Valid: true}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
