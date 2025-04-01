// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


package io

import (
	"errors"
	"os"
	"testing"
)

func TestMockFileSystem(t *testing.T) {
	fs := NewMockFileSystem()

	// Test MkdirAll
	testDir := "test/nested/dir"
	if err := fs.MkdirAll(testDir, 0755); err != nil {
		t.Errorf("MkdirAll failed: %v", err)
	}

	// Test Exists
	if !fs.Exists(testDir) {
		t.Errorf("Exists returned false for existing directory")
	}
	if !fs.Exists("test/nested") {
		t.Errorf("Exists returned false for parent directory")
	}
	if fs.Exists("nonexistent") {
		t.Errorf("Exists returned true for non-existent path")
	}

	// Test WriteFile
	testFile := "test/nested/dir/test.txt"
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

	// Test ReadDir for root directory
	entries, err = fs.ReadDir("test")
	if err != nil {
		t.Errorf("ReadDir failed for root: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("ReadDir returned wrong number of entries for root: got %d, want 1", len(entries))
	}
	if entries[0].Name() != "nested" {
		t.Errorf("ReadDir returned wrong entry name for root: got %s, want nested", entries[0].Name())
	}

	// Test ReadDir for non-existent directory
	_, err = fs.ReadDir("nonexistent")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadDir returned wrong error for non-existent directory: got %v, want %v", err, os.ErrNotExist)
	}

	// Test ReadDir for file
	_, err = fs.ReadDir(testFile)
	if !errors.Is(err, os.ErrInvalid) {
		t.Errorf("ReadDir returned wrong error for file: got %v, want %v", err, os.ErrInvalid)
	}

	// Test WalkDir
	foundFiles := 0
	err = fs.WalkDir("test", func(path string, d os.DirEntry, err error) error {
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