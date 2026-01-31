package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
	"github.com/JeanGrijp/AutoBanca/internal/service"
)

type QuestionHandler struct {
	svc  *service.QuestionService
	csvc *service.ChoiceService
	isvc *service.ImportService
}

func NewQuestionHandler(svc *service.QuestionService, csvc *service.ChoiceService, isvc *service.ImportService) *QuestionHandler {
	return &QuestionHandler{svc: svc, csvc: csvc, isvc: isvc}
}

func (h *QuestionHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Question handler")

	var body struct {
		Statement    string      `json:"statement"`
		Year         int32       `json:"year"`
		TopicID      pgtype.UUID `json:"topic_id"`
		Position     pgtype.Text `json:"position"`
		Level        pgtype.Text `json:"level"`
		Difficulty   pgtype.Text `json:"difficulty"`
		Modality     pgtype.Text `json:"modality"`
		PracticeArea pgtype.Text `json:"practice_area"`
		FieldOfStudy pgtype.Text `json:"field_of_study"`
	}

	slog.InfoContext(r.Context(), "Decoding request body")
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question, err := h.svc.CreateQuestion(r.Context(), db.Question{
		Statement:    body.Statement,
		Year:         body.Year,
		TopicID:      body.TopicID,
		Position:     body.Position,
		Level:        body.Level,
		Difficulty:   body.Difficulty,
		Modality:     body.Modality,
		PracticeArea: body.PracticeArea,
		FieldOfStudy: body.FieldOfStudy,
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

type importError struct {
	Linha   int      `json:"linha"`
	Erros   []string `json:"erros"`
	Valores []string `json:"valores,omitempty"`
}

type importResponse struct {
	Total      int           `json:"total"`
	Criadas    int           `json:"criadas"`
	Ignoradas  int           `json:"ignoradas"`
	Falharam   int           `json:"falharam"`
	Detalhes   []importError `json:"detalhes"`
	ColunasCSV []string      `json:"colunas_csv"`
}

func (h *QuestionHandler) ImportQuestionsCSV(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Importing questions from CSV")

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Novo formato com 6 colunas extras para as alternativas
	// choice_a, choice_b, choice_c, choice_d, choice_e, correct_choice (A-E)
	expectedHeaders := []string{
		"statement", "year", "topic_id", "position", "level", "difficulty",
		"modality", "practice_area", "field_of_study",
		"choice_a", "choice_b", "choice_c", "choice_d", "choice_e", "correct_choice",
	}

	resp := importResponse{ColunasCSV: expectedHeaders}
	line := 0
	firstRow, err := reader.Read()
	if err == io.EOF {
		http.Error(w, "csv vazio", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	line++

	row := firstRow
	if isHeaderRow(firstRow, expectedHeaders) {
		row, err = reader.Read()
		if err == io.EOF {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		line++
	}

	for {
		if len(row) != len(expectedHeaders) {
			resp.Total++
			resp.Falharam++
			resp.Detalhes = append(resp.Detalhes, importError{
				Linha:   line,
				Erros:   []string{"quantidade de colunas inválida (esperado: 15)"},
				Valores: row,
			})
		} else {
			resp.Total++
			erros := []string{}

			statement := strings.TrimSpace(row[0])
			yearStr := strings.TrimSpace(row[1])
			topicIDStr := strings.TrimSpace(row[2])
			position := strings.TrimSpace(row[3])
			level := strings.TrimSpace(row[4])
			difficulty := strings.TrimSpace(row[5])
			modality := strings.TrimSpace(row[6])
			practiceArea := strings.TrimSpace(row[7])
			fieldOfStudy := strings.TrimSpace(row[8])

			// Novas colunas para alternativas
			choiceA := strings.TrimSpace(row[9])
			choiceB := strings.TrimSpace(row[10])
			choiceC := strings.TrimSpace(row[11])
			choiceD := strings.TrimSpace(row[12])
			choiceE := strings.TrimSpace(row[13])
			correctChoice := strings.ToUpper(strings.TrimSpace(row[14]))

			// Valida campos obrigatórios da questão
			if statement == "" || yearStr == "" || topicIDStr == "" || position == "" || level == "" || difficulty == "" || modality == "" || practiceArea == "" || fieldOfStudy == "" {
				erros = append(erros, "todos os campos da questão são obrigatórios")
			}

			// Valida alternativas
			if choiceA == "" || choiceB == "" || choiceC == "" || choiceD == "" || choiceE == "" {
				erros = append(erros, "todas as 5 alternativas (A-E) são obrigatórias")
			}

			// Valida correct_choice
			if correctChoice != "A" && correctChoice != "B" && correctChoice != "C" && correctChoice != "D" && correctChoice != "E" {
				erros = append(erros, "correct_choice deve ser A, B, C, D ou E")
			}

			year64, err := strconv.ParseInt(yearStr, 10, 32)
			if err != nil {
				erros = append(erros, "year inválido")
			}

			topicID := pgtype.UUID{}
			if err := topicID.Scan(topicIDStr); err != nil {
				erros = append(erros, "topic_id inválido")
			}

			if len(erros) == 0 {
				question := db.Question{
					Statement:    statement,
					Year:         int32(year64),
					TopicID:      topicID,
					Position:     pgtype.Text{String: position, Valid: true},
					Level:        pgtype.Text{String: level, Valid: true},
					Difficulty:   pgtype.Text{String: difficulty, Valid: true},
					Modality:     pgtype.Text{String: modality, Valid: true},
					PracticeArea: pgtype.Text{String: practiceArea, Valid: true},
					FieldOfStudy: pgtype.Text{String: fieldOfStudy, Valid: true},
				}

				// Monta as choices com o indicador de qual é correta
				choices := []service.ChoiceInput{
					{Text: choiceA, IsCorrect: correctChoice == "A"},
					{Text: choiceB, IsCorrect: correctChoice == "B"},
					{Text: choiceC, IsCorrect: correctChoice == "C"},
					{Text: choiceD, IsCorrect: correctChoice == "D"},
					{Text: choiceE, IsCorrect: correctChoice == "E"},
				}

				// Usa o ImportService com transação para criar questão + alternativas atomicamente
				input := service.QuestionWithChoicesInput{
					Question: question,
					Choices:  choices,
				}
				_, _, createErr := h.isvc.CreateQuestionWithChoices(r.Context(), input)
				if createErr != nil {
					// Verifica se é erro de duplicidade
					if errors.Is(createErr, service.ErrQuestionAlreadyExists) {
						resp.Ignoradas++
					} else {
						erros = append(erros, createErr.Error())
					}
				} else {
					resp.Criadas++
				}
			}

			if len(erros) > 0 {
				resp.Falharam++
				resp.Detalhes = append(resp.Detalhes, importError{
					Linha:   line,
					Erros:   erros,
					Valores: row,
				})
			}
		}

		row, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			resp.Total++
			resp.Falharam++
			resp.Detalhes = append(resp.Detalhes, importError{
				Linha: line + 1,
				Erros: []string{err.Error()},
			})
			break
		}
		line++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func isHeaderRow(row []string, expected []string) bool {
	if len(row) != len(expected) {
		return false
	}
	for i, v := range row {
		if strings.ToLower(strings.TrimSpace(v)) != expected[i] {
			return false
		}
	}
	return true
}

func (h *QuestionHandler) ListQuestionsByFilters(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Listing questions by filters")

	var body struct {
		TopicID      *pgtype.UUID `json:"topic_id"`
		Position     *pgtype.Text `json:"position"`
		Level        *pgtype.Text `json:"level"`
		Difficulty   *pgtype.Text `json:"difficulty"`
		Modality     *pgtype.Text `json:"modality"`
		PracticeArea *pgtype.Text `json:"practice_area"`
		FieldOfStudy *pgtype.Text `json:"field_of_study"`
	}

	// Try to decode body, but allow empty body (list all questions)
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err.Error() != "EOF" {
			slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	filters := service.QuestionFilter{
		TopicID:      body.TopicID,
		Position:     body.Position,
		Level:        body.Level,
		Difficulty:   body.Difficulty,
		Modality:     body.Modality,
		PracticeArea: body.PracticeArea,
		FieldOfStudy: body.FieldOfStudy,
	}

	questions, err := h.svc.ListQuestionsByFilters(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error listing questions by filters", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type questionWithChoices struct {
		Question db.Question `json:"question"`
		Choices  []db.Choice `json:"choices"`
	}

	questionsWithChoices := make([]questionWithChoices, 0, len(questions))
	for _, q := range questions {
		choices, err := h.csvc.ListChoicesByQuestion(r.Context(), q.ID)
		if err != nil {
			slog.ErrorContext(r.Context(), "Error listing choices for question", "question_id", q.ID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		questionsWithChoices = append(questionsWithChoices, questionWithChoices{
			Question: q,
			Choices:  choices,
		})
	}

	slog.InfoContext(r.Context(), "Questions listed successfully", "count", len(questionsWithChoices))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questionsWithChoices)

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
		Statement    string      `json:"statement"`
		Year         int32       `json:"year"`
		TopicID      pgtype.UUID `json:"topic_id"`
		Position     pgtype.Text `json:"position"`
		Level        pgtype.Text `json:"level"`
		Difficulty   pgtype.Text `json:"difficulty"`
		Modality     pgtype.Text `json:"modality"`
		PracticeArea pgtype.Text `json:"practice_area"`
		FieldOfStudy pgtype.Text `json:"field_of_study"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.ErrorContext(r.Context(), "Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question, err := h.svc.UpdateQuestion(r.Context(), db.Question{
		ID:           body.ID,
		Statement:    body.Statement,
		Year:         body.Year,
		TopicID:      body.TopicID,
		Position:     body.Position,
		Level:        body.Level,
		Difficulty:   body.Difficulty,
		Modality:     body.Modality,
		PracticeArea: body.PracticeArea,
		FieldOfStudy: body.FieldOfStudy,
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
