.PHONY: run test lint format tidy

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
