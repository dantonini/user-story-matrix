// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

// Standard workflow template paths
const (
	// StandardWorkflowYAML is the filename for the standard workflow definition
	StandardWorkflowYAML = "workflow.yaml"
)

// promptSourceType identifies where a prompt comes from
type promptSourceType int //nolint:unused

const (
	promptSourceEmbedded promptSourceType = iota //nolint:unused
	promptSourceFile //nolint:unused
)

// promptSource tracks the origin of a prompt
type promptSource struct { //nolint:unused
	sourceType promptSourceType
	filePath   string // Only used when sourceType is promptSourceFile
}

// WorkflowFileDefinition represents the structure of a workflow.yaml file
type WorkflowFileDefinition struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Steps       []WorkflowFileStep `yaml:"steps"`
}

// WorkflowFileStep represents a step in a workflow.yaml file
type WorkflowFileStep struct {
	ID          string              `yaml:"id"`
	Description string              `yaml:"description"`
	Prompt      string              `yaml:"prompt"` // Path to prompt file, relative to workflow dir
	Variables   map[string]string   `yaml:"variables,omitempty"` // Variables for template substitution
}
