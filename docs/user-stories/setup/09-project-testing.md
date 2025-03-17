# Set Up Unit Testing Framework
The project should include a testing framework.

As a developer,  
I want to set up unit testing for the CLI,  
so that I can ensure features work correctly.

## Acceptance criteria

- The CLI includes Go’s built-in `testing` package.
- A sample test (`usm_test.go`) is created for basic functionality.
- Running `go test ./...` executes all tests.
- The repository includes a `Makefile` target (`make test`) to run tests.