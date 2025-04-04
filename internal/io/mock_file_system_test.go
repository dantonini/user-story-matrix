// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package io

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockFileSystem(t *testing.T) {
	fs := NewMockFileSystem()

	// Test MkdirAll
	testDir := "test/nested/dir"
	err := fs.MkdirAll(testDir, 0755)
	assert.NoError(t, err, "MkdirAll failed")

	// Test Exists
	assert.True(t, fs.Exists(testDir), "Exists returned false for existing directory")
	assert.True(t, fs.Exists("test/nested"), "Exists returned false for parent directory")
	assert.False(t, fs.Exists("nonexistent"), "Exists returned true for non-existent path")

	// Test WriteFile
	testFile := "test/nested/dir/test.txt"
	testContent := []byte("test content")
	err = fs.WriteFile(testFile, testContent, 0644)
	assert.NoError(t, err, "WriteFile failed")

	// Test ReadFile
	content, err := fs.ReadFile(testFile)
	assert.NoError(t, err, "ReadFile failed")
	assert.Equal(t, string(testContent), string(content), "ReadFile returned wrong content")

	// Test ReadDir
	entries, err := fs.ReadDir(testDir)
	assert.NoError(t, err, "ReadDir failed")
	assert.Equal(t, 1, len(entries), "ReadDir returned wrong number of entries")
	assert.Equal(t, "test.txt", entries[0].Name(), "ReadDir returned wrong entry name")
}

func TestMockFileSystemWalkDir(t *testing.T) {
	fs := NewMockFileSystem()
	
	// Create a test structure
	fs.MkdirAll("test/nested", 0755)
	fs.WriteFile("test/file1.txt", []byte("content1"), 0644)
	fs.WriteFile("test/nested/file2.txt", []byte("content2"), 0644)
	
	// Test WalkDir
	foundFiles := 0
	foundDirs := 0
	err := fs.WalkDir("test", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			foundDirs++
		} else {
			foundFiles++
		}
		return nil
	})
	assert.NoError(t, err, "WalkDir failed")
	assert.GreaterOrEqual(t, foundFiles, 1, "WalkDir found wrong number of files")
	assert.GreaterOrEqual(t, foundDirs, 1, "WalkDir found wrong number of directories")
}

// TestMockFileSystemStat tests the Stat method of the mock file system
func TestMockFileSystemStat(t *testing.T) {
	fs := NewMockFileSystem()
	
	// Add a directory
	fs.AddDirectory("test-dir")
	
	// Add a file
	fileContent := []byte("test content")
	fs.AddFile("test-file.txt", fileContent)
	
	// Test getting info for a directory
	dirInfo, err := fs.Stat("test-dir")
	assert.NoError(t, err)
	assert.True(t, dirInfo.IsDir())
	assert.Equal(t, "test-dir", dirInfo.Name())
	
	// Test getting info for a file
	fileInfo, err := fs.Stat("test-file.txt")
	assert.NoError(t, err)
	assert.False(t, fileInfo.IsDir())
	assert.Equal(t, "test-file.txt", fileInfo.Name())
	assert.Equal(t, int64(len(fileContent)), fileInfo.Size())
	
	// Test getting info for a non-existent file
	_, err = fs.Stat("non-existent-file.txt")
	assert.Error(t, err)
} 