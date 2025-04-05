---
file_path: docs/user-stories/setup/09-project-testing.md
created_at: 2025-03-17T08:26:17+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: ac1065958fcd3b7e3fc433bf634462dadd0cc2307500864cba42a8c7740ea0a2
---

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