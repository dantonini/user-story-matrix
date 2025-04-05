---
file_path: docs/user-stories/setup/06-project-structure.md
created_at: 2025-03-17T08:24:19+01:00
last_updated: 2025-04-05T11:15:04+02:00
_content_hash: 551c299e46457225b7f66dd1a18bc108afa8777bd4c432bfa59059d02c04cf9e
---

# Define Project Directory Structure
[The CLI project should have a well-organized structure.]

As a developer,  
I want a structured project directory,  
so that the codebase remains maintainable and scalable.

## Acceptance criteria

- The project contains the following directory structure:
usm-cli/ 
├── cmd/ # CLI commands 
├── internal/ # Internal packages 
├── main.go # Entry point 
├── go.mod # Module dependencies 
├── go.sum # Dependency lock file 
├── README.md # Documentation 
├── .gitignore # Git ignored files
- The repository includes a `.gitignore` file with common Go exclusions.