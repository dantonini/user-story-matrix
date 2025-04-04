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
	Files    map[string][]byte
	DirItems map[string][]os.DirEntry
	DirInfo  map[string]os.FileInfo
	FileInfo map[string]os.FileInfo
}

// NewMockFileSystem creates a new in-memory file system for testing
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:    make(map[string][]byte),
		DirItems: make(map[string][]os.DirEntry),
		DirInfo:  make(map[string]os.FileInfo),
		FileInfo: make(map[string]os.FileInfo),
	}
}

// AddDirectory adds a mock directory
func (fs *MockFileSystem) AddDirectory(path string) {
	fs.DirItems[path] = []os.DirEntry{}
	fs.DirInfo[path] = MockFileInfo{
		name:    filepath.Base(path),
		isDir:   true,
		mode:    os.ModeDir | 0755,
		modTime: time.Now(),
	}
}

// AddFile adds a mock file with content
func (fs *MockFileSystem) AddFile(path string, content []byte) {
	fs.Files[path] = content
	dir := filepath.Dir(path)
	
	// Create directory if it doesn't exist
	if _, exists := fs.DirItems[dir]; !exists {
		fs.AddDirectory(dir)
	}
	
	// Add file to directory entries
	fs.DirItems[dir] = append(fs.DirItems[dir], MockFileEntry{
		name:  filepath.Base(path),
		isDir: false,
	})
	
	// Add file info
	fs.FileInfo[path] = MockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}
}

// ReadDir reads the directory named by dirname and returns a list of directory entries
func (fs *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	if entries, exists := fs.DirItems[path]; exists {
		return entries, nil
	}
	return nil, fmt.Errorf("directory not found: %s", path)
}

// ReadFile reads the file named by filename and returns the contents
func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, exists := fs.Files[path]; exists {
		return content, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// WriteFile writes data to a file named by filename
func (fs *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	
	// Create directory if it doesn't exist
	if _, exists := fs.DirItems[dir]; !exists {
		fs.AddDirectory(dir)
	}
	
	// Add or update file
	fs.AddFile(path, data)
	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents
func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
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
		if _, exists := fs.DirItems[current]; !exists {
			fs.AddDirectory(current)
		}
	}
	
	return nil
}

// Exists checks if a file or directory exists
func (fs *MockFileSystem) Exists(path string) bool {
	_, fileExists := fs.Files[path]
	_, dirExists := fs.DirItems[path]
	return fileExists || dirExists
}

// Stat returns file info for the named file
func (fs *MockFileSystem) Stat(path string) (os.FileInfo, error) {
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

// WalkDir walks the file tree rooted at root, calling fn for each file or directory
func (fs *MockFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	// First process the root directory
	if !fs.Exists(root) {
		return fmt.Errorf("root directory not found: %s", root)
	}
	
	// Create a queue for BFS
	queue := []string{root}
	
	for len(queue) > 0 {
		// Dequeue
		path := queue[0]
		queue = queue[1:]
		
		// Get info
		info, err := fs.Stat(path)
		if err != nil {
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
		err = fn(path, entry, nil)
		if err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
		
		// Enqueue children if it's a directory
		if info.IsDir() {
			if entries, exists := fs.DirItems[path]; exists {
				for _, childEntry := range entries {
					childPath := filepath.Join(path, childEntry.Name())
					queue = append(queue, childPath)
				}
			}
		}
	}
	
	return nil
} 