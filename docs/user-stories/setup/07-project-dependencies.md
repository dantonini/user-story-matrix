# Set Up Dependency Management
The project should have a proper dependency management system.

As a developer,  
I want to manage dependencies using Go modules,  
so that I can keep the project clean and reproducible.

## Acceptance criteria

- The project initializes `go mod` with the correct module name.
- Dependencies are defined in `go.mod` and locked in `go.sum`.
- The repository includes instructions on how to install dependencies.
- Running `go mod tidy` cleans up unused dependencies.