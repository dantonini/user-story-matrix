// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package io

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MockFileInfo implements os.FileInfo for testing purposes
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// Name returns the base name of the file
func (m MockFileInfo) Name() string {
	return m.name
}

// Size returns the length in bytes for regular files
func (m MockFileInfo) Size() int64 {
	return m.size
}

// Mode returns the file mode bits
func (m MockFileInfo) Mode() os.FileMode {
	return m.mode
}

// ModTime returns the modification time
func (m MockFileInfo) ModTime() time.Time {
	return m.modTime
}

// IsDir reports whether the file is a directory
func (m MockFileInfo) IsDir() bool {
	return m.isDir
}

// Sys returns the underlying system data
func (m MockFileInfo) Sys() interface{} {
	return nil
}

// MockFileEntry implements os.DirEntry for testing purposes
type MockFileEntry struct {
	name  string
	isDir bool
}

// Name returns the name of the file
func (m MockFileEntry) Name() string {
	return m.name
}

// IsDir reports whether the file is a directory
func (m MockFileEntry) IsDir() bool {
	return m.isDir
}

// Type returns the file mode
func (m MockFileEntry) Type() os.FileMode {
	if m.isDir {
		return os.ModeDir
	}
	return 0
}

// Info returns file info for the file
func (m MockFileEntry) Info() (os.FileInfo, error) {
	return MockFileInfo{
		name:    m.name,
		isDir:   m.isDir,
		mode:    m.Type(),
		modTime: time.Now(),
	}, nil
}

// MockFileSystem is an in-memory file system for testing
type MockFileSystem struct {
	mu       sync.RWMutex
	Files    map[string][]byte
	DirItems map[string][]os.DirEntry
	DirInfo  map[string]os.FileInfo
	FileInfo map[string]os.FileInfo
	// Track write operations for testing
	WriteOps []FileWriteOperation
}

// FileWriteOperation tracks write operations for testing
type FileWriteOperation struct {
	Path    string
	Content []byte
	Mode    os.FileMode
	Time    time.Time
}

// NewMockFileSystem creates a new in-memory file system for testing
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:    make(map[string][]byte),
		DirItems: make(map[string][]os.DirEntry),
		DirInfo:  make(map[string]os.FileInfo),
		FileInfo: make(map[string]os.FileInfo),
		WriteOps: make([]FileWriteOperation, 0),
	}
}

// AddDirectory adds a mock directory
func (fs *MockFileSystem) AddDirectory(path string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)

	fs.DirItems[path] = []os.DirEntry{}
	fs.DirInfo[path] = MockFileInfo{
		name:    filepath.Base(path),
		isDir:   true,
		mode:    os.ModeDir | 0755,
		modTime: time.Now(),
	}

	// Ensure parent directories exist
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" && dir != path {
		fs.mu.Unlock() // Avoid deadlock
		fs.AddDirectory(dir)
		fs.mu.Lock()
	}
}

// AddFile adds a mock file with content
func (fs *MockFileSystem) AddFile(path string, content []byte) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)

	// Make a copy of the content to avoid unexpected modifications
	contentCopy := make([]byte, len(content))
	copy(contentCopy, content)

	fs.Files[path] = contentCopy
	dir := filepath.Dir(path)
	
	// Create directory if it doesn't exist
	if _, exists := fs.DirItems[dir]; !exists {
		fs.mu.Unlock() // Avoid deadlock
		fs.AddDirectory(dir)
		fs.mu.Lock()
	}
	
	// Add file to directory entries if not already there
	fileEntry := MockFileEntry{
		name:  filepath.Base(path),
		isDir: false,
	}

	// Check if this file already exists in the directory entries
	var exists bool
	for _, entry := range fs.DirItems[dir] {
		if entry.Name() == fileEntry.Name() {
			exists = true
			break
		}
	}

	// Only add to directory entries if it doesn't already exist
	if !exists {
		fs.DirItems[dir] = append(fs.DirItems[dir], fileEntry)
	}
	
	// Add or update file info
	fs.FileInfo[path] = MockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(contentCopy)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}

	// Track this write operation
	fs.WriteOps = append(fs.WriteOps, FileWriteOperation{
		Path:    path,
		Content: contentCopy,
		Mode:    0644,
		Time:    time.Now(),
	})
}

// ReadDir reads the directory named by dirname and returns a list of directory entries
func (fs *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)

	if entries, exists := fs.DirItems[path]; exists {
		return entries, nil
	}
	return nil, fmt.Errorf("directory not found: %s", path)
}

