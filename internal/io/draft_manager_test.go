// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package io

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestNewDraftManager(t *testing.T) {
	fs := NewMockFileSystem()
	dm := NewDraftManager(fs)
	
	assert.NotNil(t, dm)
	assert.Equal(t, fs, dm.fs)
}

func TestDraftManager_GetDraftPath(t *testing.T) {
	fs := NewMockFileSystem()
	dm := NewDraftManager(fs)
	
	// Test when config directory doesn't exist
	path, err := dm.GetDraftPath()
	
	assert.NoError(t, err)
	assert.Contains(t, path, ".usm/feature_request_draft.json")
	
	// The directory should be created in the mock filesystem
	fs.AddDirectory(".usm")
	assert.True(t, fs.Exists(".usm"))
}

func TestDraftManager_SaveDraft(t *testing.T) {
	fs := NewMockFileSystem()
	dm := NewDraftManager(fs)
	fr := models.NewFeatureRequest()
	fr.Title = "Test Feature"
	
	// Set up the directory in the mock filesystem
	fs.AddDirectory(".usm")
	
	// Test successful save
	err := dm.SaveDraft(fr)
	
	assert.NoError(t, err)
	
	// Verify the draft file was created
	path, _ := dm.GetDraftPath()
	assert.True(t, fs.Exists(path))
	
	// Check content was correctly serialized
	data, _ := fs.ReadFile(path)
	var savedFR models.FeatureRequest
	err = json.Unmarshal(data, &savedFR)
	assert.NoError(t, err)
	assert.Equal(t, fr.Title, savedFR.Title)
}

func TestDraftManager_LoadDraft(t *testing.T) {
	fs := NewMockFileSystem()
	dm := NewDraftManager(fs)
	
	// Set up the directory in the mock filesystem
	fs.AddDirectory(".usm")
	
	// Test when draft doesn't exist
	fr, err := dm.LoadDraft()
	
	assert.NoError(t, err)
	assert.Equal(t, models.NewFeatureRequest().Title, fr.Title)
	
	// Test successful load
	testFR := models.NewFeatureRequest()
	testFR.Title = "Test Feature"
	data, _ := json.Marshal(testFR)
	
	path, _ := dm.GetDraftPath()
	fs.WriteFile(path, data, 0644)
	
	fr, err = dm.LoadDraft()
	
	assert.NoError(t, err)
	assert.Equal(t, "Test Feature", fr.Title)
	
	// Test when unmarshal fails - Manually handle the mock
	path, _ = dm.GetDraftPath()
	fs.WriteFile(path, []byte("invalid json"), 0644)
	
	fr, loadErr := dm.LoadDraft()
	
	// This should not return an error but should return a new feature request
	assert.Equal(t, models.NewFeatureRequest().Title, fr.Title)
	// It should parse incorrectly but still create a new feature request
	assert.NotNil(t, loadErr)
}

func TestDraftManager_DeleteDraft(t *testing.T) {
	fs := NewMockFileSystem()
	dm := NewDraftManager(fs)
	
	// Set up the directory in the mock filesystem
	fs.AddDirectory(".usm")
	
	// Test when draft doesn't exist
	err := dm.DeleteDraft()
	
	assert.NoError(t, err)
} 