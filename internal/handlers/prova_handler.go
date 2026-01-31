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

type ProvaHandler struct {
	svc *service.ProvaService
}

func NewProvaHandler(svc *service.ProvaService) *ProvaHandler {
	return &ProvaHandler{svc: svc}
}

func (h *ProvaHandler) GenerateProva(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "Generating prova")

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

	pdfBytes, err := h.svc.GenerateProva(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error generating prova", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeStamp := time.Now()

	provaName := fmt.Sprintf("%s_prova_%s_%s_%s.pdf", "AutoBanca", timeStamp.Format("2006-01-02"), timeStamp.Format("15-04-05"), timeStamp.Format("000"))

	slog.InfoContext(r.Context(), "Prova generated successfully", "prova_name", provaName)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", provaName))
	w.Write(pdfBytes)
}
