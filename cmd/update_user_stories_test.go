package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// TestExtractExistingMetadata tests the extraction of metadata from content
func TestExtractExistingMetadata(t *testing.T) {
	content := `---
file_path: docs/user-stories/example/sample.md
created_at: 2023-01-01T12:00:00Z
last_updated: 2023-01-02T12:00:00Z
_content_hash: abcdef1234567890
---

# Sample User Story

This is a sample user story.
`
	
	metadata := extractExistingMetadata(content)
	
	expected := map[string]string{
		"file_path":     "docs/user-stories/example/sample.md",
		"created_at":    "2023-01-01T12:00:00Z",
		"last_updated":  "2023-01-02T12:00:00Z",
		"_content_hash": "abcdef1234567890",
	}
	
	// Check if all expected keys exist with the right values
	for key, expectedValue := range expected {
		value, exists := metadata[key]
		if !exists {
			t.Errorf("Expected metadata to contain key %q but it doesn't", key)
		} else if value != expectedValue {
			t.Errorf("Expected metadata[%q] to be %q but got %q", key, expectedValue, value)
		}
	}
	
	// Check if there are no unexpected keys
	for key := range metadata {
		if _, exists := expected[key]; !exists {
			t.Errorf("Unexpected key %q in metadata", key)
		}
	}
}

// TestGetContentWithoutMetadata tests the removal of metadata from content
func TestGetContentWithoutMetadata(t *testing.T) {
	content := `---
file_path: docs/user-stories/example/sample.md
created_at: 2023-01-01T12:00:00Z
last_updated: 2023-01-02T12:00:00Z
_content_hash: abcdef1234567890
---

# Sample User Story

This is a sample user story.
`
	
	expected := `# Sample User Story

This is a sample user story.
`
	
	result := getContentWithoutMetadata(content)
	
	if result != expected {
		t.Errorf("Expected content without metadata to be:\n%q\nbut got:\n%q", expected, result)
	}
}

// TestCalculateContentHash tests the hash calculation of content
func TestCalculateContentHash(t *testing.T) {
	content := "# Sample User Story\n\nThis is a sample user story.\n"
	
	hash := calculateContentHash(content)
	
	// The expected hash is the MD5 hash of the content
	expectedHash := "00db2e256db8faa52e9c56dad9e0b9bd"
	
	if hash != expectedHash {
		t.Errorf("Expected hash to be %q but got %q", expectedHash, hash)
	}
}

// TestGenerateMetadata tests the generation of metadata
func TestGenerateMetadata(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "user-story-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a sample user story file
	filePath := filepath.Join(tempDir, "sample.md")
	err = os.WriteFile(filePath, []byte("# Sample User Story\n\nThis is a sample user story.\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sample file: %v", err)
	}
	
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	
	// Set up existing metadata
	existingMetadata := map[string]string{
		"created_at": "2023-01-01T12:00:00Z",
	}
	
	// Set the content hash
	contentHash := "00db2e256db8faa52e9c56dad9e0b9bd"
	
	// Generate metadata
	metadata := generateMetadata(filePath, tempDir, fileInfo, existingMetadata, contentHash)
	
	// Check if the metadata contains the expected fields
	expectedFields := []string{
		"file_path:", 
		"created_at: 2023-01-01T12:00:00Z", 
		"last_updated:", 
		"_content_hash: 00db2e256db8faa52e9c56dad9e0b9bd",
	}
	
	for _, field := range expectedFields {
		if !strings.Contains(metadata, field) {
			t.Errorf("Expected metadata to contain %q but it doesn't: %s", field, metadata)
		}
	}
	
	// Check if the metadata format is correct
	if !strings.HasPrefix(metadata, "---\n") || !strings.Contains(metadata, "\n---\n\n") {
		t.Errorf("Metadata format is incorrect: %s", metadata)
	}
}

