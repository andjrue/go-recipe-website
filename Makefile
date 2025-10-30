include .env
export

APP_NAME=recipe-api

build:
	go build -o bin/$(APP_NAME) ./cmd

run:
	go run ./cmd main.go

tidy:
	go mod tidy

fmt:
	go fmt ./...

db-status:
	@GOOSEDRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DB_STRING) goose -dir=$(GOOSE_MIGRATION_DIR) status

db-up:
	@GOOSEDRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DB_STRING) goose -dir=$(GOOSE_MIGRATION_DIR) up

db-down:
	@GOOSEDRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DB_STRING) goose -dir=$(GOOSE_MIGRATION_DIR) down

db-reset:
	@GOOSEDRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DB_STRING) goose -dir=$(GOOSE_MIGRATION_DIR) reset

db-create:
	@read -p "Enter migration name: " name; \
	goose -dir=$(GOOSE_MIGRATION_DIR) create $$name sql
