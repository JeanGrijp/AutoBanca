package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jung-kurt/gofpdf"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

// ExamService gerencia a geração de provas
type ExamService struct {
	q           db.Querier
	svcSubject  *SubjectService
	svcTopic    *TopicService
	svcQuestion *QuestionService
}

// SubjectFilter representa o filtro de matéria para geração de prova
type SubjectFilter struct {
	Name          string   `json:"name"`
	QuestionCount int32    `json:"question_count"`
	Topics        []string `json:"topics,omitempty"`
}

// SubjectAndCount agrupa matéria com contagem
type SubjectAndCount struct {
	Subject db.Subject
	Count   int32
	Topics  []string
}

// GenerateExamFilters representa os filtros para geração de prova
type GenerateExamFilters struct {
	Subjects     []SubjectFilter `json:"subjects"`
	Difficulty   pgtype.Text     `json:"difficulty"`
	Level        pgtype.Text     `json:"level"`
	Modality     pgtype.Text     `json:"modality"`
	Position     pgtype.Text     `json:"position"`
	FieldOfStudy pgtype.Text     `json:"field_of_study"`
	MinYear      pgtype.Int4     `json:"min_year"`
	MaxYear      pgtype.Int4     `json:"max_year"`
}

// QuestionWithChoices agrupa uma questão com suas alternativas
type QuestionWithChoices struct {
	Question db.GetQuestionsForExamRow
	Choices  []db.Choice
}

// SubjectQuestions agrupa questões por matéria
type SubjectQuestions struct {
	SubjectName string
	Questions   []QuestionWithChoices
}

// GabaritoItem representa um item do gabarito
type GabaritoItem struct {
	Number  int
	Answer  string
	Subject string
}

// IsValid verifica se os filtros são válidos
func (gef *GenerateExamFilters) IsValid() bool {
	if len(gef.Subjects) == 0 {
		return false
	}
	for _, s := range gef.Subjects {
		if s.QuestionCount <= 0 || s.Name == "" {
			return false
		}
	}
	if !gef.Modality.Valid || gef.Modality.String == "" {
		return false
	}
	if !gef.FieldOfStudy.Valid || gef.FieldOfStudy.String == "" {
		return false
	}
	return true
}

// NewExamService cria uma nova instância do ExamService
func NewExamService(q db.Querier, svcSubject *SubjectService, svcTopic *TopicService, svcQuestion *QuestionService) *ExamService {
	return &ExamService{
		q:           q,
		svcSubject:  svcSubject,
		svcTopic:    svcTopic,
		svcQuestion: svcQuestion,
	}
}

// GenerateExam gera uma prova em PDF com base nos filtros fornecidos
func (s *ExamService) GenerateExam(ctx context.Context, filters GenerateExamFilters) ([]byte, error) {
	slog.InfoContext(ctx, "-----------------------------")
	slog.InfoContext(ctx, "Generate Exam Service")
	slog.InfoContext(ctx, "-----------------------------")

	if !filters.IsValid() {
		slog.ErrorContext(ctx, "Invalid filters")
		return nil, fmt.Errorf("invalid filters provided")
	}

	// 1. Buscar dados do banco
	subjectQuestionsList, gabarito, err := s.fetchExamData(ctx, filters)
	if err != nil {
		return nil, err
	}

	totalQuestions := s.countTotalQuestions(subjectQuestionsList)
	slog.InfoContext(ctx, "Total questions fetched", "count", totalQuestions)

	if totalQuestions == 0 {
		slog.ErrorContext(ctx, "Nenhuma questão encontrada para os filtros fornecidos", "filters", filters)
		return nil, fmt.Errorf("nenhuma questão encontrada para os filtros fornecidos")
	}

	// 2. Gerar PDF
	slog.InfoContext(ctx, "Gerando PDF com as questões")
	pdfBytes, err := s.generatePDF(subjectQuestionsList, gabarito, totalQuestions)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "PDF gerado com sucesso!!!", "total_questions", totalQuestions)
	return pdfBytes, nil
}

