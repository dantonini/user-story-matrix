package changerequest

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
)

// FindIncomplete finds all change requests that have a blueprint file
// but no implementation file
func FindIncomplete(fs io.FileSystem) ([]models.ChangeRequest, error) {
	var incompleteChangeRequests []models.ChangeRequest

	// Define the change requests directory
	changeRequestsDir := "docs/changes-request"

	// Check if the directory exists
	if !fs.Exists(changeRequestsDir) {
		return incompleteChangeRequests, fmt.Errorf("change requests directory not found: %s", changeRequestsDir)
	}

	// Get all files in the directory
	entries, err := fs.ReadDir(changeRequestsDir)
	if err != nil {
		return incompleteChangeRequests, fmt.Errorf("failed to read directory: %s", err)
	}

	// Iterate through the files to find blueprint files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".blueprint.md") {
			continue
		}

		// Load the blueprint file
		blueprintPath := filepath.Join(changeRequestsDir, filename)
		content, err := fs.ReadFile(blueprintPath)
		if err != nil {
			logger.Debug(fmt.Sprintf("Failed to read blueprint file %s: %s", blueprintPath, err))
			continue
		}

		// Parse the change request
		changeRequest, err := models.LoadChangeRequestFromContent(blueprintPath, content)
		if err != nil {
			logger.Debug(fmt.Sprintf("Failed to parse change request from %s: %s", blueprintPath, err))
			continue
		}

		// Check if there is a corresponding implementation file
		implementationFilename := strings.Replace(filename, ".blueprint.md", ".implementation.md", 1)
		implementationPath := filepath.Join(changeRequestsDir, implementationFilename)
		
		if !fs.Exists(implementationPath) {
			// This is an incomplete change request (has blueprint but no implementation)
			incompleteChangeRequests = append(incompleteChangeRequests, changeRequest)
		}
	}

	return incompleteChangeRequests, nil
}

// FormatDescription formats a change request for display in selection list
func FormatDescription(cr models.ChangeRequest) string {
	// Format the creation date
	createdAt := cr.CreatedAt.Format("2006-01-02 15:04:05")
	
	// Count the number of user stories
	userStoryCount := len(cr.UserStories)
	
	// Create a description
	return fmt.Sprintf("%s (Created: %s, Stories: %d)", cr.Name, createdAt, userStoryCount)
} 