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

type ExamService struct {
	q           db.Querier
	svcSubject  *SubjectService
	svcTopic    *TopicService
	svcQuestion *QuestionService
}

type SubjectFilter struct {
	Name          string   `json:"name"`
	QuestionCount int32    `json:"question_count"`
	Topics        []string `json:"topics,omitempty"` // opcional - nomes dos tópicos específicos
}

type SubjectAndCount struct {
	Subject db.Subject
	Count   int32
	Topics  []string // tópicos específicos para filtrar
}

type GenerateExamFilters struct {
	Subjects     []SubjectFilter `json:"subjects"`       // obrigatório
	Difficulty   pgtype.Text     `json:"difficulty"`     // opcional: easy/medium/hard
	Level        pgtype.Text     `json:"level"`          // opcional
	Modality     pgtype.Text     `json:"modality"`       // obrigatório: multiple_choice/true_false
	Position     pgtype.Text     `json:"position"`       // opcional: cargo específico
	FieldOfStudy pgtype.Text     `json:"field_of_study"` // obrigatório: law/medicine/engineering
	MinYear      pgtype.Int4     `json:"min_year"`       // opcional: ano mínimo
	MaxYear      pgtype.Int4     `json:"max_year"`       // opcional: ano máximo
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

func (gef *GenerateExamFilters) IsValid() bool {
	if len(gef.Subjects) == 0 {
		return false
	}
	// Verificar se cada subject tem pelo menos 1 questão
	for _, s := range gef.Subjects {
		if s.QuestionCount <= 0 || s.Name == "" {
			return false
		}
	}
	// Modality e FieldOfStudy são obrigatórios
	if !gef.Modality.Valid || gef.Modality.String == "" {
		return false
	}
	if !gef.FieldOfStudy.Valid || gef.FieldOfStudy.String == "" {
		return false
	}
	return true
}

func NewExamService(q db.Querier, svcSubject *SubjectService, svcTopic *TopicService, svcQuestion *QuestionService) *ExamService {
	return &ExamService{q: q, svcSubject: svcSubject, svcTopic: svcTopic, svcQuestion: svcQuestion}
}

// GenerateExam generates a new exam based on the provided details.
func (s *ExamService) GenerateExam(ctx context.Context, filters GenerateExamFilters) ([]byte, error) {

	slog.InfoContext(ctx, "-----------------------------")
	slog.InfoContext(ctx, "Generate Exam Service")
	slog.InfoContext(ctx, "-----------------------------")

	slog.InfoContext(ctx, "Generating exam with filters")
	slog.InfoContext(ctx, "questions filters", "filters", filters)

	if !filters.IsValid() {
		slog.ErrorContext(ctx, "Invalid filters")
		return nil, fmt.Errorf("invalid filters provided")
	}

	// 1. Buscar questões do banco de dados com base nos filtros fornecidos
	// Agrupadas por matéria
	var subjectQuestionsList []SubjectQuestions
	var gabarito []struct {
		Number  int
		Answer  string
		Subject string
	}
	questionNumber := 1

	for _, subjectFilter := range filters.Subjects {
		slog.InfoContext(ctx, "Fetching subject", "name", subjectFilter.Name)
		subject, err := s.svcSubject.GetSubjectByName(ctx, subjectFilter.Name)
		if err != nil {
			slog.ErrorContext(ctx, "Error fetching subject", "error", err)
			return nil, fmt.Errorf("error fetching subject %s: %v", subjectFilter.Name, err)
		}
		slog.InfoContext(ctx, "Subject found", "subject", subject)

		// TODO: Se Topics não for vazio, buscar TopicID para cada tópico
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
			return nil, fmt.Errorf("error fetching questions for subject %s: %v", subject.Name, err)
		}

		slog.InfoContext(ctx, "Questions fetched for subject", "subject", subject.Name, "count", len(questions))

		// Buscar alternativas para cada questão
		var questionsWithChoices []QuestionWithChoices
		for _, q := range questions {
			choices, err := s.q.ListChoicesByQuestion(ctx, q.ID)
			if err != nil {
				slog.ErrorContext(ctx, "Error fetching choices for question", "question_id", q.ID, "error", err)
				return nil, fmt.Errorf("error fetching choices for question: %v", err)
			}

			questionsWithChoices = append(questionsWithChoices, QuestionWithChoices{
				Question: q,
				Choices:  choices,
			})

			// Encontrar resposta correta para o gabarito
			correctAnswer := "-"
			for i, choice := range choices {
				if choice.IsCorrect.Bool {
					correctAnswer = string(rune('A' + i))
					break
				}
			}
			gabarito = append(gabarito, struct {
				Number  int
				Answer  string
				Subject string
			}{
				Number:  questionNumber,
				Answer:  correctAnswer,
				Subject: subject.Name,
			})
			questionNumber++
		}

		if len(questionsWithChoices) > 0 {
			subjectQuestionsList = append(subjectQuestionsList, SubjectQuestions{
				SubjectName: subject.Name,
				Questions:   questionsWithChoices,
			})
		}
	}

	// Contar total de questões
	totalQuestions := 0
	for _, sq := range subjectQuestionsList {
		totalQuestions += len(sq.Questions)
	}

	slog.InfoContext(ctx, "Total questions fetched", "count", totalQuestions)

	if totalQuestions == 0 {
		slog.ErrorContext(ctx, "Nenhuma questão encontrada para os filtros fornecidos", "filters", filters)
		return nil, fmt.Errorf("nenhuma questão encontrada para os filtros fornecidos")
	}

	// 2. Gerar PDF com as questões
	slog.InfoContext(ctx, "Gerando PDF com as questões")

	pdf := gofpdf.New("P", "mm", "A4", "")

	// Configurar tradução de caracteres para suporte a acentos
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// ========== PÁGINA DE CAPA / IDENTIFICAÇÃO ==========
	pdf.AddPage()

	// Cabeçalho da banca
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(190, 15, tr("AUTOBANCA"))
	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, tr("PROVA OBJETIVA"))
	pdf.Ln(15)

	// Data da prova
	pdf.SetFont("Arial", "", 12)
	dataAtual := time.Now().Format("02/01/2006")
	pdf.Cell(190, 8, tr(fmt.Sprintf("Data: %s", dataAtual)))
	pdf.Ln(15)

	// Informações do candidato
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

	// Instruções
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

	// Resumo das matérias
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

	// ========== PÁGINAS DE QUESTÕES ==========
	questionNumber = 1
	for _, sq := range subjectQuestionsList {
		// Nova página para cada matéria
		pdf.AddPage()

		// Cabeçalho da matéria
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(190, 10, tr(sq.SubjectName), "1", 0, "C", true, 0, "")
		pdf.Ln(15)

		for _, qwc := range sq.Questions {
			// Verificar se precisa de nova página (margem de 60mm para questão + alternativas)
			if pdf.GetY() > 230 {
				pdf.AddPage()
				// Repetir cabeçalho da matéria
				pdf.SetFont("Arial", "B", 14)
				pdf.SetFillColor(220, 220, 220)
				pdf.CellFormat(190, 10, tr(fmt.Sprintf("%s (continuação)", sq.SubjectName)), "1", 0, "C", true, 0, "")
				pdf.Ln(15)
			}

			// Número e enunciado da questão
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(15, 7, tr(fmt.Sprintf("%d.", questionNumber)))

			pdf.SetFont("Arial", "", 11)
			// Usar MultiCell para enunciados longos
			startY := pdf.GetY()
			pdf.SetX(25)
			pdf.MultiCell(175, 6, tr(qwc.Question.Statement), "0", "J", false)
			pdf.Ln(3)

			// Alternativas
			pdf.SetFont("Arial", "", 10)
			for i, choice := range qwc.Choices {
				letra := string(rune('A' + i))
				pdf.SetX(25)
				pdf.Cell(8, 6, tr(fmt.Sprintf("(%s)", letra)))
				pdf.SetX(35)
				pdf.MultiCell(165, 6, tr(choice.ChoiceText), "0", "L", false)
				pdf.Ln(1)
			}

			pdf.Ln(8)
			questionNumber++

			_ = startY // evitar warning de variável não utilizada
		}
	}

	// ========== PÁGINA DE GABARITO ==========
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 12, tr("GABARITO"))
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)

	// Cabeçalho da tabela de gabarito
	pdf.CellFormat(25, 8, tr("Questão"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, tr("Resposta"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(100, 8, tr("Disciplina"), "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 245, 245)
	for i, g := range gabarito {
		fill := i%2 == 0
		pdf.CellFormat(25, 7, fmt.Sprintf("%d", g.Number), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(25, 7, g.Answer, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(100, 7, tr(g.Subject), "1", 0, "L", fill, 0, "")
		pdf.Ln(7)

		// Nova página se necessário
		if pdf.GetY() > 270 {
			pdf.AddPage()
			pdf.SetFont("Arial", "B", 16)
			pdf.Cell(190, 12, tr("GABARITO (continuação)"))
			pdf.Ln(15)

			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(200, 200, 200)
			pdf.CellFormat(25, 8, tr("Questão"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, tr("Resposta"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(100, 8, tr("Disciplina"), "1", 0, "C", true, 0, "")
			pdf.Ln(8)

			pdf.SetFont("Arial", "", 10)
			pdf.SetFillColor(245, 245, 245)
		}
	}

	// 3. Retorno como Buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		slog.ErrorContext(ctx, "Erro ao gerar buffer do PDF", "error", err)
		return nil, fmt.Errorf("erro ao gerar buffer do PDF: %v", err)
	}

	slog.InfoContext(ctx, "PDF gerado com sucesso!!!", "total_questions", totalQuestions)

	return buf.Bytes(), nil
}
