APP_NAME=recipe-api
MIGRATIONS_DIR=internal/database/migrations
BACKUP_DIR=backups

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

db-backup:
	mkdir -p $(BACKUP_DIR)
	docker compose exec -T postgres pg_dump --username $(DB_USER) --dbname $(DB_NAME) --format=custom > $(BACKUP_DIR)/recipe-$$(date +%Y%m%d-%H%M%S).dump

images-backup:
	mkdir -p $(BACKUP_DIR)
	docker compose exec -T api tar -czf - -C /data images > $(BACKUP_DIR)/images-$$(date +%Y%m%d-%H%M%S).tar.gz

backup: db-backup images-backup

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres $(GOOSE_DBSTRING) up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres $(GOOSE_DBSTRING) down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres $(GOOSE_DBSTRING) status

# usage: make migrate-create name=create_recipes
migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql
