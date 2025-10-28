APP_NAME=recipe-api

build:
	go build -o bin/$(APP_NAME) ./cmd

run:
	go run ./cmd main.go

tidy:
	go mod tidy

fmt:
	go fmt ./...
