// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package models

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// UserStory represents a user story document
type UserStory struct {
	Title            string    `json:"title"`
	FilePath         string    `json:"file_path"`
	ContentHash      string    `json:"content_hash"`
	SequentialNumber string    `json:"sequential_number"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdated      time.Time `json:"last_updated"`
	Content          string    `json:"content"`
	Description      string    `json:"description"`
	Criteria         []string  `json:"criteria"`
	IsImplemented    bool      `json:"is_implemented"`
	MatchScore      float64   `json:"match_score"`
}

// ExtractTitleFromContent extracts the title from the markdown content
func ExtractTitleFromContent(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// ExtractMetadataFromContent extracts the metadata from the markdown content
func ExtractMetadataFromContent(content string) (map[string]string, error) {
	metadata := make(map[string]string)
	
	// Looking for metadata section at the beginning of the file
	// Format:
	// ---
	// key: value
	// ---
	
	metadataRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	matches := metadataRegex.FindStringSubmatch(content)
	
	if len(matches) < 2 {
		return metadata, nil
	}
	
	metadataContent := matches[1]
	lines := strings.Split(metadataContent, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		metadata[key] = value
	}
	
	return metadata, nil
}

// GenerateContentHash calculates the MD5 hash of the content
func GenerateContentHash(content string) string {
	hash := md5.New()
	io.WriteString(hash, content)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// ExtractSequentialNumberFromFilename extracts the sequential number from a filename
func ExtractSequentialNumberFromFilename(filename string) string {
	re := regexp.MustCompile(`^(\d+)-`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// SlugifyTitle converts a title to a slug for use in filenames
func SlugifyTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)
	
	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	
	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")
	
	return slug
}

// GenerateFilename generates a filename for a user story
func GenerateFilename(sequentialNumber, title string) string {
	slug := SlugifyTitle(title)
	return fmt.Sprintf("%s-%s.md", sequentialNumber, slug)
}

// GetNextSequentialNumber calculates the next sequential number in a directory
func GetNextSequentialNumber(dirEntries []os.DirEntry) string {
	maxNum := 0
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		
		seqNum := ExtractSequentialNumberFromFilename(entry.Name())
		if seqNum != "" {
			num := 0
			fmt.Sscanf(seqNum, "%d", &num)
			if num > maxNum {
				maxNum = num
			}
		}
	}
	
	return fmt.Sprintf("%02d", maxNum+1)
}

// GenerateUserStoryTemplate generates a template for a new user story
func GenerateUserStoryTemplate(title, filePath string) string {
	template := `---
file_path: {{file_path}}
created_at: {{created_at}}
last_updated: {{last_updated}}
_content_hash: {{content_hash}}
---

# {{title}}

As a <type of user>,  
I want <some goal>,  
so that <some reason>.

## Acceptance criteria

- First criteria
- Second criteria
- Third criteria
`

	// Fill in the file path
	template = strings.ReplaceAll(template, "{{file_path}}", filePath)

	// Fill in the title
	template = strings.ReplaceAll(template, "{{title}}", title)

	// Fill in the dates
	now := time.Now().Format(time.RFC3339)
	template = strings.ReplaceAll(template, "{{created_at}}", now)
	template = strings.ReplaceAll(template, "{{last_updated}}", now)

	// Generate a placeholder content hash
	template = strings.ReplaceAll(template, "{{content_hash}}", "placeholder")

	return template
}

// GenerateUserStoryFilename generates a filename for a user story
func GenerateUserStoryFilename(sequentialNumber, title string) string {
	// Format: <sequential-number>-<slugified-title>.md
	slug := SlugifyTitle(title)
	return fmt.Sprintf("%s-%s.md", sequentialNumber, slug)
}

// LoadUserStoryFromFile loads a user story from file content
func LoadUserStoryFromFile(filePath string, content []byte) (UserStory, error) {
	us := UserStory{
		FilePath: filePath,
	}

	contentStr := string(content)

	// Extract metadata
	metadata, err := ExtractMetadataFromContent(contentStr)
	if err != nil {
		return us, err
	}

	// Get file path
	if filePath, ok := metadata["file_path"]; ok {
		us.FilePath = filePath
	}

	// Get content hash
	if contentHash, ok := metadata["_content_hash"]; ok {
		us.ContentHash = contentHash
	}

	// Parse creation date
	if createdAt, ok := metadata["created_at"]; ok {
		t, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			us.CreatedAt = t
		}
	}

	// Parse last updated date
	if lastUpdated, ok := metadata["last_updated"]; ok {
		t, err := time.Parse(time.RFC3339, lastUpdated)
		if err == nil {
			us.LastUpdated = t
		}
	}

	// Extract sequential number from filename
	base := filepath.Base(filePath)
	seqRegex := regexp.MustCompile(`^(\d+)-`)
	if match := seqRegex.FindStringSubmatch(base); len(match) > 1 {
		us.SequentialNumber = match[1]
	}

	// Extract title from content
	titleRegex := regexp.MustCompile(`(?m)^# (.+)$`)
	if match := titleRegex.FindStringSubmatch(contentStr); len(match) > 1 {
		us.Title = match[1]
	}

	// Store full content
	us.Content = contentStr

	// Extract description (everything between title and first ## heading)
	descRegex := regexp.MustCompile(`(?ms)^# .+\n\n(.*?)\n\n##`)
	if match := descRegex.FindStringSubmatch(contentStr); len(match) > 1 {
		us.Description = strings.TrimSpace(match[1])
	}

	// Extract acceptance criteria
	criteriaRegex := regexp.MustCompile(`(?m)^- (.+)$`)
	matches := criteriaRegex.FindAllStringSubmatch(contentStr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			us.Criteria = append(us.Criteria, match[1])
		}
	}

	return us, nil
}