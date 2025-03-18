package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/models"
)

func TestDisplayImplementationMessage(t *testing.T) {
	// Create a mock terminal
	mockTerminal := io.NewMockIO()

	// Create a test change request
	cr := models.ChangeRequest{
		FilePath: "docs/changes-request/2025-03-18-073012-implement-a-change-request.blueprint.md",
	}

	// Call the function
	displayImplementationMessage(mockTerminal, cr)

	// Assert expectations
	assert.Len(t, mockTerminal.Messages, 1)
	assert.Contains(t, mockTerminal.Messages[0], "Read the blueprint file in")
	assert.Contains(t, mockTerminal.Messages[0], "validate the blueprint against the code base")
}

func TestDisplayNoChangeRequestsMessage(t *testing.T) {
	// Create a mock terminal
	mockTerminal := io.NewMockIO()

	// Call the function
	displayNoChangeRequestsMessage(mockTerminal)

	// Assert expectations
	assert.Len(t, mockTerminal.Messages, 1)
	assert.Contains(t, mockTerminal.Messages[0], "No change requests to implement")
	assert.Contains(t, mockTerminal.Messages[0], "create change-request")
} 