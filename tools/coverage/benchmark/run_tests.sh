#!/bin/bash

# Run tests with coverage
go test -coverprofile=coverage.out

# Generate HTML report for reference
go tool cover -html=coverage.out -o coverage.html

# Print message
echo "Coverage data generated in coverage.out"
echo "HTML report generated in coverage.html"
echo
echo "Run the coverage analysis tool with:"
echo "go run ../main.go -cov coverage.out -file coverage_examples.go" 