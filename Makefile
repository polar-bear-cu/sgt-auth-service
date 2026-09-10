.PHONY: run test lint format tidy compose-up compose-down

run:
	go run .

test:
	go test ./... -cover

format:
	golangci-lint fmt

lint:
	golangci-lint run

tidy:
	go mod tidy

compose-up:
	docker compose up --build -d --wait

compose-down:
	docker compose down
