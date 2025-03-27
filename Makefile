.PHONY: build test clean run demo-tui

# Binary name
BINARY_NAME=usm
VERSION=0.1.0

# Build the binary
build:
	go build -o $(BINARY_NAME) -v

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html

# Create a coverage report showing uncovered lines
coverage-report:
	@echo "Generating coverage report..."
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep -v "100.0%" | sort -k 3 -r
	@echo "HTML report available with: make coverage-html"

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64-$(VERSION)
	rm -f $(BINARY_NAME)-darwin-amd64-$(VERSION)
	rm -f $(BINARY_NAME)-darwin-arm64-$(VERSION)
	rm -f $(BINARY_NAME)-windows-amd64-$(VERSION).exe
	rm -f coverage.out coverage.html

# Run the application
run:
	go run main.go

# Build for all platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64-$(VERSION) -v
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64-$(VERSION) -v
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64-$(VERSION) -v
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64-$(VERSION).exe -v

# Install dependencies
deps:
	go mod tidy

# Demo applications
demo-tui:
	@echo "Building and running TUI demo..."
	@go run internal/ui/pages/cmd/main.go

# Default target
all: clean build 