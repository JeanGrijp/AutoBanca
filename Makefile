.PHONY: help build run logs dev docker-up docker-down docker-logs docker-watch db-init

help:
	@echo "AutoBanca - Available commands:"
	@echo "  make build        - Build the Go application"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Build, run and show logs (development mode)"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make docker-logs  - Show Docker logs"
	@echo "  make docker-watch - Start Docker containers and watch logs"
	@echo "  make db-init      - Initialize database schema"

build:
	@echo "Building AutoBanca..."
	@cd cmd/api && go build -o ../../autobanca main.go
	@echo "Build complete: ./autobanca"

build-app:
	@echo "Building AutoBanca application..."
	@docker-compose build autobanca_app
	@echo "Build complete."

run:
	@echo "Starting AutoBanca..."
	@./autobanca

dev: build run

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Waiting for services to start..."
	@sleep 2
	@echo "Services started. Run 'make docker-logs' to see logs."

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

docker-watch:
	@echo "Down all Docker containers..."
	@make docker-down
	@echo "Starting Docker containers..."
	@make docker-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Building the application..."
	@make build-app
	@echo "All systems ready! Run 'make run' in another terminal to start the app"
	@echo "Watching the containers. Press Ctrl+C to stop."
	@make docker-logs

db-init:
	@echo "Initializing database schema..."
	@docker-compose up -d autobanca_db
	@echo "Waiting for database to be ready..."
	@until docker-compose exec autobanca_db pg_isready -U autobanca_user; do sleep 1; done
	@docker-compose exec -T autobanca_db psql -U autobanca_user -d autobanca_db -f - < ./internal/adapter/database/sqlc/schema/schema.sql