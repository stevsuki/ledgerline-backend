-include .env
export

BINARY      := bin/api
MAIN        := ./cmd/api

# Recursive '=' so the DB_* values from .env are read at use time, not before the include.
MIGRATE_DSN  = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: help
help: ## Show the list of commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## Install supporting tools (swag, migrate, golangci-lint, air)
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest

.PHONY: run
run: ## Run the application
	go run $(MAIN)

.PHONY: dev
dev: ## Run with hot reload (air)
	air

.PHONY: build
build: swag ## Build the binary into bin/
	go build -trimpath -ldflags="-s -w" -o $(BINARY) $(MAIN)

.PHONY: test
test: ## Run unit tests
	go test ./... -race -count=1

.PHONY: test-cover
test-cover: ## Unit tests + HTML coverage report
	go test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Report: coverage.html"

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format the code
	go fmt ./...
	go vet ./...

.PHONY: tidy
tidy: ## Tidy up dependencies
	go mod tidy

.PHONY: swag
swag: ## Generate the Swagger documentation
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

.PHONY: migrate-up
migrate-up: ## Run all migrations
	migrate -path db/migrations -database "$(MIGRATE_DSN)" up

.PHONY: migrate-down
migrate-down: ## Roll back one migration
	migrate -path db/migrations -database "$(MIGRATE_DSN)" down 1

.PHONY: migrate-create
migrate-create: ## Create a new migration: make migrate-create name=create_xxx_table
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: migrate-version
migrate-version: ## Show the migration version
	migrate -path db/migrations -database "$(MIGRATE_DSN)" version

.PHONY: migrate-force
migrate-force: ## Force the migration version: make migrate-force version=1
	migrate -path db/migrations -database "$(MIGRATE_DSN)" force $(version)

.PHONY: docker-up
docker-up: ## Start the docker compose stack
	docker compose up -d --build

.PHONY: docker-down
docker-down: ## Stop the docker compose stack
	docker compose down -v
