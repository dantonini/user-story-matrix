# Create Basic CLI Command Structure
The CLI should have a structured way to handle commands.

As a developer,  
I want a command structure for usm-cli,  
so that I can add new commands easily.

## Acceptance criteria

- The CLI uses the `cobra` package for command management.
- A root command (`usm`) is available.
- Running `usm --help` shows usage instructions.
- The CLI has a placeholder command (`usm example`) that prints "Hello, USM!".