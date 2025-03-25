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

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64-$(VERSION)
	rm -f $(BINARY_NAME)-darwin-amd64-$(VERSION)
	rm -f $(BINARY_NAME)-windows-amd64-$(VERSION).exe

# Run the application
run:
	go run main.go

# Build for all platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64-$(VERSION) -v
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64-$(VERSION) -v
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