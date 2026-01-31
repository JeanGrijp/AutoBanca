package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jung-kurt/gofpdf"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type ProvaService struct {
	q db.Querier
}

func NewProvaService(q db.Querier) *ProvaService {
	return &ProvaService{q: q}
}

// GenerateProva generates a new prova based on the provided details.
func (s *ProvaService) GenerateProva(ctx context.Context, filters QuestionFilter) ([]byte, error) {

	questions, err := s.q.ListQuestionsByFilters(ctx, db.ListQuestionsByFiltersParams{
		AssuntoID:    *filters.AssuntoID,
		Instituicao:  *filters.Instituicao,
		Cargo:        *filters.Cargo,
		Nivel:        *filters.Nivel,
		Dificuldade:  *filters.Dificuldade,
		Modalidade:   *filters.Modalidade,
		AreaAtuacao:  *filters.AreaAtuacao,
		AreaFormacao: *filters.AreaFormacao,
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao buscar questões no banco de dados: %v", err)
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("nenhuma questão encontrada para os filtros fornecidos")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Prova - AutoBanca") // Título fixo da banca
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)

	for i, q := range questions {
		texto := fmt.Sprintf("%d) %s", i+1, q.Enunciado)
		// MultiCell é melhor para enunciados longos, pois quebra a linha automaticamente
		pdf.MultiCell(190, 8, texto, "0", "L", false)
		pdf.Ln(4)
	}

	// 4. Retorno como Buffer (para facilitar o envio via HTTP ou salvar em disco)
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar buffer do PDF: %v", err)
	}

	return buf.Bytes(), nil

}