// TestUpdateFileMetadata tests the update of file metadata
func TestUpdateFileMetadata(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "user-story-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test cases
	testCases := []struct {
		name           string
		initialContent string
		expectedUpdate bool
	}{
		{
			name:           "New file without metadata",
			initialContent: "# Sample User Story\n\nThis is a sample user story.\n",
			expectedUpdate: true,
		},
		{
			name: "File with outdated metadata",
			initialContent: `---
file_path: sample.md
created_at: 2023-01-01T12:00:00Z
last_updated: 2023-01-02T12:00:00Z
_content_hash: outdated-hash
---

# Sample User Story

This is a sample user story.
`,
			expectedUpdate: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the sample file
			filePath := filepath.Join(tempDir, "sample.md")
			err = os.WriteFile(filePath, []byte(tc.initialContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create sample file: %v", err)
			}
			
			// Update the metadata
			updated, hash, err := updateFileMetadata(filePath, tempDir)
			if err != nil {
				t.Fatalf("updateFileMetadata failed: %v", err)
			}
			
			// Check if the file was updated as expected
			if updated != tc.expectedUpdate {
				t.Errorf("Expected update to be %v but got %v", tc.expectedUpdate, updated)
			}
			
			// Read the updated content
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read updated file: %v", err)
			}
			
			// Check if the content contains metadata
			if !strings.Contains(string(content), "---\n") || !strings.Contains(string(content), "\n---\n\n") {
				t.Errorf("Updated content doesn't contain metadata: %s", string(content))
			}
			
			// Check if the content contains the right content hash
			if !strings.Contains(string(content), "_content_hash: "+hash) {
				t.Errorf("Updated content doesn't contain correct hash: %s", string(content))
			}
			
			// Run update again to ensure no changes are made the second time
			updated, _, err = updateFileMetadata(filePath, tempDir)
			if err != nil {
				t.Fatalf("Second updateFileMetadata failed: %v", err)
			}
			
			// Check that no updates were made the second time
			if updated {
				t.Error("Expected no updates on second run but file was updated")
			}
		})
	}
}

// TestFindMarkdownFiles tests finding markdown files in a directory
func TestFindMarkdownFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "user-story-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a directory structure
	userStoriesDir := filepath.Join(tempDir, "user-stories")
	subDir1 := filepath.Join(userStoriesDir, "subdir1")
	subDir2 := filepath.Join(userStoriesDir, "subdir2")
	ignoredDir := filepath.Join(userStoriesDir, "node_modules")
	
	for _, dir := range []string{userStoriesDir, subDir1, subDir2, ignoredDir} {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	
	// Create some markdown files
	files := map[string]string{
		filepath.Join(userStoriesDir, "root.md"):         "# Root File",
		filepath.Join(subDir1, "subdir1.md"):             "# Subdir1 File",
		filepath.Join(subDir2, "subdir2.md"):             "# Subdir2 File",
		filepath.Join(ignoredDir, "ignored.md"):          "# Ignored File",
		filepath.Join(userStoriesDir, "not-markdown.txt"): "Not a markdown file",
	}
	
	for path, content := range files {
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}
	
	// Find markdown files
	markdownFiles, err := findMarkdownFiles(userStoriesDir)
	if err != nil {
		t.Fatalf("findMarkdownFiles failed: %v", err)
	}
	
	// Expected files to find (ignoring files in node_modules)
	expectedFiles := []string{
		filepath.Join(userStoriesDir, "root.md"),
		filepath.Join(subDir1, "subdir1.md"),
		filepath.Join(subDir2, "subdir2.md"),
	}
	
	// Check if all expected files were found
	if len(markdownFiles) != len(expectedFiles) {
		t.Errorf("Expected to find %d files but found %d", len(expectedFiles), len(markdownFiles))
	}
	
	// Create a map of found files for easier checking
	foundFiles := make(map[string]bool)
	for _, file := range markdownFiles {
		foundFiles[file] = true
	}
	
	// Check if each expected file was found
	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected to find file %s but it wasn't found", expected)
		}
	}
	
	// Check that ignored files weren't found
	if foundFiles[filepath.Join(ignoredDir, "ignored.md")] {
		t.Error("Found a file in an ignored directory")
	}
	
	// Check that non-markdown files weren't found
	if foundFiles[filepath.Join(userStoriesDir, "not-markdown.txt")] {
		t.Error("Found a non-markdown file")
	}
}

