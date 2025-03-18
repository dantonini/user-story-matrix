package io

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MockFileSystem implements FileSystem interface for testing
type MockFileSystem struct {
	// Map of paths to file contents
	Files map[string][]byte
	// Map of directory paths
	Dirs map[string]bool
}

// MockDirEntry implements os.DirEntry interface for testing
type MockDirEntry struct {
	name  string
	isDir bool
}

func (e *MockDirEntry) Name() string {
	return e.name
}

func (e *MockDirEntry) IsDir() bool {
	return e.isDir
}

func (e *MockDirEntry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}

func (e *MockDirEntry) Info() (fs.FileInfo, error) {
	// Not implemented for mock
	return nil, nil
}

// NewMockFileSystem creates a new instance of MockFileSystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

// ReadDir returns a slice of mock directory entries
func (fs *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	if !fs.Exists(path) {
		return nil, os.ErrNotExist
	}

	// If path is not a directory, return an error
	if !fs.Dirs[path] {
		return nil, os.ErrInvalid
	}

	// Gather all files and directories that are direct children of path
	prefix := path + "/"
	if path == "" || path == "." {
		prefix = ""
	}

	var entries []os.DirEntry
	// Find all directories that are direct children of the given path
	for dir := range fs.Dirs {
		if strings.HasPrefix(dir, prefix) {
			// Get the relative path to the directory
			remaining := strings.TrimPrefix(dir, prefix)
			// If the remaining path contains a '/', it's a nested directory
			if strings.Contains(remaining, "/") {
				continue
			}
			// Empty remaining means this is the same directory, not a child
			if remaining != "" {
				entries = append(entries, &MockDirEntry{
					name:  remaining,
					isDir: true,
				})
			}
		}
	}

	// Find all files that are direct children of the given path
	for file := range fs.Files {
		if strings.HasPrefix(file, prefix) {
			// Get the relative path to the file
			remaining := strings.TrimPrefix(file, prefix)
			// If the remaining path contains a '/', it's in a nested directory
			if strings.Contains(remaining, "/") {
				continue
			}
			entries = append(entries, &MockDirEntry{
				name:  remaining,
				isDir: false,
			})
		}
	}

	// Sort entries by name for consistent results
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

// ReadFile returns file contents
func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, ok := fs.Files[path]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

// WriteFile writes data to a file
func (fs *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if dir != "." && !fs.Exists(dir) {
		if err := fs.MkdirAll(dir, perm); err != nil {
			return err
		}
	}
	fs.Files[path] = data
	return nil
}

// MkdirAll creates all directories in the path
func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if path == "" {
		return nil
	}

	// Adding directory and all parent directories
	parts := strings.Split(path, "/")
	currentPath := ""
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i > 0 {
			currentPath += "/"
		}
		currentPath += part
		fs.Dirs[currentPath] = true
	}
	return nil
}

// Exists checks if a file or directory exists
func (fs *MockFileSystem) Exists(path string) bool {
	if path == "" || path == "." {
		return true
	}
	_, fileExists := fs.Files[path]
	_, dirExists := fs.Dirs[path]
	return fileExists || dirExists
}

// WalkDir simulates walking the file tree
func (fs *MockFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	if !fs.Exists(root) {
		return os.ErrNotExist
	}

	// Call fn for the root directory
	rootEntry := &MockDirEntry{
		name:  filepath.Base(root),
		isDir: true,
	}
	if err := fn(root, rootEntry, nil); err != nil {
		return err
	}

	// Walk all directories
	var allDirs []string
	for dir := range fs.Dirs {
		if dir == root || strings.HasPrefix(dir, root+"/") {
			allDirs = append(allDirs, dir)
		}
	}
	sort.Strings(allDirs)

	for _, dir := range allDirs {
		if dir == root {
			continue // Already processed
		}
		dirEntry := &MockDirEntry{
			name:  filepath.Base(dir),
			isDir: true,
		}
		if err := fn(dir, dirEntry, nil); err != nil {
			return err
		}
	}

	// Walk all files
	var allFiles []string
	for file := range fs.Files {
		if strings.HasPrefix(file, root+"/") {
			allFiles = append(allFiles, file)
		}
	}
	sort.Strings(allFiles)

	for _, file := range allFiles {
		fileEntry := &MockDirEntry{
			name:  filepath.Base(file),
			isDir: false,
		}
		if err := fn(file, fileEntry, nil); err != nil {
			return err
		}
	}

	return nil
} 