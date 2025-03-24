// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package implementation

import (
	"path/filepath"
	"strings"

	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"github.com/user-story-matrix/usm/internal/models"
)

// IsUserStoryImplemented checks if a user story is referenced by any implemented change request
func IsUserStoryImplemented(userStory models.UserStory, fs io.FileSystem) (bool, error) {
	// Define the change requests directory
	changeRequestsDir := "docs/changes-request"

	// Check if the directory exists
	if !fs.Exists(changeRequestsDir) {
		return false, nil // No change requests directory means no implementations
	}

	// Get all files in the directory
	entries, err := fs.ReadDir(changeRequestsDir)
	if err != nil {
		return false, err
	}

	// First find all blueprint files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".blueprint.md") {
			continue
		}

		// Check if there is a corresponding implementation file
		implementationFilename := strings.Replace(filename, ".blueprint.md", ".implementation.md", 1)
		implementationPath := filepath.Join(changeRequestsDir, implementationFilename)
		
		if fs.Exists(implementationPath) {
			// This is an implemented change request, check if it references our user story
			blueprintPath := filepath.Join(changeRequestsDir, filename)
			content, err := fs.ReadFile(blueprintPath)
			if err != nil {
				logger.Debug("Failed to read blueprint file: " + err.Error())
				continue
			}

			// Parse the change request
			changeRequest, err := models.LoadChangeRequestFromContent(blueprintPath, content)
			if err != nil {
				logger.Debug("Failed to parse change request: " + err.Error())
				continue
			}

			// Check if the user story is referenced
			for _, reference := range changeRequest.UserStories {
				if reference.FilePath == userStory.FilePath {
					return true, nil
				}
			}
		}
	}

	// If we get here, the user story is not referenced in any implemented change request
	return false, nil
}

// UpdateImplementationStatus updates the IsImplemented flag on a user story
func UpdateImplementationStatus(userStory *models.UserStory, fs io.FileSystem) error {
	implemented, err := IsUserStoryImplemented(*userStory, fs)
	if err != nil {
		return err
	}
	
	userStory.IsImplemented = implemented
	return nil
} 