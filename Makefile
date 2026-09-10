DB_URL ?= postgres://postgres:postgres@localhost:5435/auth?sslmode=disable

.PHONY: run test lint format tidy compose-up compose-down migrate-up migrate-down migrate-create

run:
	swag init -g main.go -o docs --parseInternal
	go run .

test:
	go test ./... -cover

format:
	golangci-lint fmt

lint:
	golangci-lint run

tidy:
	go mod tidy

docs:
	swag init -g main.go -o docs --parseInternal

compose-up:
	docker compose up -d --wait --remove-orphans

compose-down:
	docker compose down

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<snake>"; exit 1)
	migrate create -ext sql -dir migrations "$(name)"
