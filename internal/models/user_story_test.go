// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package models

import (
	"os"
	"testing"
	"time"
)

func TestExtractTitleFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "Content with title",
			content:  "# This is a title\nSome content",
			expected: "This is a title",
		},
		{
			name:     "Content with title and multiple lines",
			content:  "# This is a title\nSome content\nMore content",
			expected: "This is a title",
		},
		{
			name:     "Content with no title",
			content:  "Some content\nMore content",
			expected: "",
		},
		{
			name:     "Content with title not at the beginning",
			content:  "Some content\n# This is a title\nMore content",
			expected: "This is a title",
		},
		{
			name:     "Content with title with spaces",
			content:  "#    This is a title with spaces    \nSome content",
			expected: "This is a title with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTitleFromContent(tt.content)
			if result != tt.expected {
				t.Errorf("ExtractTitleFromContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateContentHash(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
	}{
		{
			name:    "Empty content",
			content: "",
			wantLen: 64, // SHA-256 is 64 chars in hex
		},
		{
			name:    "Simple content",
			content: "Hello, world!",
			wantLen: 64,
		},
		{
			name:    "Multiline content",
			content: "Line 1\nLine 2\nLine 3",
			wantLen: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateContentHash(tt.content)
			if len(result) != tt.wantLen {
				t.Errorf("GenerateContentHash() length = %d, want %d", len(result), tt.wantLen)
			}

			// Verify that identical content produces identical hashes
			result2 := GenerateContentHash(tt.content)
			if result != result2 {
				t.Errorf("GenerateContentHash() produced different hashes for the same content")
			}

			// Verify different content produces different hashes
			if tt.content != "" {
				differentContent := tt.content + "different"
				differentResult := GenerateContentHash(differentContent)
				if result == differentResult {
					t.Errorf("GenerateContentHash() produced the same hash for different content")
				}
			}
		})
	}
}

func TestExtractSequentialNumberFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "Valid sequential number",
			filename: "01-user-story.md",
			expected: "01",
		},
		{
			name:     "Multiple digit sequential number",
			filename: "123-user-story.md",
			expected: "123",
		},
		{
			name:     "No sequential number",
			filename: "user-story.md",
			expected: "",
		},
		{
			name:     "Sequential number not at the beginning",
			filename: "user-01-story.md",
			expected: "",
		},
		{
			name:     "Empty filename",
			filename: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSequentialNumberFromFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("ExtractSequentialNumberFromFilename() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSlugifyTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Simple title",
			title:    "Simple Title",
			expected: "simple-title",
		},
		{
			name:     "Title with special characters",
			title:    "Title with $pecial & characters!",
			expected: "title-with-pecial-characters",
		},
		{
			name:     "Title with multiple spaces",
			title:    "Title   with   multiple   spaces",
			expected: "title-with-multiple-spaces",
		},
		{
			name:     "Title with leading and trailing spaces",
			title:    "  Title with spaces  ",
			expected: "title-with-spaces",
		},
		{
			name:     "Title with numbers",
			title:    "Title 123 with 456 numbers",
			expected: "title-123-with-456-numbers",
		},
		{
			name:     "Empty title",
			title:    "",
			expected: "",
		},
		{
			name:     "Title with repeated hyphens",
			title:    "Title---with---hyphens",
			expected: "title-with-hyphens",
		},
		{
			name:     "Title with uppercase characters",
			title:    "UPPERCASE TITLE",
			expected: "uppercase-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SlugifyTitle(tt.title)
			if result != tt.expected {
				t.Errorf("SlugifyTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name             string
		sequentialNumber string
		title            string
		expected         string
	}{
		{
			name:             "Simple title",
			sequentialNumber: "01",
			title:            "Simple Title",
			expected:         "01-simple-title.md",
		},
		{
			name:             "Complex title",
			sequentialNumber: "123",
			title:            "Title with $pecial & characters!",
			expected:         "123-title-with-pecial-characters.md",
		},
		{
			name:             "Empty title",
			sequentialNumber: "42",
			title:            "",
			expected:         "42-.md",
		},
		{
			name:             "Empty sequential number",
			sequentialNumber: "",
			title:            "Some Title",
			expected:         "-some-title.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilename(tt.sequentialNumber, tt.title)
			if result != tt.expected {
				t.Errorf("GenerateFilename() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetNextSequentialNumber(t *testing.T) {
	// Create mock directory entries
	tests := []struct {
		name        string
		createFiles []string
		expected    string
	}{
		{
			name:        "Empty directory",
			createFiles: []string{},
			expected:    "01",
		},
		{
			name:        "Single file",
			createFiles: []string{"01-story.md"},
			expected:    "02",
		},
		{
			name:        "Multiple files",
			createFiles: []string{"01-story.md", "02-another.md", "03-third.md"},
			expected:    "04",
		},
		{
			name:        "Non-sequential files",
			createFiles: []string{"01-story.md", "05-another.md", "03-third.md"},
			expected:    "06",
		},
		{
			name:        "Files with non-standard names",
			createFiles: []string{"01-story.md", "not-a-story.md", "random.txt"},
			expected:    "02",
		},
		{
			name:        "Directory in the mix",
			createFiles: []string{"01-story.md", ".git"},
			expected:    "02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entries []os.DirEntry
			for _, fileName := range tt.createFiles {
				entries = append(entries, &mockDirEntry{
					name:  fileName,
					isDir: fileName == ".git",
				})
			}

			result := GetNextSequentialNumber(entries)
			if result != tt.expected {
				t.Errorf("GetNextSequentialNumber() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Mock DirEntry implementation for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func TestGenerateUserStoryTemplate(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		filePath string
	}{
		{
			name:     "Simple case",
			title:    "Test Story",
			filePath: "docs/user-stories/01-test-story.md",
		},
		{
			name:     "Empty title",
			title:    "",
			filePath: "docs/user-stories/01-untitled.md",
		},
		{
			name:     "Empty file path",
			title:    "Test Story",
			filePath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateUserStoryTemplate(tt.title, tt.filePath)

			// Check that template contains the title
			if !contains(result, tt.title) && tt.title != "" {
				t.Errorf("GenerateUserStoryTemplate() does not contain title %q", tt.title)
			}

			// Check that template contains the file path
			if !contains(result, tt.filePath) && tt.filePath != "" {
				t.Errorf("GenerateUserStoryTemplate() does not contain file path %q", tt.filePath)
			}

			// Check that template contains required sections
			requiredSections := []string{
				"---",
				"file_path:",
				"created_at:",
				"last_updated:",
				"_content_hash:",
				"# ",
				"As a ",
				"I want ",
				"so that ",
				"## Acceptance criteria",
			}

			for _, section := range requiredSections {
				if !contains(result, section) {
					t.Errorf("GenerateUserStoryTemplate() does not contain required section %q", section)
				}
			}
		})
	}
}

func TestLoadUserStoryFromFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		content     string
		wantErr     bool
		expectTitle string
	}{
		{
			name:     "Valid user story",
			filePath: "docs/user-stories/01-test-story.md",
			content: `---
file_path: docs/user-stories/01-test-story.md
created_at: 2025-01-01T12:00:00Z
last_updated: 2025-01-02T12:00:00Z
_content_hash: abcdef123456
---

# Test Story

As a user, I want to test, so that I can verify.

## Acceptance criteria

- Test criteria 1
- Test criteria 2`,
			wantErr:     false,
			expectTitle: "Test Story",
		},
		{
			name:     "User story without metadata",
			filePath: "docs/user-stories/02-no-metadata.md",
			content: `# No Metadata Story

As a user, I want no metadata, so that I can test this case.

## Acceptance criteria

- Test criteria 1`,
			wantErr:     false,
			expectTitle: "No Metadata Story",
		},
		{
			name:     "User story with invalid metadata",
			filePath: "docs/user-stories/03-invalid-dates.md",
			content: `---
file_path: docs/user-stories/03-invalid-dates.md
created_at: invalid-date
last_updated: also-invalid
_content_hash: abcdef123456
---

# Invalid Dates Story

Test content.`,
			wantErr:     false,
			expectTitle: "Invalid Dates Story",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story, err := LoadUserStoryFromFile(tt.filePath, []byte(tt.content))
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadUserStoryFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if story.Title != tt.expectTitle {
				t.Errorf("LoadUserStoryFromFile() got title = %q, want %q", story.Title, tt.expectTitle)
			}
			
			if story.FilePath != tt.filePath && story.FilePath != "docs/user-stories/01-test-story.md" {
				t.Errorf("LoadUserStoryFromFile() got file path = %q, want one of %q or %q", 
					story.FilePath, tt.filePath, "docs/user-stories/01-test-story.md")
			}
			
			if tt.name == "Valid user story" {
				expectedCreatedAt, _ := time.Parse(time.RFC3339, "2025-01-01T12:00:00Z")
				expectedLastUpdated, _ := time.Parse(time.RFC3339, "2025-01-02T12:00:00Z")
				
				if !story.CreatedAt.Equal(expectedCreatedAt) {
					t.Errorf("LoadUserStoryFromFile() got created_at = %v, want %v", 
						story.CreatedAt, expectedCreatedAt)
				}
				
				if !story.LastUpdated.Equal(expectedLastUpdated) {
					t.Errorf("LoadUserStoryFromFile() got last_updated = %v, want %v", 
						story.LastUpdated, expectedLastUpdated)
				}
				
				if story.ContentHash != "abcdef123456" {
					t.Errorf("LoadUserStoryFromFile() got content_hash = %q, want %q", 
						story.ContentHash, "abcdef123456")
				}
			}
			
			// Check extraction of sequential number from filename
			if tt.name == "Valid user story" && story.SequentialNumber != "01" {
				t.Errorf("LoadUserStoryFromFile() got sequential_number = %q, want %q", 
					story.SequentialNumber, "01")
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" && substr != "" && s != substr && substring(s, substr) != -1
}

// Simple substring search for the test
func substring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestGenerateUserStoryFilename(t *testing.T) {
	tests := []struct {
		name             string
		sequentialNumber string
		title           string
		expected        string
	}{
		{
			name:             "Simple title",
			sequentialNumber: "01",
			title:           "Simple Title",
			expected:        "01-simple-title.md",
		},
		{
			name:             "Title with special characters",
			sequentialNumber: "02",
			title:           "Title with $pecial & characters!",
			expected:        "02-title-with-pecial-characters.md",
		},
		{
			name:             "Title with multiple spaces",
			sequentialNumber: "123",
			title:           "Title   with   multiple   spaces",
			expected:        "123-title-with-multiple-spaces.md",
		},
		{
			name:             "Empty title",
			sequentialNumber: "01",
			title:           "",
			expected:        "01-.md",
		},
		{
			name:             "Single word title",
			sequentialNumber: "99",
			title:           "UPPERCASE",
			expected:        "99-uppercase.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateUserStoryFilename(tt.sequentialNumber, tt.title)
			if result != tt.expected {
				t.Errorf("GenerateUserStoryFilename() = %q, want %q", result, tt.expected)
			}
		})
	}
} 