// fetchExamData busca as questões e alternativas do banco de dados
func (s *ExamService) fetchExamData(ctx context.Context, filters GenerateExamFilters) ([]SubjectQuestions, []GabaritoItem, error) {
	var subjectQuestionsList []SubjectQuestions
	var gabarito []GabaritoItem
	questionNumber := 1

	for _, subjectFilter := range filters.Subjects {
		slog.InfoContext(ctx, "Fetching subject", "name", subjectFilter.Name)

		subject, err := s.svcSubject.GetSubjectByName(ctx, subjectFilter.Name)
		if err != nil {
			slog.ErrorContext(ctx, "Error fetching subject", "error", err)
			return nil, nil, fmt.Errorf("error fetching subject %s: %v", subjectFilter.Name, err)
		}
		slog.InfoContext(ctx, "Subject found", "subject", subject)

		questionsWithChoices, itemsGabarito, newQuestionNumber, err := s.fetchQuestionsForSubject(
			ctx, subject, subjectFilter, filters, questionNumber,
		)
		if err != nil {
			return nil, nil, err
		}

		gabarito = append(gabarito, itemsGabarito...)
		questionNumber = newQuestionNumber

		if len(questionsWithChoices) > 0 {
			subjectQuestionsList = append(subjectQuestionsList, SubjectQuestions{
				SubjectName: subject.Name,
				Questions:   questionsWithChoices,
			})
		}
	}

	return subjectQuestionsList, gabarito, nil
}

// fetchQuestionsForSubject busca questões e alternativas para uma matéria específica
func (s *ExamService) fetchQuestionsForSubject(
	ctx context.Context,
	subject db.Subject,
	subjectFilter SubjectFilter,
	filters GenerateExamFilters,
	questionNumber int,
) ([]QuestionWithChoices, []GabaritoItem, int, error) {

	var topicID pgtype.UUID // vazio = não filtra por tópico

	questions, err := s.q.GetQuestionsForExam(ctx, db.GetQuestionsForExamParams{
		ID:           subject.ID,
		Limit:        subjectFilter.QuestionCount,
		TopicID:      topicID,
		Position:     filters.Position,
		Level:        filters.Level,
		Difficulty:   filters.Difficulty,
		Modality:     filters.Modality,
		FieldOfStudy: filters.FieldOfStudy,
		MinYear:      filters.MinYear,
		MaxYear:      filters.MaxYear,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error fetching questions for subject", "subject", subject.Name, "error", err)
		return nil, nil, questionNumber, fmt.Errorf("error fetching questions for subject %s: %v", subject.Name, err)
	}

	slog.InfoContext(ctx, "Questions fetched for subject", "subject", subject.Name, "count", len(questions))

	var questionsWithChoices []QuestionWithChoices
	var gabaritoItems []GabaritoItem

	for _, q := range questions {
		choices, err := s.q.ListChoicesByQuestion(ctx, q.ID)
		if err != nil {
			slog.ErrorContext(ctx, "Error fetching choices for question", "question_id", q.ID, "error", err)
			return nil, nil, questionNumber, fmt.Errorf("error fetching choices for question: %v", err)
		}

		questionsWithChoices = append(questionsWithChoices, QuestionWithChoices{
			Question: q,
			Choices:  choices,
		})

		correctAnswer := s.findCorrectAnswer(choices)
		gabaritoItems = append(gabaritoItems, GabaritoItem{
			Number:  questionNumber,
			Answer:  correctAnswer,
			Subject: subject.Name,
		})
		questionNumber++
	}

	return questionsWithChoices, gabaritoItems, questionNumber, nil
}

// findCorrectAnswer encontra a letra da alternativa correta
func (s *ExamService) findCorrectAnswer(choices []db.Choice) string {
	for i, choice := range choices {
		if choice.IsCorrect.Bool {
			return string(rune('A' + i))
		}
	}
	return "-"
}

// countTotalQuestions conta o total de questões
func (s *ExamService) countTotalQuestions(subjectQuestionsList []SubjectQuestions) int {
	total := 0
	for _, sq := range subjectQuestionsList {
		total += len(sq.Questions)
	}
	return total
}

// generatePDF gera o documento PDF completo
func (s *ExamService) generatePDF(subjectQuestionsList []SubjectQuestions, gabarito []GabaritoItem, totalQuestions int) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	s.buildCoverPage(pdf, tr, subjectQuestionsList, totalQuestions)
	s.buildQuestionsPages(pdf, tr, subjectQuestionsList)
	s.buildAnswerKeyPage(pdf, tr, gabarito)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("erro ao gerar buffer do PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// buildCoverPage constrói a página de capa/identificação
func (s *ExamService) buildCoverPage(pdf *gofpdf.Fpdf, tr func(string) string, subjectQuestionsList []SubjectQuestions, totalQuestions int) {
	pdf.AddPage()

	// Cabeçalho
	s.buildCoverHeader(pdf, tr)

	// Identificação do candidato
	s.buildCandidateIdentification(pdf, tr)

	// Instruções
	s.buildInstructions(pdf, tr, totalQuestions)

	// Resumo das matérias
	s.buildContentSummary(pdf, tr, subjectQuestionsList)
}

// buildCoverHeader constrói o cabeçalho da capa
func (s *ExamService) buildCoverHeader(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(190, 15, tr("AUTOBANCA"))
	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, tr("PROVA OBJETIVA"))
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 12)
	dataAtual := time.Now().Format("02/01/2006")
	pdf.Cell(190, 8, tr(fmt.Sprintf("Data: %s", dataAtual)))
	pdf.Ln(15)
}