// ReadFile reads the file named by filename and returns the contents
func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)

	if content, exists := fs.Files[path]; exists {
		// Return a copy of the content to avoid unexpected modifications
		contentCopy := make([]byte, len(content))
		copy(contentCopy, content)
		return contentCopy, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// WriteFile writes data to a file named by filename
func (fs *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)

	// We need to do this in multiple steps with proper locking
	fs.mu.Lock()
	
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	needCreateDir := false
	if _, exists := fs.DirItems[dir]; !exists {
		needCreateDir = true
	}
	
	// First release the lock if we need to create a directory
	if needCreateDir {
		fs.mu.Unlock()
		err := fs.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		fs.mu.Lock()
	}
	
	// Make a copy of the data to avoid unexpected modifications
	contentCopy := make([]byte, len(data))
	copy(contentCopy, data)
	
	// Update the file content
	fs.Files[path] = contentCopy
	
	// Create or update file info
	fs.FileInfo[path] = MockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(contentCopy)),
		mode:    perm,
		modTime: time.Now(),
		isDir:   false,
	}
	
	// Add file to directory entries if not already there
	fileEntry := MockFileEntry{
		name:  filepath.Base(path),
		isDir: false,
	}
	
	// Check if this file already exists in the directory entries
	var exists bool
	dirEntries := fs.DirItems[dir]
	for _, entry := range dirEntries {
		if entry.Name() == fileEntry.Name() {
			exists = true
			break
		}
	}
	
	// Only add to directory entries if it doesn't already exist
	if !exists {
		fs.DirItems[dir] = append(dirEntries, fileEntry)
	}
	
	// Track this write operation
	fs.WriteOps = append(fs.WriteOps, FileWriteOperation{
		Path:    path,
		Content: contentCopy,
		Mode:    perm,
		Time:    time.Now(),
	})
	
	fs.mu.Unlock()
	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents
func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)
	
	// Create all parent directories
	parts := strings.Split(path, string(filepath.Separator))
	current := ""
	
	for i, part := range parts {
		if i == 0 && part == "" {
			// Handle absolute paths that start with /
			current = string(filepath.Separator)
			continue
		}
		
		if current == "" {
			current = part
		} else if current == string(filepath.Separator) {
			current = filepath.Join(current, part)
		} else {
			current = filepath.Join(current, part)
		}
		
		// Create directory if it doesn't exist
		fs.mu.RLock()
		_, exists := fs.DirItems[current]
		fs.mu.RUnlock()

		if !exists {
			fs.AddDirectory(current)
		}
	}
	
	return nil
}

// Exists checks if a file or directory exists
func (fs *MockFileSystem) Exists(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)
	
	_, fileExists := fs.Files[path]
	_, dirExists := fs.DirItems[path]
	return fileExists || dirExists
}

// Stat returns file info for the named file
func (fs *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)
	
	// Check if it's a file
	if info, exists := fs.FileInfo[path]; exists {
		return info, nil
	}
	
	// Check if it's a directory
	if info, exists := fs.DirInfo[path]; exists {
		return info, nil
	}
	
	return nil, fmt.Errorf("file or directory not found: %s", path)
}

// GetLastWrite returns the last write operation for a file
func (fs *MockFileSystem) GetLastWrite(path string) (FileWriteOperation, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Normalize path to avoid inconsistencies
	path = filepath.Clean(path)
	
	// Look for the last write operation for this path
	for i := len(fs.WriteOps) - 1; i >= 0; i-- {
		if fs.WriteOps[i].Path == path {
			// Return a copy to avoid modifications
			op := fs.WriteOps[i]
			contentCopy := make([]byte, len(op.Content))
			copy(contentCopy, op.Content)
			
			return FileWriteOperation{
				Path:    op.Path,
				Content: contentCopy,
				Mode:    op.Mode,
				Time:    op.Time,
			}, true
		}
	}
	return FileWriteOperation{}, false
}

// WalkDir walks the file tree rooted at root, calling fn for each file or directory
func (fs *MockFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	// Normalize path to avoid inconsistencies
	root = filepath.Clean(root)
	
	fs.mu.RLock()
	
	// First check if root exists
	if !fs.Exists(root) {
		fs.mu.RUnlock()
		return fmt.Errorf("root directory not found: %s", root)
	}
	
	// Create a queue for BFS (breadth-first search)
	queue := []string{root}
	fs.mu.RUnlock()
	
	// Process the queue
	for len(queue) > 0 {
		// Dequeue
		path := queue[0]
		queue = queue[1:]
		
		fs.mu.RLock()
		// Get info
		info, err := fs.Stat(path)
		if err != nil {
			fs.mu.RUnlock()
			if err := fn(path, nil, err); err != nil && err != filepath.SkipDir {
				return err
			}
			continue
		}
		
		// Create DirEntry
		entry := MockFileEntry{
			name:  info.Name(),
			isDir: info.IsDir(),
		}
		
		// Process current node
		fs.mu.RUnlock()
		err = fn(path, entry, nil)
		if err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
		
		// Enqueue children if it's a directory
		if info.IsDir() {
			fs.mu.RLock()
			if entries, exists := fs.DirItems[path]; exists {
				for _, childEntry := range entries {
					childPath := filepath.Join(path, childEntry.Name())
					queue = append(queue, childPath)
				}
			}
			fs.mu.RUnlock()
		}
	}
	
	return nil
} 