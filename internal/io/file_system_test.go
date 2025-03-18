package io

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOSFileSystem(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "usm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	// Cleanup after test
	defer os.RemoveAll(tempDir)

	fs := NewOSFileSystem()

	// Test MkdirAll
	testDir := filepath.Join(tempDir, "test/nested/dir")
	if err := fs.MkdirAll(testDir, 0755); err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Test Exists
	if !fs.Exists(testDir) {
		t.Errorf("Exists returned false for existing directory")
	}
	if fs.Exists(filepath.Join(tempDir, "nonexistent")) {
		t.Errorf("Exists returned true for non-existent path")
	}

	// Test WriteFile
	testFile := filepath.Join(testDir, "test.txt")
	testContent := []byte("test content")
	if err := fs.WriteFile(testFile, testContent, 0644); err != nil {
		t.Errorf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	content, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("ReadFile returned wrong content: got %s, want %s", content, testContent)
	}

	// Test ReadDir
	entries, err := fs.ReadDir(testDir)
	if err != nil {
		t.Errorf("ReadDir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("ReadDir returned wrong number of entries: got %d, want 1", len(entries))
	}
	if entries[0].Name() != "test.txt" {
		t.Errorf("ReadDir returned wrong entry name: got %s, want test.txt", entries[0].Name())
	}

	// Test WalkDir
	foundFiles := 0
	err = fs.WalkDir(tempDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			foundFiles++
		}
		return nil
	})
	if err != nil {
		t.Errorf("WalkDir failed: %v", err)
	}
	if foundFiles != 1 {
		t.Errorf("WalkDir found wrong number of files: got %d, want 1", foundFiles)
	}
} 