// buildCandidateIdentification constrói a seção de identificação do candidato
func (s *ExamService) buildCandidateIdentification(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, tr("IDENTIFICAÇÃO DO CANDIDATO"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)

	// Campo Nome
	pdf.Cell(25, 8, tr("Nome:"))
	pdf.Cell(165, 8, "________________________________________________________________________")
	pdf.Ln(12)

	// Campo CPF e RG
	pdf.Cell(25, 8, tr("CPF:"))
	pdf.Cell(60, 8, "_______________________________")
	pdf.Cell(25, 8, tr("RG:"))
	pdf.Cell(60, 8, "_______________________________")
	pdf.Ln(12)

	// Campo Inscrição
	pdf.Cell(50, 8, tr("Número de Inscrição:"))
	pdf.Cell(140, 8, "_______________________________________________________")
	pdf.Ln(12)

	// Campo Cargo/Posição
	pdf.Cell(35, 8, tr("Cargo/Posição:"))
	pdf.Cell(155, 8, "______________________________________________________________")
	pdf.Ln(20)
}

// buildInstructions constrói a seção de instruções
func (s *ExamService) buildInstructions(pdf *gofpdf.Fpdf, tr func(string) string, totalQuestions int) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, tr("INSTRUÇÕES"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	instrucoes := []string{
		"1. Confira se a prova está completa e se corresponde ao cargo para o qual você se inscreveu.",
		"2. Preencha corretamente todos os campos de identificação.",
		"3. Utilize caneta esferográfica de tinta preta ou azul.",
		"4. Não é permitido o uso de corretivo, lápis ou borracha.",
		fmt.Sprintf("5. Esta prova contém %d questões objetivas.", totalQuestions),
		"6. Marque apenas uma alternativa por questão.",
		"7. As questões estão organizadas por disciplina/matéria.",
	}

	for _, instrucao := range instrucoes {
		pdf.MultiCell(190, 6, tr(instrucao), "0", "L", false)
		pdf.Ln(2)
	}

	pdf.Ln(10)
}

