SHELL := /bin/sh

.PHONY: run test generate

run:
	go run ./cmd/api

test:
	go test ./...

generate:
	sqlc generate
