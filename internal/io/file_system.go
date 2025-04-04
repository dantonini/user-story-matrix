// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package io

import (
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem is an interface for file system operations
type FileSystem interface {
	// ReadDir reads the named directory and returns a list of directory entries
	ReadDir(dirname string) ([]os.DirEntry, error)
	
	// ReadFile reads the file at the specified path and returns its contents
	ReadFile(filename string) ([]byte, error)
	
	// WriteFile writes data to a file at the specified path
	WriteFile(filename string, data []byte, perm os.FileMode) error
	
	// MkdirAll creates a directory with the specified name and permission, along with any necessary parents
	MkdirAll(path string, perm os.FileMode) error
	
	// Stat returns a FileInfo describing the named file
	Stat(name string) (os.FileInfo, error)
	
	// WalkDir walks the file tree rooted at root, calling fn for each file or directory
	WalkDir(root string, fn fs.WalkDirFunc) error
	
	// Exists checks if a file or directory exists
	Exists(path string) bool
}

// OSFileSystem implements FileSystem interface with standard os operations
type OSFileSystem struct{}

// NewOSFileSystem creates a new instance of OSFileSystem
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadDir reads the directory named by dirname and returns a list of directory entries
func (fs *OSFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// ReadFile reads the file named by filename and returns the contents
func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file named by filename
func (fs *OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Exists checks if a file or directory exists
func (fs *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Stat returns file info for the named file
func (fs *OSFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// WalkDir walks the file tree rooted at root, calling fn for each file or directory
func (fs *OSFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
} 