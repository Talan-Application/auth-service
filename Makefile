.PHONY: build run test lint tidy docker-build docker-up docker-down migrate-up migrate-down

APP      := auth-service
BIN      := bin/$(APP)
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable

build:
	go build -ldflags="-s -w" -o $(BIN) ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-build:
	docker build -t $(APP) .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down
