package io

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/user-story-matrix/usm/internal/models"
)

// DraftManager handles feature request drafts
type DraftManager struct {
	fs FileSystem
}

// NewDraftManager creates a new draft manager
func NewDraftManager(fs FileSystem) *DraftManager {
	return &DraftManager{fs: fs}
}

// GetDraftPath returns the path to the draft file
func (dm *DraftManager) GetDraftPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".usm")
	if !dm.fs.Exists(configDir) {
		if err := dm.fs.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}
	
	return filepath.Join(configDir, "feature_request_draft.json"), nil
}

// SaveDraft saves a feature request draft to disk
func (dm *DraftManager) SaveDraft(fr models.FeatureRequest) error {
	draftPath, err := dm.GetDraftPath()
	if err != nil {
		return err
	}
	
	data, err := json.Marshal(fr)
	if err != nil {
		return err
	}
	
	return dm.fs.WriteFile(draftPath, data, 0644)
}

// LoadDraft loads a feature request draft from disk
func (dm *DraftManager) LoadDraft() (models.FeatureRequest, error) {
	draftPath, err := dm.GetDraftPath()
	if err != nil {
		return models.NewFeatureRequest(), err
	}
	
	if !dm.fs.Exists(draftPath) {
		return models.NewFeatureRequest(), nil
	}
	
	data, err := dm.fs.ReadFile(draftPath)
	if err != nil {
		return models.NewFeatureRequest(), err
	}
	
	var fr models.FeatureRequest
	if err := json.Unmarshal(data, &fr); err != nil {
		return models.NewFeatureRequest(), err
	}
	
	return fr, nil
}

// DeleteDraft deletes the draft file
func (dm *DraftManager) DeleteDraft() error {
	draftPath, err := dm.GetDraftPath()
	if err != nil {
		return err
	}
	
	if dm.fs.Exists(draftPath) {
		return os.Remove(draftPath)
	}
	
	return nil
} 