SHELL := /bin/sh

.PHONY: build test lint run

build:
	go build ./...

test:
	go test ./... -race

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		test -z "$$(gofmt -l .)" || (echo "gofmt required for:"; gofmt -l .; exit 1); \
		go vet ./...; \
	fi

run:
	STORE=$${STORE:-memory} go run ./cmd/api
