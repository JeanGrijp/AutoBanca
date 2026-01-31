package handlers

import (
	"encoding/csv"
	"encoding/json"
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

type importError struct {
	Linha   int      `json:"linha"`
	Erros   []string `json:"erros"`
	Valores []string `json:"valores,omitempty"`
}

type importResponse struct {
	Total      int           `json:"total"`
	Criadas    int           `json:"criadas"`
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

	expectedHeaders := []string{"enunciado", "ano", "assunto_id", "instituicao", "cargo", "nivel", "dificuldade", "modalidade", "area_atuacao", "area_formacao"}

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
				Erros:   []string{"quantidade de colunas inválida"},
				Valores: row,
			})
		} else {
			resp.Total++
			erros := []string{}

			enunciado := strings.TrimSpace(row[0])
			anoStr := strings.TrimSpace(row[1])
			assuntoIDStr := strings.TrimSpace(row[2])
			instituicao := strings.TrimSpace(row[3])
			cargo := strings.TrimSpace(row[4])
			nivel := strings.TrimSpace(row[5])
			dificuldade := strings.TrimSpace(row[6])
			modalidade := strings.TrimSpace(row[7])
			areaAtuacao := strings.TrimSpace(row[8])
			areaFormacao := strings.TrimSpace(row[9])

			if enunciado == "" || anoStr == "" || assuntoIDStr == "" || instituicao == "" || cargo == "" || nivel == "" || dificuldade == "" || modalidade == "" || areaAtuacao == "" || areaFormacao == "" {
				erros = append(erros, "todos os campos são obrigatórios")
			}

			ano64, err := strconv.ParseInt(anoStr, 10, 32)
			if err != nil {
				erros = append(erros, "ano inválido")
			}

			assuntoID := pgtype.UUID{}
			if err := assuntoID.Scan(assuntoIDStr); err != nil {
				erros = append(erros, "assunto_id inválido")
			}

			if len(erros) == 0 {
				question := db.Question{
					Enunciado:    enunciado,
					Ano:          int32(ano64),
					AssuntoID:    assuntoID,
					Instituicao:  pgtype.Text{String: instituicao, Valid: true},
					Cargo:        pgtype.Text{String: cargo, Valid: true},
					Nivel:        pgtype.Text{String: nivel, Valid: true},
					Dificuldade:  pgtype.Text{String: dificuldade, Valid: true},
					Modalidade:   pgtype.Text{String: modalidade, Valid: true},
					AreaAtuacao:  pgtype.Text{String: areaAtuacao, Valid: true},
					AreaFormacao: pgtype.Text{String: areaFormacao, Valid: true},
				}

				if _, err := h.svc.CreateQuestion(r.Context(), question); err != nil {
					erros = append(erros, err.Error())
				}
			}

			if len(erros) > 0 {
				resp.Falharam++
				resp.Detalhes = append(resp.Detalhes, importError{
					Linha:   line,
					Erros:   erros,
					Valores: row,
				})
			} else {
				resp.Criadas++
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
