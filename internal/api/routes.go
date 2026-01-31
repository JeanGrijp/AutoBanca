// Copyright 2024 Jean Grijp. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package api sets up the HTTP routes for the AutoBanca application.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/JeanGrijp/AutoBanca/internal/handlers"
)

type RouterHandlers struct {
	DisciplinaHandler  *handlers.DisciplinaHandler
	AssuntoHandler     *handlers.AssuntoHandler
	AlternativaHandler *handlers.AlternativaHandler
	QuestionHandler    *handlers.QuestionHandler
	ProvaHandler       *handlers.ProvaHandler
}

func NewRouter(handlers *RouterHandlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/disciplinas", func(r chi.Router) {
		r.Get("/", handlers.DisciplinaHandler.ListDisciplinas)
		r.Post("/", handlers.DisciplinaHandler.CreateDisciplina)
		r.Get("/{id}", handlers.DisciplinaHandler.GetDisciplina)
		r.Put("/{id}", handlers.DisciplinaHandler.UpdateDisciplina)
		r.Delete("/{id}", handlers.DisciplinaHandler.DeleteDisciplina)
	})

	r.Route("/assuntos", func(r chi.Router) {
		r.Get("/", handlers.AssuntoHandler.ListAssuntos)
		r.Post("/", handlers.AssuntoHandler.CreateAssunto)
		r.Get("/{id}", handlers.AssuntoHandler.GetAssunto)
		r.Put("/{id}", handlers.AssuntoHandler.UpdateAssunto)
		r.Delete("/{id}", handlers.AssuntoHandler.DeleteAssunto)
		r.Get("/disciplina/{disciplina_id}", handlers.AssuntoHandler.ListAssuntosByDisciplina)
	})

	r.Route("/alternativas", func(r chi.Router) {
		r.Get("/", handlers.AlternativaHandler.ListAlternativasByQuestion)
		r.Post("/", handlers.AlternativaHandler.CreateAlternativa)
		r.Get("/{id}", handlers.AlternativaHandler.GetAlternativa)
		r.Put("/{id}", handlers.AlternativaHandler.UpdateAlternativa)
		r.Delete("/{id}", handlers.AlternativaHandler.DeleteAlternativa)
	})

	r.Route("/questions", func(r chi.Router) {
		r.Get("/", handlers.QuestionHandler.ListQuestionsByFilters)
		r.Post("/", handlers.QuestionHandler.CreateQuestion)
		r.Get("/{id}", handlers.QuestionHandler.GetQuestion)
		r.Put("/{id}", handlers.QuestionHandler.UpdateQuestion)
		r.Delete("/{id}", handlers.QuestionHandler.DeleteQuestion)
	})

	r.Route("/provas", func(r chi.Router) {
		r.Post("/", handlers.ProvaHandler.GenerateProva)
	})

	// slog all routes with a for loop
	_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.InfoContext(context.Background(), "Route configured", "method", method, "route", route)
		return nil
	})

	return r
}
