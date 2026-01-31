package main

import (
	"context"
	"fmt"
	"log"
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

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Inicializa o Querier do sqlc
	queries := db.New(conn)

	// Inicializa os Services e Handlers (arquitetura simplificada)
	disciplinaService := service.NewDisciplinaService(queries)
	assuntoService := service.NewAssuntoService(queries)

	assuntoHandler := handlers.NewAssuntoHandler(ctx, assuntoService)

	disciplinaHandler := handlers.NewDisciplinaHandler(disciplinaService)

	// Inicializa o Router
	r := api.NewRouter(&api.RouterHandlers{
		DisciplinaHandler: disciplinaHandler,
		AssuntoHandler:    assuntoHandler,
	})

	fmt.Println("Server executing on port 8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
