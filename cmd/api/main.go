package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"

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

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to connect to database", "error", err)
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Inicializa o Querier do sqlc
	queries := db.New(conn)
	slog.InfoContext(ctx, "Database connection established")

	// Inicializa os Services e Handlers (arquitetura simplificada)
	disciplinaService := service.NewDisciplinaService(queries)
	assuntoService := service.NewAssuntoService(queries)
	alternativaService := service.NewAlternativaService(queries)
	questionService := service.NewQuestionService(queries)
	provaService := service.NewProvaService(queries)

	slog.InfoContext(ctx, "Initializing handlers")

	assuntoHandler := handlers.NewAssuntoHandler(ctx, assuntoService)
	alternativaHandler := handlers.NewAlternativaHandler(alternativaService)
	questionHandler := handlers.NewQuestionHandler(questionService)
	provaHandler := handlers.NewProvaHandler(provaService)

	disciplinaHandler := handlers.NewDisciplinaHandler(disciplinaService)

	// Inicializa o Router
	r := api.NewRouter(&api.RouterHandlers{
		DisciplinaHandler:  disciplinaHandler,
		AssuntoHandler:     assuntoHandler,
		AlternativaHandler: alternativaHandler,
		QuestionHandler:    questionHandler,
		ProvaHandler:       provaHandler,
	})

	slog.InfoContext(ctx, "Server executing on port 8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		slog.ErrorContext(ctx, "Error starting server", "error", err)
		log.Fatalf("Error starting server: %v	", err)
	}
}
