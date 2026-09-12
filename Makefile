.PHONY: build test lint run

build:
	go build ./...

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

run:
	go run main.go
