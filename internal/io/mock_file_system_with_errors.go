package io

import (
	"io/fs"
	"os"
)

// MockFileSystemWithErrors extends MockFileSystem with error simulation capabilities
type MockFileSystemWithErrors struct {
	MockFileSystem
	writeErrors map[string]error
	readErrors  map[string]error
	statErrors  map[string]error
}

// NewMockFileSystemWithErrors creates a new MockFileSystemWithErrors
func NewMockFileSystemWithErrors() *MockFileSystemWithErrors {
	return &MockFileSystemWithErrors{
		MockFileSystem: *NewMockFileSystem(),
		writeErrors:    make(map[string]error),
		readErrors:     make(map[string]error),
		statErrors:     make(map[string]error),
	}
}

// SetWriteError sets an error to be returned when writing to a specific path
func (fs *MockFileSystemWithErrors) SetWriteError(path string, err error) {
	fs.writeErrors[path] = err
}

// SetReadError sets an error to be returned when reading from a specific path
func (fs *MockFileSystemWithErrors) SetReadError(path string, err error) {
	fs.readErrors[path] = err
}

// SetStatError sets an error to be returned when getting stats for a specific path
func (fs *MockFileSystemWithErrors) SetStatError(path string, err error) {
	fs.statErrors[path] = err
}

// WriteFile overrides MockFileSystem's WriteFile to simulate errors
func (fs *MockFileSystemWithErrors) WriteFile(path string, data []byte, perm os.FileMode) error {
	if err, ok := fs.writeErrors[path]; ok {
		return err
	}
	return fs.MockFileSystem.WriteFile(path, data, perm)
}

// ReadFile overrides MockFileSystem's ReadFile to simulate errors
func (fs *MockFileSystemWithErrors) ReadFile(path string) ([]byte, error) {
	if err, ok := fs.readErrors[path]; ok {
		return nil, err
	}
	return fs.MockFileSystem.ReadFile(path)
}

// Stat overrides MockFileSystem's Stat to simulate errors
func (fs *MockFileSystemWithErrors) Stat(path string) (os.FileInfo, error) {
	if err, ok := fs.statErrors[path]; ok {
		return nil, err
	}
	return fs.MockFileSystem.Stat(path)
}

// ReadDir overrides MockFileSystem's ReadDir
func (fs *MockFileSystemWithErrors) ReadDir(path string) ([]os.DirEntry, error) {
	return fs.MockFileSystem.ReadDir(path)
}

// MkdirAll overrides MockFileSystem's MkdirAll
func (fs *MockFileSystemWithErrors) MkdirAll(path string, perm os.FileMode) error {
	return fs.MockFileSystem.MkdirAll(path, perm)
}

// WalkDir overrides MockFileSystem's WalkDir
func (fs *MockFileSystemWithErrors) WalkDir(root string, fn fs.WalkDirFunc) error {
	return fs.MockFileSystem.WalkDir(root, fn)
}

// Exists overrides MockFileSystem's Exists to simulate errors
func (fs *MockFileSystemWithErrors) Exists(path string) bool {
	if _, ok := fs.statErrors[path]; ok {
		return false
	}
	return fs.MockFileSystem.Exists(path)
} 