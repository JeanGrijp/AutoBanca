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
	SubjectHandler  *handlers.SubjectHandler
	TopicHandler    *handlers.TopicHandler
	ChoiceHandler   *handlers.ChoiceHandler
	QuestionHandler *handlers.QuestionHandler
	ExamHandler     *handlers.ExamHandler
}

func NewRouter(handlers *RouterHandlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/subjects", func(r chi.Router) {
		r.Get("/", handlers.SubjectHandler.ListSubjects)
		r.Post("/", handlers.SubjectHandler.CreateSubject)
		r.Get("/{id}", handlers.SubjectHandler.GetSubject)
		r.Put("/{id}", handlers.SubjectHandler.UpdateSubject)
		r.Delete("/{id}", handlers.SubjectHandler.DeleteSubject)
	})

	r.Route("/topics", func(r chi.Router) {
		r.Get("/", handlers.TopicHandler.ListTopics)
		r.Post("/", handlers.TopicHandler.CreateTopic)
		r.Get("/{id}", handlers.TopicHandler.GetTopic)
		r.Put("/{id}", handlers.TopicHandler.UpdateTopic)
		r.Delete("/{id}", handlers.TopicHandler.DeleteTopic)
		r.Get("/subject/{subject_id}", handlers.TopicHandler.ListTopicsBySubject)
	})

	r.Route("/choices", func(r chi.Router) {
		r.Get("/", handlers.ChoiceHandler.ListChoicesByQuestion)
		r.Post("/", handlers.ChoiceHandler.CreateChoice)
		r.Get("/{id}", handlers.ChoiceHandler.GetChoice)
		r.Put("/{id}", handlers.ChoiceHandler.UpdateChoice)
		r.Delete("/{id}", handlers.ChoiceHandler.DeleteChoice)
	})

	r.Route("/questions", func(r chi.Router) {
		r.Get("/", handlers.QuestionHandler.ListQuestionsByFilters)
		r.Post("/", handlers.QuestionHandler.CreateQuestion)
		r.Post("/import", handlers.QuestionHandler.ImportQuestionsCSV)
		r.Get("/{id}", handlers.QuestionHandler.GetQuestion)
		r.Put("/{id}", handlers.QuestionHandler.UpdateQuestion)
		r.Delete("/{id}", handlers.QuestionHandler.DeleteQuestion)
	})

	r.Route("/exams", func(r chi.Router) {
		r.Post("/", handlers.ExamHandler.GenerateExam)
	})

	// slog all routes with a for loop
	_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.InfoContext(context.Background(), "Route configured", "method", method, "route", route)
		return nil
	})

	return r
}