// buildContentSummary constrói o resumo de conteúdo da prova
func (s *ExamService) buildContentSummary(pdf *gofpdf.Fpdf, tr func(string) string, subjectQuestionsList []SubjectQuestions) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, tr("CONTEÚDO DA PROVA"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	questionStart := 1
	for _, sq := range subjectQuestionsList {
		questionEnd := questionStart + len(sq.Questions) - 1
		pdf.Cell(190, 7, tr(fmt.Sprintf("• %s: Questões %d a %d (%d questões)",
			sq.SubjectName, questionStart, questionEnd, len(sq.Questions))))
		pdf.Ln(7)
		questionStart = questionEnd + 1
	}
}

// Constantes para layout de duas colunas
const (
	columnWidth   = 90.0  // Largura de cada coluna
	columnGap     = 10.0  // Espaço entre colunas
	leftMargin    = 10.0  // Margem esquerda
	rightColStart = 105.0 // Início da coluna direita (leftMargin + columnWidth + columnGap)
	pageHeight    = 280.0 // Altura útil da página
	headerHeight  = 25.0  // Altura do cabeçalho da matéria
)

// buildQuestionsPages constrói as páginas de questões em duas colunas
func (s *ExamService) buildQuestionsPages(pdf *gofpdf.Fpdf, tr func(string) string, subjectQuestionsList []SubjectQuestions) {
	questionNumber := 1

	for _, sq := range subjectQuestionsList {
		// Nova página para cada matéria
		pdf.AddPage()
		s.buildSubjectHeader(pdf, tr, sq.SubjectName, false)

		// Controle de colunas
		currentColumn := 0 // 0 = esquerda, 1 = direita
		columnStartY := pdf.GetY()
		leftColumnY := columnStartY
		rightColumnY := columnStartY

		for _, qwc := range sq.Questions {
			// Determinar em qual coluna desenhar
			var currentX float64
			var currentY float64

			if currentColumn == 0 {
				currentX = leftMargin
				currentY = leftColumnY
			} else {
				currentX = rightColStart
				currentY = rightColumnY
			}

			// Verificar se precisa de nova página
			if currentY > pageHeight {
				if currentColumn == 0 {
					// Passou da coluna esquerda, vai para a direita
					currentColumn = 1
					currentX = rightColStart
					currentY = columnStartY
					rightColumnY = columnStartY
				} else {
					// Ambas colunas cheias, nova página
					pdf.AddPage()
					s.buildSubjectHeader(pdf, tr, sq.SubjectName, true)
					currentColumn = 0
					columnStartY = pdf.GetY()
					leftColumnY = columnStartY
					rightColumnY = columnStartY
					currentX = leftMargin
					currentY = columnStartY
				}
			}

			pdf.SetXY(currentX, currentY)
			endY := s.buildQuestionTwoColumns(pdf, tr, qwc, questionNumber, currentX)

			// Atualizar posição Y da coluna atual
			if currentColumn == 0 {
				leftColumnY = endY + 3
				// Verificar se deve mudar para coluna direita
				if leftColumnY > pageHeight {
					currentColumn = 1
				}
			} else {
				rightColumnY = endY + 3
				// Verificar se deve ir para nova página
				if rightColumnY > pageHeight {
					// Próxima questão vai para nova página
					currentColumn = 2 // Força nova página na próxima iteração
				}
			}

			// Alternar entre colunas se ainda houver espaço
			if currentColumn == 0 && rightColumnY < pageHeight {
				currentColumn = 1
			} else if currentColumn == 1 && leftColumnY < pageHeight {
				currentColumn = 0
			} else if currentColumn == 2 {
				// Nova página necessária
				pdf.AddPage()
				s.buildSubjectHeader(pdf, tr, sq.SubjectName, true)
				currentColumn = 0
				columnStartY = pdf.GetY()
				leftColumnY = columnStartY
				rightColumnY = columnStartY
			}

			questionNumber++
		}
	}
}

// buildSubjectHeader constrói o cabeçalho de uma matéria (largura total)
func (s *ExamService) buildSubjectHeader(pdf *gofpdf.Fpdf, tr func(string) string, subjectName string, isContinuation bool) {
	pdf.SetX(leftMargin)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(220, 220, 220)

	title := subjectName
	if isContinuation {
		title = fmt.Sprintf("%s (continuação)", subjectName)
	}

	pdf.CellFormat(190, 8, tr(title), "1", 0, "C", true, 0, "")
	pdf.Ln(12)
}

