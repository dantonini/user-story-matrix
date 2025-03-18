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
	Title           string    `json:"title"`
	FilePath        string    `json:"file_path"`
	ContentHash     string    `json:"content_hash"`
	SequentialNumber string   `json:"sequential_number"`
	CreatedAt       time.Time `json:"created_at"`
	LastUpdated     time.Time `json:"last_updated"`
	Content         string    `json:"content"`
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
func GenerateUserStoryTemplate(title string) string {
	template := `---
file_path: {{file_path}}
created_at: {{created_at}}
last_updated: {{last_updated}}
_content_hash: {{content_hash}}
---

# {{title}}
Add here a description of the user story

As a ... 
I want ... 
so that ...

## Acceptance criteria
- Add here your acceptance criteria
- As many as needed
`
	
	// Fill in only the title for now, other fields will be filled later
	template = strings.ReplaceAll(template, "{{title}}", title)
	
	return template
}

// FinalizeUserStoryTemplate finalizes the template by filling in the metadata
func FinalizeUserStoryTemplate(template, filePath string) string {
	now := time.Now().Format(time.RFC3339)
	
	template = strings.ReplaceAll(template, "{{file_path}}", filePath)
	template = strings.ReplaceAll(template, "{{created_at}}", now)
	template = strings.ReplaceAll(template, "{{last_updated}}", now)
	
	// Calculate content hash after removing the placeholder
	templateWithoutHash := strings.ReplaceAll(template, "_content_hash: {{content_hash}}", "_content_hash: ")
	contentHash := GenerateContentHash(templateWithoutHash)
	
	template = strings.ReplaceAll(template, "{{content_hash}}", contentHash)
	
	return template
}

// LoadUserStoryFromFile loads a user story from a file
func LoadUserStoryFromFile(filePath string, content []byte) (UserStory, error) {
	us := UserStory{}
	
	// Extract basic information from the file path
	us.FilePath = filePath
	fileName := filepath.Base(filePath)
	us.SequentialNumber = ExtractSequentialNumberFromFilename(fileName)
	
	// Extract information from content
	contentStr := string(content)
	us.Content = contentStr
	us.Title = ExtractTitleFromContent(contentStr)
	
	// Extract metadata
	metadata, err := ExtractMetadataFromContent(contentStr)
	if err != nil {
		return us, err
	}
	
	// Parse timestamps
	if createdAt, ok := metadata["created_at"]; ok {
		t, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			us.CreatedAt = t
		}
	}
	
	if lastUpdated, ok := metadata["last_updated"]; ok {
		t, err := time.Parse(time.RFC3339, lastUpdated)
		if err == nil {
			us.LastUpdated = t
		}
	}
	
	// Get content hash
	if contentHash, ok := metadata["_content_hash"]; ok {
		us.ContentHash = contentHash
	} else {
		us.ContentHash = GenerateContentHash(contentStr)
	}
	
	return us, nil
} 