// TestEndToEndUpdateProcess tests the entire update process
func TestEndToEndUpdateProcess(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "user-story-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a user stories directory structure
	userStoriesDir := filepath.Join(tempDir, "docs", "user-stories")
	err = os.MkdirAll(userStoriesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user stories directory: %v", err)
	}
	
	// Create a sample user story file
	sampleFile := filepath.Join(userStoriesDir, "sample.md")
	initialContent := "# Sample User Story\n\nThis is a sample user story.\n"
	err = os.WriteFile(sampleFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create sample file: %v", err)
	}
	
	// Update the metadata (simulating what the command would do)
	updated, hash1, err := updateFileMetadata(sampleFile, tempDir)
	if err != nil {
		t.Fatalf("updateFileMetadata failed: %v", err)
	}
	
	if !updated {
		t.Error("Expected file to be updated but it wasn't")
	}
	
	// Read the updated content
	content, err := os.ReadFile(sampleFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	
	// Check that metadata was added
	if !strings.Contains(string(content), "---\n") || !strings.Contains(string(content), "\n---\n\n") {
		t.Errorf("Updated content doesn't contain metadata: %s", string(content))
	}
	
	// Run update again to ensure no changes are made the second time
	updated, hash2, err := updateFileMetadata(sampleFile, tempDir)
	if err != nil {
		t.Fatalf("Second updateFileMetadata failed: %v", err)
	}
	
	if updated {
		t.Error("Expected no updates on second run but file was updated")
	}
	
	if hash1 != hash2 {
		t.Errorf("Expected hash to remain the same but got %s then %s", hash1, hash2)
	}
	
	// Modify the file and verify metadata is updated
	modifiedContent := "# Sample User Story\n\nThis is a modified sample user story.\n"
	
	// Extract the metadata from the current file
	currentContent, err := os.ReadFile(sampleFile)
	if err != nil {
		t.Fatalf("Failed to read file for modification: %v", err)
	}
	
	metadataSection := metadataRegex.FindString(string(currentContent))
	
	// Write the file with existing metadata and modified content
	err = os.WriteFile(sampleFile, []byte(metadataSection+modifiedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write modified file: %v", err)
	}
	
	// Update the metadata again
	updated, hash3, err := updateFileMetadata(sampleFile, tempDir)
	if err != nil {
		t.Fatalf("Third updateFileMetadata failed: %v", err)
	}
	
	if !updated {
		t.Error("Expected file to be updated after content change but it wasn't")
	}
	
	if hash2 == hash3 {
		t.Errorf("Expected hash to change after content modification but got the same hash: %s", hash2)
	}
	
	// Check that last_updated field was updated
	content, err = os.ReadFile(sampleFile)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}
	
	metadata := extractExistingMetadata(string(content))
	
	// The last_updated time should be recent (within the last minute)
	lastUpdated := metadata["last_updated"]
	lastUpdatedTime, err := time.Parse(time.RFC3339, lastUpdated)
	if err != nil {
		t.Fatalf("Failed to parse last_updated time: %v", err)
	}
	
	if time.Since(lastUpdatedTime) > time.Minute {
		t.Errorf("Expected last_updated to be recent but got %s", lastUpdated)
	}
}

// TestUpdateUserStoriesCommand tests the update command execution
func TestUpdateUserStoriesCommand(t *testing.T) {
	// Skip this test on CI environments or when running all tests
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	
	// Save the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "usm-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Make sure to change back when we're done
	defer os.Chdir(currentDir)
	
	// Create docs/user-stories directory
	userStoriesDir := filepath.Join(tempDir, "docs", "user-stories")
	err = os.MkdirAll(userStoriesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user stories directory: %v", err)
	}
	
	// Create a test user story file
	testFile := filepath.Join(userStoriesDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test User Story\n\nThis is a test user story.\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// First run - capture output
	t.Log("First run - adding metadata")
	var outputBuffer strings.Builder
	origStdout := os.Stdout
	
	// Create a pipe for capturing stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Execute the command
	updateUserStoriesCmd.Run(nil, nil)
	
	// Close the write end of the pipe to read all output
	w.Close()
	
	// Read all output
	_, err = io.Copy(&outputBuffer, r)
	if err != nil {
		t.Fatalf("Failed to capture command output: %v", err)
	}
	
	// Restore stdout
	os.Stdout = origStdout
	
	// Check if the output contains expected text
	output := outputBuffer.String()
	t.Logf("First run output: %s", output)
	
	expectedOutputs := []string{
		"Updated metadata for:", 
		"Processed 1 files (1 updated, 0 unchanged)",
	}
	
	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q but it doesn't: %s", expected, output)
		}
	}
	
	// Check if metadata was actually added to the file
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	// Verify the file has metadata
	if !strings.Contains(string(content), "---\n") || !strings.Contains(string(content), "\n---\n\n") {
		t.Errorf("Expected metadata to be added but it wasn't: %s", string(content))
	}
	
	// Verify the file has metadata fields
	for _, field := range []string{"file_path:", "created_at:", "last_updated:", "_content_hash:"} {
		if !strings.Contains(string(content), field) {
			t.Errorf("Expected metadata to contain %q but it doesn't: %s", field, string(content))
		}
	}
	
	// Second run - capture output
	t.Log("Second run - checking for no changes")
	outputBuffer.Reset()
	
	// Create another pipe
	r, w, _ = os.Pipe()
	os.Stdout = w
	
	// Execute the command again
	updateUserStoriesCmd.Run(nil, nil)
	
	// Close the write end of the pipe
	w.Close()
	
	// Read all output
	_, err = io.Copy(&outputBuffer, r)
	if err != nil {
		t.Fatalf("Failed to capture command output: %v", err)
	}
	
	// Restore stdout
	os.Stdout = origStdout
	
	// Check the second run output
	output = outputBuffer.String()
	t.Logf("Second run output: %s", output)
	
	// Check if the output contains expected text for second run
	expectedSecondRunOutputs := []string{
		"No changes needed", 
		"Processed 1 files (0 updated, 1 unchanged)",
	}
	
	for _, expected := range expectedSecondRunOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected second run output to contain %q but it doesn't: %s", expected, output)
		}
	}
}