// buildQuestion constrói uma questão individual (versão antiga - mantida para compatibilidade)
func (s *ExamService) buildQuestion(pdf *gofpdf.Fpdf, tr func(string) string, qwc QuestionWithChoices, questionNumber int) {
	// Número e enunciado da questão
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(10, 5, tr(fmt.Sprintf("%d.", questionNumber)))

	pdf.SetFont("Arial", "", 8)
	pdf.SetX(25)
	pdf.MultiCell(175, 4, tr(qwc.Question.Statement), "0", "J", false)
	pdf.Ln(2)

	// Alternativas
	pdf.SetFont("Arial", "", 8)
	for i, choice := range qwc.Choices {
		letra := string(rune('A' + i))
		pdf.SetX(25)
		pdf.Cell(6, 4, tr(fmt.Sprintf("(%s)", letra)))
		pdf.SetX(32)
		pdf.MultiCell(165, 4, tr(choice.ChoiceText), "0", "L", false)
	}

	pdf.Ln(4)
}

// buildQuestionTwoColumns constrói uma questão em layout de duas colunas
func (s *ExamService) buildQuestionTwoColumns(pdf *gofpdf.Fpdf, tr func(string) string, qwc QuestionWithChoices, questionNumber int, startX float64) float64 {
	// Número e enunciado da questão
	pdf.SetFont("Arial", "B", 8)
	pdf.SetX(startX)
	pdf.Cell(8, 4, tr(fmt.Sprintf("%d.", questionNumber)))

	pdf.SetFont("Arial", "", 7)
	pdf.SetX(startX + 8)
	pdf.MultiCell(columnWidth-8, 3.5, tr(qwc.Question.Statement), "0", "J", false)

	// Alternativas
	pdf.SetFont("Arial", "", 7)
	for i, choice := range qwc.Choices {
		letra := string(rune('A' + i))
		pdf.SetX(startX + 3)
		pdf.Cell(5, 3.5, tr(fmt.Sprintf("(%s)", letra)))
		pdf.SetX(startX + 9)
		pdf.MultiCell(columnWidth-12, 3.5, tr(choice.ChoiceText), "0", "L", false)
	}

	pdf.Ln(2)
	return pdf.GetY()
}

// buildAnswerKeyPage constrói a página do gabarito
func (s *ExamService) buildAnswerKeyPage(pdf *gofpdf.Fpdf, tr func(string) string, gabarito []GabaritoItem) {
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 12, tr("GABARITO"))
	pdf.Ln(15)

	s.buildAnswerKeyHeader(pdf, tr)

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 245, 245)

	for i, g := range gabarito {
		// Nova página se necessário
		if pdf.GetY() > 270 {
			pdf.AddPage()
			s.buildAnswerKeyContinuationHeader(pdf, tr)
		}

		fill := i%2 == 0
		pdf.CellFormat(25, 7, fmt.Sprintf("%d", g.Number), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(25, 7, g.Answer, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(100, 7, tr(g.Subject), "1", 0, "L", fill, 0, "")
		pdf.Ln(7)
	}
}

// buildAnswerKeyHeader constrói o cabeçalho da tabela do gabarito
func (s *ExamService) buildAnswerKeyHeader(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(25, 8, tr("Questão"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, tr("Resposta"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(100, 8, tr("Disciplina"), "1", 0, "C", true, 0, "")
	pdf.Ln(8)
}

// buildAnswerKeyContinuationHeader constrói o cabeçalho de continuação do gabarito
func (s *ExamService) buildAnswerKeyContinuationHeader(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 12, tr("GABARITO (continuação)"))
	pdf.Ln(15)

	s.buildAnswerKeyHeader(pdf, tr)

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 245, 245)
}
