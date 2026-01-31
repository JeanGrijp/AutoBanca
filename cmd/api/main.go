package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JeanGrijp/AutoBanca/internal/adapter/database/sqlc/db"
	"github.com/JeanGrijp/AutoBanca/internal/api"
	"github.com/JeanGrijp/AutoBanca/internal/handlers"
	"github.com/JeanGrijp/AutoBanca/internal/service"
)

func main() {
	ctx := context.Background()

	slog.InfoContext(ctx, "Starting server")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	slog.InfoContext(ctx, "Connecting to database", "host", host, "port", port, "user", user, "dbname", dbname)

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to connect to database", "error", err)
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Inicializa o Querier do sqlc
	queries := db.New(pool)
	slog.InfoContext(ctx, "Database connection established")

	// Inicializa os Services e Handlers (arquitetura simplificada)
	subjectService := service.NewSubjectService(queries)
	topicService := service.NewTopicService(queries)
	choiceService := service.NewChoiceService(queries)
	questionService := service.NewQuestionService(queries)
	importService := service.NewImportService(pool)
	examService := service.NewExamService(queries, subjectService, topicService, questionService)

	slog.InfoContext(ctx, "Initializing handlers")

	topicHandler := handlers.NewTopicHandler(ctx, topicService)
	choiceHandler := handlers.NewChoiceHandler(choiceService)
	questionHandler := handlers.NewQuestionHandler(questionService, choiceService, importService)
	examHandler := handlers.NewExamHandler(examService)

	subjectHandler := handlers.NewSubjectHandler(subjectService)

	// Inicializa o Router
	r := api.NewRouter(&api.RouterHandlers{
		SubjectHandler:  subjectHandler,
		TopicHandler:    topicHandler,
		ChoiceHandler:   choiceHandler,
		QuestionHandler: questionHandler,
		ExamHandler:     examHandler,
	})

	slog.InfoContext(ctx, "Server executing on port 8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		slog.ErrorContext(ctx, "Error starting server", "error", err)
		log.Fatalf("Error starting server: %v	", err)
	}
}
