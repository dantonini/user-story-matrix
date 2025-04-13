// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/logger"
)

// TestUpdateUserStoriesWorkflow is an integration test that verifies the complete workflow
// from updating user story metadata to updating references in blueprint files.
// This test simulates the real-world scenario where a user runs the update command
// and it both adds metadata to user stories and updates references in blueprints.
func TestUpdateUserStoriesWorkflow(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir := t.TempDir()
	
	// Create the directory structure
	docsDir := filepath.Join(tempDir, "docs")
	userStoriesDir := filepath.Join(docsDir, "user-stories")
	changesRequestDir := filepath.Join(docsDir, "changes-request")
	
	// Create the required directories
	err := os.MkdirAll(userStoriesDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(changesRequestDir, 0755)
	assert.NoError(t, err)
	
	// Create a user story file WITHOUT metadata
	userStoryContent := `# Test User Story
This is a test user story.
`
	userStoryPath := filepath.Join(userStoriesDir, "test-story.md")
	err = os.WriteFile(userStoryPath, []byte(userStoryContent), 0600)
	assert.NoError(t, err)
	
	// Create a blueprint file that references the user story with an empty content-hash
	blueprintContent := `---
name: test-blueprint
created-at: 2025-04-01T09:00:00+02:00
user-stories:
  - title: Test User Story
    file: docs/user-stories/test-story.md
    content-hash: 
---
# Test Blueprint
`
	blueprintPath := filepath.Join(changesRequestDir, "test.blueprint.md")
	err = os.WriteFile(blueprintPath, []byte(blueprintContent), 0600)
	assert.NoError(t, err)
	
	// Save the original stdout
	origStdout := os.Stdout
	
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Ensure we restore stdout when done
	defer func() {
		os.Stdout = origStdout
		logger.SetDebugMode(false)
	}()
	
	// Create a command for testing
	cmd := &cobra.Command{}
	cmd.Flags().Bool("skip-references", false, "")
	cmd.Flags().Bool("debug", true, "")
	cmd.Flags().String("test-root", tempDir, "")
	
	// Run the update command function
	err = updateUserStoriesCmd.RunE(cmd, []string{})
	assert.NoError(t, err)
	
	// Close the write end of the pipe
	w.Close()
	
	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	
	// Print the output for debugging
	t.Logf("Command output:\n%s", output)
	
	// Verify the output contains expected messages
	assert.Contains(t, output, "Updated user story metadata")
	assert.Contains(t, output, "test-story.md")
	assert.Contains(t, output, "Updating references in change requests")
	
	// Read the updated blueprint file to verify content-hash was updated
	updatedBlueprint, err := os.ReadFile(blueprintPath)
	assert.NoError(t, err)
	
	// The updated blueprint should no longer have an empty content-hash
	assert.NotContains(t, string(updatedBlueprint), "content-hash: \n")
	
	// It should have the hash with a single space after the colon
	// We don't know the exact hash value, but we can check for the format
	assert.Contains(t, string(updatedBlueprint), "content-hash: ")
	
	// Make sure there are no double spaces after the colon
	assert.NotContains(t, string(updatedBlueprint), "content-hash:  ")
} 