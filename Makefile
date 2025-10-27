APP_NAME=recipe-api

build:
	go build -o bin/$(APP_NAME) ./cmd

run: 
	run ./cmd

tidy: 
	go mod tidy

fmt:
	got fmt ./...
