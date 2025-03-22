// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


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
	fr.UserStory = "As a user I want to test features so that I can verify they work"
	fr.AcceptanceCriteria = []string{
		"Feature should be testable",
		"Feature should work correctly",
	}
	
	form := NewFeatureForm(fr)
	
	assert.NotNil(t, form)
	assert.Equal(t, fr, form.fr)
	assert.Equal(t, "Test Feature", form.titleInput.Value())
	assert.Equal(t, "This is a test feature", form.descInput.Value())
	assert.Equal(t, "user", form.userStoryAsInput.Value())
	assert.Equal(t, "to test features", form.userStoryWantInput.Value())
	assert.Equal(t, "I can verify they work", form.userStorySoThatInput.Value())
	assert.Equal(t, "Feature should be testable", form.acInputs[0].Value())
	assert.Equal(t, "Feature should work correctly", form.acInputs[1].Value())
}

func TestFeatureFormSaveDraft(t *testing.T) {
	fr := models.NewFeatureRequest()
	form := NewFeatureForm(fr)
	
	// Set some form values
	form.titleInput.SetValue("Draft Feature")
	form.descInput.SetValue("This is a draft feature")
	form.userStoryAsInput.SetValue("user")
	form.userStoryWantInput.SetValue("to save drafts")
	form.userStorySoThatInput.SetValue("I can resume later")
	form.acInputs[0].SetValue("Draft should be saveable")
	form.acInputs[1].SetValue("Draft should be resumable")
	
	// Save the draft
	savedFR := form.SaveDraft()
	
	// Verify the saved values
	assert.Equal(t, "Draft Feature", savedFR.Title)
	assert.Equal(t, "This is a draft feature", savedFR.Description)
	assert.Equal(t, "As a user I want to save drafts so that I can resume later", savedFR.UserStory)
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
	form.userStoryAsInput.SetValue("user")
	form.userStoryWantInput.SetValue("to update features")
	form.userStorySoThatInput.SetValue("I can improve them")
	form.acInputs[0].SetValue("Update should work")
	form.acInputs[1].SetValue("Update should be easy")
	
	// Update the feature request
	form.updateFeatureRequest()
	
	// Verify the updated values
	assert.Equal(t, "Updated Feature", form.fr.Title)
	assert.Equal(t, "This feature was updated", form.fr.Description)
	assert.Equal(t, "As a user I want to update features so that I can improve them", form.fr.UserStory)
	assert.Equal(t, 2, len(form.fr.AcceptanceCriteria))
	assert.Equal(t, "Update should work", form.fr.AcceptanceCriteria[0])
	assert.Equal(t, "Update should be easy", form.fr.AcceptanceCriteria[1])
}

func TestGetFeatureRequest(t *testing.T) {
	fr := models.NewFeatureRequest()
	fr.Title = "Complete Feature"
	fr.Description = "This is a complete feature"
	fr.UserStory = "As a user I want to get features so that I can use them"
	fr.AcceptanceCriteria = []string{
		"Feature should be gettable",
		"Feature should be complete",
	}
	
	form := NewFeatureForm(fr)
	
	// Get the feature request
	gotFR := form.GetFeatureRequest()
	
	// Verify the values
	assert.Equal(t, fr, gotFR)
}

func TestEmptyFieldsNoUserStory(t *testing.T) {
	fr := models.NewFeatureRequest()
	form := NewFeatureForm(fr)
	
	// Simulate tabbing through all fields without entering any values
	form.nextField() // Title -> Description
	form.nextField() // Description -> UserStoryAs
	form.nextField() // UserStoryAs -> UserStoryWant
	form.nextField() // UserStoryWant -> UserStorySoThat
	form.nextField() // UserStorySoThat -> AC1
	form.nextField() // AC1 -> AC2
	form.nextField() // AC2 -> AC3
	form.nextField() // AC3 -> AC4
	form.nextField() // AC4 -> AC5
	form.nextField() // AC5 -> Review
	
	// Get the feature request
	savedFR := form.GetFeatureRequest()
	
	// Verify that no user story was created
	assert.Equal(t, "", savedFR.UserStory)
	assert.Equal(t, "", savedFR.Title)
	assert.Equal(t, "", savedFR.Description)
	assert.Equal(t, 0, len(savedFR.AcceptanceCriteria))
} 