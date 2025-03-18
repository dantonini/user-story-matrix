package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestNewFeatureForm(t *testing.T) {
	fr := models.NewFeatureRequest()
	fr.Title = "Test Feature"
	fr.Description = "This is a test feature"
	fr.Importance = "It's very important"
	fr.UserStory = "As a user, I want to test features, so that I can verify they work"
	fr.AcceptanceCriteria = []string{"Feature should be testable", "Feature should work correctly"}
	
	form := NewFeatureForm(fr)
	
	assert.NotNil(t, form)
	assert.Equal(t, fr, form.fr)
	assert.Equal(t, "Test Feature", form.titleInput.Value())
	assert.Equal(t, "This is a test feature", form.descInput.Value())
	assert.Equal(t, "It's very important", form.importanceInput.Value())
	assert.Equal(t, "As a user, I want to test features, so that I can verify they work", form.userStoryInput.Value())
	assert.Equal(t, "Feature should be testable\nFeature should work correctly", form.acInput.Value())
}

func TestFeatureFormSaveDraft(t *testing.T) {
	fr := models.NewFeatureRequest()
	form := NewFeatureForm(fr)
	
	// Set some form values
	form.titleInput.SetValue("Draft Feature")
	form.descInput.SetValue("This is a draft feature")
	form.importanceInput.SetValue("Important draft")
	form.userStoryInput.SetValue("As a user, I want to save drafts")
	form.acInput.SetValue("Draft should be saveable\nDraft should be resumable")
	
	// Save the draft
	savedFR := form.SaveDraft()
	
	// Verify the saved values
	assert.Equal(t, "Draft Feature", savedFR.Title)
	assert.Equal(t, "This is a draft feature", savedFR.Description)
	assert.Equal(t, "Important draft", savedFR.Importance)
	assert.Equal(t, "As a user, I want to save drafts", savedFR.UserStory)
	assert.Equal(t, 2, len(savedFR.AcceptanceCriteria))
	assert.Equal(t, "Draft should be saveable", savedFR.AcceptanceCriteria[0])
	assert.Equal(t, "Draft should be resumable", savedFR.AcceptanceCriteria[1])
}

func TestUpdateFeatureRequest(t *testing.T) {
	fr := models.NewFeatureRequest()
	form := NewFeatureForm(fr)
	
	// Set form values
	form.titleInput.SetValue("Updated Feature")
	form.descInput.SetValue("This feature was updated")
	form.importanceInput.SetValue("Very important update")
	form.userStoryInput.SetValue("As a user, I want to update features")
	form.acInput.SetValue("Update should work\nUpdate should be easy")
	
	// Update the feature request
	form.updateFeatureRequest()
	
	// Verify the updated values
	assert.Equal(t, "Updated Feature", form.fr.Title)
	assert.Equal(t, "This feature was updated", form.fr.Description)
	assert.Equal(t, "Very important update", form.fr.Importance)
	assert.Equal(t, "As a user, I want to update features", form.fr.UserStory)
	assert.Equal(t, 2, len(form.fr.AcceptanceCriteria))
	assert.Equal(t, "Update should work", form.fr.AcceptanceCriteria[0])
	assert.Equal(t, "Update should be easy", form.fr.AcceptanceCriteria[1])
}

func TestGetFeatureRequest(t *testing.T) {
	fr := models.NewFeatureRequest()
	fr.Title = "Complete Feature"
	fr.Description = "This is a complete feature"
	fr.Importance = "It's critically important"
	fr.UserStory = "As a user, I want to get features"
	fr.AcceptanceCriteria = []string{"Feature should be gettable", "Feature should be complete"}
	
	form := NewFeatureForm(fr)
	
	// Get the feature request
	gotFR := form.GetFeatureRequest()
	
	// Verify the values
	assert.Equal(t, fr, gotFR)
} 