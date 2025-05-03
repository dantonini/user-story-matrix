.PHONY: build test clean run demo-tui lint lint-tests lint-ci build-full install-hooks lint-report deadcode deadcode-ast deadcode-remove

# Binary name
BINARY_NAME=usm
VERSION=0.1.5
COVERAGE_TOOL=coverage

# Detect golangci-lint version for compatibility
GOLANGCI_VERSION := $(shell golangci-lint --version 2>/dev/null | grep -o 'version [0-9.]*' | sed 's/version //' || echo "0.0.0")
SUPPORTS_CACHE := $(shell echo "$(GOLANGCI_VERSION)" | awk -F. '{ if ($$1 > 1 || ($$1 == 1 && $$2 >= 54)) print "true"; else print "false"; }')

# Determine which linter to use for dead code detection
ifeq ($(shell echo "$(GOLANGCI_VERSION)" | awk -F. '{ if ($$1 > 1 || ($$1 == 1 && $$2 >= 49)) print "true"; else print "false"; }'),true)
    DEADCODE_LINTER=unused
else
    DEADCODE_LINTER=deadcode
endif

# Extract Go version - safely handle non-numeric parts
GO_VERSION := $(shell go version | grep -o 'go[0-9.]*' | sed 's/go//' || echo "0.0.0")
GO_VERSION_SAFE := $(shell echo "$(GO_VERSION)" | awk -F. '{ if ($$1 >= 1 && $$2 >= 22) print "true"; else print "false"; }')

# For newer Go versions, avoid staticcheck to prevent issues
ifeq ($(GO_VERSION_SAFE),true)
	SAFE_LINTERS=errcheck,govet,$(DEADCODE_LINTER)
else
	SAFE_LINTERS=errcheck,govet,$(DEADCODE_LINTER),staticcheck
endif

# Cache flags based on version
ifeq ($(SUPPORTS_CACHE),true)
	CACHE_FLAG=--cache
else
	CACHE_FLAG=
endif

# Install golangci-lint if needed
define ensure_golangci_lint
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2; \
	fi
endef

# Build the binary
build:
	go build -o $(BINARY_NAME) -v

# Lint the code without building
lint: lint-clean
	@echo "Linting completed"

# Lint only test files
lint-tests:
	@echo "Running linters on test files only..."
	$(call ensure_golangci_lint)
	@cd cmd && golangci-lint run $(CACHE_FLAG) --tests=true --skip-files="^[^_]*\.go$$" ./... | grep -v "output/"
	@cd internal && golangci-lint run $(CACHE_FLAG) --tests=true --skip-files="^[^_]*\.go$$" ./... | grep -v "output/"

# Lint for CI environments
lint-ci:
	@echo "Running linters for CI..."
	$(call ensure_golangci_lint)
	@golangci-lint run --timeout=5m --out-format=colored-line-number $(CACHE_FLAG) ./... | grep -v "output/" || echo "Linting found issues in CI, but continuing"

# Generate lint report
lint-report:
	@echo "Generating lint report..."
	$(call ensure_golangci_lint)
	@mkdir -p output/reports
	golangci-lint run --timeout=5m --out-format=json $(CACHE_FLAG) ./... | grep -v "output/" > output/reports/lint-report.json || true
	@echo "Report saved to output/reports/lint-report.json"
	@echo "Summary of issues:"
	@cat output/reports/lint-report.json | grep -o '"Pos":{"Filename":"[^"]*"' | grep -v "output/" | sort | uniq -c | sort -nr || true

# Build with lint checks (full build)
build-full: 
	@echo "Running linters..."
	$(call ensure_golangci_lint)
	@echo "Using configuration from .golangci.yml"
	@golangci-lint run $(CACHE_FLAG) --timeout=2m ./... | grep -v "output/" || true
	@echo "Running full build with linting..."
	go build -o $(BINARY_NAME) -v

build-coverage:
	go build -o $(COVERAGE_TOOL) tools/coverage/main.go
# Run tests with coverage
test: build-coverage
	go test -v -coverprofile=coverage.out -covermode=atomic $(shell go list ./... | grep -v /output/)
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html

coverage: build-coverage
	./$(COVERAGE_TOOL) --cov coverage.out

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64-$(VERSION)
	rm -f $(BINARY_NAME)-darwin-amd64-$(VERSION)
	rm -f $(BINARY_NAME)-darwin-arm64-$(VERSION)
	rm -f $(BINARY_NAME)-windows-amd64-$(VERSION).exe
	rm -f $(COVERAGE_TOOL)
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

# Install pre-commit hooks
install-hooks:
	@echo "Installing pre-commit hooks..."
	@mkdir -p .git/hooks
	@cp hooks/pre-commit .git/hooks/
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed successfully!"

# Demo applications
demo-tui:
	@echo "Building and running TUI demo..."
	@go run internal/ui/pages/cmd/main.go

# Special lint command that explicitly only lints specified directories
lint-clean:
	@echo "Running linters on main source code only..."
	$(call ensure_golangci_lint)
	@set -e; \
	echo "Checking cmd directory..."; \
	(cd cmd && golangci-lint run $(CACHE_FLAG) ./...); \
	echo "Checking internal directory..."; \
	(cd internal && golangci-lint run $(CACHE_FLAG) ./...); \
	echo "Checking main.go..."; \
	golangci-lint run $(CACHE_FLAG) main.go; \
	echo "All linting checks passed!"

# Run the deadcode detection and removal tool
deadcode:
	@echo "Running dead code detection and removal (AST-based)..."
	@chmod +x scripts/deadcode.sh
	@scripts/deadcode.sh --dry-run

# Run the deadcode tool with actual removal
deadcode-remove:
	@echo "Running dead code removal (AST-based)..."
	@chmod +x scripts/deadcode.sh
	@scripts/deadcode.sh

# Run the AST-based dead code removal tool (safer and more comprehensive)
deadcode-ast:
	@echo "Running AST-based dead code remover (dry run)..."
	@go run ./cmd/deadcode --dry-run
	@echo "For actual code removal, run: go run ./cmd/deadcode"

# Default target
all: clean build 