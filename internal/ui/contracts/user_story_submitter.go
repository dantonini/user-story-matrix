// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package contracts

import (
	"github.com/user-story-matrix/usm/internal/models"
)

// UserStorySubmitter defines the interface for user story form components
// that can provide a user story and confirm submission status
type UserStorySubmitter interface {
	// GetUserStory returns the user story data from the form
	GetUserStory() models.UserStory
	
	// GetConfirmSubmission returns whether the user confirmed the submission
	// If false, the form was canceled or closed without submitting
	GetConfirmSubmission() bool
}

// Deprecated: FormResult is the old name for UserStorySubmitter
// This alias is kept for backward compatibility and should be removed in a future version
type FormResult = UserStorySubmitter

// UserStoryFormResult extends UserStorySubmitter with user story specific methods
type UserStoryFormResult interface {
	UserStorySubmitter
	
	// Additional methods specific to user story forms could be added here
} 