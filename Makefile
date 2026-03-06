.PHONY: build test lint clean install help

# Build variables
BINARY_NAME := a6
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -X github.com/api7/a6/internal/version.Version=$(VERSION) \
           -X github.com/api7/a6/internal/version.Commit=$(COMMIT) \
           -X github.com/api7/a6/internal/version.Date=$(BUILD_DATE)

# Go variables
GOBIN ?= $(shell go env GOPATH)/bin
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

## build: Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/a6

## install: Install the binary to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/a6

## test: Run all tests
test:
	go test -race -coverprofile=coverage.out ./...

## test-verbose: Run all tests with verbose output
test-verbose:
	go test -race -v -coverprofile=coverage.out ./...

## coverage: Show test coverage in browser
coverage: test
	go tool cover -html=coverage.out

## lint: Run linters
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	gofmt -s -w .
	goimports -w .

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
