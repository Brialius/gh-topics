# Makefile for gh-topics

# Variables
BINARY_NAME=gh-topics
GO=go
GOFLAGS=-trimpath
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
GOLANGCI_LINT=$(shell which golangci-lint)
GO_FILES=$(shell find . -type f -name "*.go" -not -path "./vendor/*")
COVERAGE_FILE=coverage.txt

# Override for a release; dev builds use embedded Go VCS metadata
VERSION ?= dev

# Check if Go version is at least 1.22
GO_VERSION_CHECK=$(shell $(GO) version | grep -q "go1.2[2-9]" && echo "pass" || echo "fail")

.PHONY: all build test lint clean run install deps help

all: lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@if [ "$(GO_VERSION_CHECK)" = "fail" ]; then \
		echo "Error: Go 1.22 or higher is required"; \
		exit 1; \
	fi
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests with race detection and coverage
test:
	@echo "Running tests..."
	$(GO) test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

# Run linter
lint:
	@echo "Running linter..."
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found, installing..."; \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0; \
	fi
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -rf dist/

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run .

# Install the application locally
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) $(LDFLAGS) .

# Download and update dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	@echo "Verifying dependencies..."
	$(GO) mod verify

# Show help
help:
	@echo "Available targets:"
	@echo "  all      - Run lint, test, and build"
	@echo "  build    - Build the application"
	@echo "  test     - Run tests with race detection and coverage"
	@echo "  lint     - Run linter"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Run the application"
	@echo "  install  - Install the application locally"
	@echo "  deps     - Download and update dependencies"
	@echo "  help     - Show this help message"
	@echo ""
	@echo "Current version: $(VERSION)"
	@echo "Go version check: $(GO_VERSION_CHECK)"

# Default target
.DEFAULT_GOAL := help