// TestUpdateUserStoriesCommandWithDebug tests the update command with debug flag
func TestUpdateUserStoriesCommandWithDebug(t *testing.T) {
	// Skip this test on CI environments or when running all tests
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	
	// Save the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "usm-cmd-debug-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Make sure to change back when we're done
	defer os.Chdir(currentDir)
	
	// Create docs/user-stories directory
	userStoriesDir := filepath.Join(tempDir, "docs", "user-stories")
	err = os.MkdirAll(userStoriesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user stories directory: %v", err)
	}
	
	// Create a test user story file
	testFile := filepath.Join(userStoriesDir, "debug-test.md")
	err = os.WriteFile(testFile, []byte("# Debug Test User Story\n\nThis is a debug test user story.\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Since we can't reliably capture debug output in tests (because it's sent to 
	// a logger that may go to stderr or to a file), we'll just run the test and
	// verify that it completes successfully and updates the file as expected
	
	// Create a cobra command with debug flag
	cmd := &cobra.Command{}
	cmd.Flags().Bool("debug", true, "")
	cmd.Flag("debug").Value.Set("true")
	
	// Execute the command with debug flag
	updateUserStoriesCmd.Run(cmd, nil)
	
	// Verify the file has metadata
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	if !strings.Contains(string(content), "---\n") || !strings.Contains(string(content), "\n---\n\n") {
		t.Errorf("Expected metadata to be added but it wasn't: %s", string(content))
	}
	
	// Check if metadata contains expected fields
	for _, field := range []string{"file_path:", "created_at:", "last_updated:", "_content_hash:"} {
		if !strings.Contains(string(content), field) {
			t.Errorf("Expected metadata to contain %q but it doesn't: %s", field, string(content))
		}
	}
	
	// Verify that running it a second time doesn't change the file
	// Get the last modified time
	fileInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	modTime1 := fileInfo.ModTime()
	
	// Run the command again
	updateUserStoriesCmd.Run(cmd, nil)
	
	// Get the new last modified time
	fileInfo, err = os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to get updated file info: %v", err)
	}
	modTime2 := fileInfo.ModTime()
	
	// The file should not have been modified
	if !modTime1.Equal(modTime2) {
		t.Errorf("Expected file not to be modified on second run, but it was")
	}
} 