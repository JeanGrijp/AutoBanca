package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
)

type ExamService struct {
	q           db.Querier
	svcSubject  *SubjectService
	svcTopic    *TopicService
	svcQuestion *QuestionService
}

type SubjectInfo struct {
	Name          string
	QuestionCount int32
}

type SubjectAndCount struct {
	Subject db.Subject
	Count   int32
}

type GenerateExamFilters struct {
	Subjects      []SubjectInfo
	Topics        *[]string
	Difficulty    pgtype.Text
	Modality      string
	PracticeArea  string
	FieldOfStudy  string
	Level         *pgtype.Text
	QuestionCount int32
}

func (gef *GenerateExamFilters) IsValid() bool {
	if (len(gef.Subjects) == 0) ||
		(gef.Difficulty == (pgtype.Text{})) ||
		(gef.Modality == "") ||
		(gef.FieldOfStudy == "") ||
		(gef.QuestionCount <= 0) {
		return false
	}
	return true
}

func NewExamService(q db.Querier, svcSubject *SubjectService, svcTopic *TopicService, svcQuestion *QuestionService) *ExamService {
	return &ExamService{q: q, svcSubject: svcSubject, svcTopic: svcTopic, svcQuestion: svcQuestion}
}

// GenerateExam generates a new exam based on the provided details.
func (s *ExamService) GenerateExam(ctx context.Context, filters GenerateExamFilters) ([]byte, error) {

	slog.InfoContext(ctx, "Generate Exam Service")
	slog.InfoContext(ctx, "Generating exam with filters")
	slog.InfoContext(ctx, "questions filters", "filters", filters)

	if !filters.IsValid() {
		slog.ErrorContext(ctx, "Invalid filters")
		return nil, fmt.Errorf("invalid filters provided")
	}

	// 2. Buscar questões do banco de dados com base nos filtros fornecidos
	var allSubjects []SubjectAndCount
	for _, subjectInfo := range filters.Subjects {
		subject, err := s.svcSubject.GetSubjectByName(ctx, subjectInfo.Name)
		if err != nil {
			slog.ErrorContext(ctx, "Error fetching subject", "error", err)
			return nil, fmt.Errorf("error fetching subject %s: %v", subjectInfo.Name, err)
		}
		slog.InfoContext(ctx, "Subject found", "subject", subject)
		allSubjects = append(allSubjects, SubjectAndCount{
			Subject: subject,
			Count:   subjectInfo.QuestionCount,
		})
	}

	var allSubjectsToExam []db.GetQuestionsForExamRow

	for _, subjectAndCount := range allSubjects {
		slog.InfoContext(ctx, "Fetching questions for subject", "subject", subjectAndCount.Subject.Name, "count", subjectAndCount.Count)
		subjectQuestions, err := s.q.GetQuestionsForExam(ctx, db.GetQuestionsForExamParams{
			Limit:      subjectAndCount.Count,
			Level:      *filters.Level,
			Difficulty: filters.Difficulty,
			TopicID:    subjectAndCount.Subject.ID,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Error fetching questions for subject", "subject", subjectAndCount.Subject.Name, "error", err)
			return nil, fmt.Errorf("error fetching questions for subject %s: %v", subjectAndCount.Subject.Name, err)
		}

		allSubjectsToExam = append(allSubjectsToExam, subjectQuestions...)
	}

	slog.InfoContext(ctx, "Number of questions fetched", "count", len(allSubjectsToExam))

	// questions, err := s.q.GetQuestionsForExam()

	// slog.InfoContext(ctx, "Quantidade de questões buscadas")

	// if err != nil {
	// 	return nil, fmt.Errorf("erro ao buscar questões no banco de dados: %v", err)
	// }

	// if len(questions) == 0 {
	// 	return nil, fmt.Errorf("nenhuma questão encontrada para os filtros fornecidos")
	// }

	// pdf := gofpdf.New("P", "mm", "A4", "")
	// pdf.AddPage()
	// pdf.SetFont("Arial", "B", 12)
	// pdf.Cell(190, 10, "Prova - AutoBanca") // Título fixo da banca
	// pdf.Ln(12)

	// pdf.SetFont("Arial", "", 12)

	// for i, q := range questions {
	// 	texto := fmt.Sprintf("%d) %s", i+1, q.Enunciado)
	// 	// MultiCell é melhor para enunciados longos, pois quebra a linha automaticamente
	// 	pdf.MultiCell(190, 8, texto, "0", "L", false)
	// 	pdf.Ln(4)
	// }

	// // 4. Retorno como Buffer (para facilitar o envio via HTTP ou salvar em disco)
	// var buf bytes.Buffer
	// err = pdf.Output(&buf)
	// if err != nil {
	// 	return nil, fmt.Errorf("erro ao gerar buffer do PDF: %v", err)
	// }

	// return buf.Bytes(), nil
	return nil, fmt.Errorf("função GenerateExam ainda não implementada")

}
