include .env
export

APP_NAME=recipe-api

build:
	@go build -o bin/$(APP_NAME) ./cmd

run:
	@echo "Starting server w/ db url: DB_URL=${DB_URL}"
	@go run ./cmd main.go

tidy:
	@go mod tidy

fmt:
	@go fmt ./...

vet:
	@go vet ./...

test:
	go test -v ./...

fmt-check:
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "Code is not formatted. Run 'make fmt'"; \
		gofmt -s -l .; \
		exit 1; \
	fi

pre-push: fmt-check vet build
	@echo "All checks passed! Safe to push."

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
