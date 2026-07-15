APP_NAME=recipe-api
MIGRATIONS_DIR=internal/database/migrations

# pull DB_* vars in from .env if present (dash = don't fail when missing)
-include .env
GOOSE_DBSTRING="host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)"

build:
	go build -o bin/$(APP_NAME) ./cmd

run:
	go run ./cmd

tidy:
	go mod tidy

fmt:
	go fmt ./...

test:
	go test ./...

frontend-install:
	npm --prefix frontend install

frontend-dev:
	npm --prefix frontend run dev

frontend-test:
	npm --prefix frontend test

frontend-build:
	npm --prefix frontend run build

compose-up:
	docker compose up -d --build

db-up:
	docker compose up -d postgres

api-up:
	docker compose up -d --build api

api-logs:
	docker compose logs -f api

db-down:
	docker compose down

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres $(GOOSE_DBSTRING) up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres $(GOOSE_DBSTRING) down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres $(GOOSE_DBSTRING) status

# usage: make migrate-create name=create_recipes
migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql
