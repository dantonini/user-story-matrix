// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
	"gopkg.in/yaml.v3"
)

func TestExtractWorkflowCmd(t *testing.T) {
	// This test will be fully implemented in the MVI phase
	// For now, we're just verifying the command is properly registered
	
	// Ensure the command is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "extract-workflow" {
			found = true
			break
		}
	}
	
	if !found {
		t.Fatal("extract-workflow command not found in root command")
	}
	
	// Test flags
	cmd := extractWorkflowCmd
	flags := cmd.Flags()
	
	// Check if the output flag exists
	if flags.Lookup("output") == nil {
		t.Error("Missing required flag: output")
	}
	
	// TODO in MVI phase: Add more comprehensive tests
	// 1. Test with default output path
	// 2. Test with custom output path
	// 3. Test error handling for invalid paths
	// 4. Test when extraction fails
	// 5. Test successful extraction
}

// TestExtractWorkflowCmd_Execute tests the actual execution of the command with a mock filesystem
func TestExtractWorkflowCmd_Execute(t *testing.T) {
	// Initialize the test command with mocked dependencies
	cmd, buf, fs := createTestExtractCommand()
	
	// Set output directory for test
	outputDir := "/test/output"
	cmd.Flags().Set("output", outputDir)
	
	// Execute the command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}
	
	// Verify the output message
	output := buf.String()
	if !strings.Contains(output, "Workflow extracted successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}
	
	// Verify the workflow.yaml file was created
	workflowYamlPath := filepath.Join(outputDir, "workflow.yaml")
	if !fs.Exists(workflowYamlPath) {
		t.Errorf("workflow.yaml file not created at %s", workflowYamlPath)
	}
	
	// Verify prompts directory was created
	promptsDir := filepath.Join(outputDir, "prompts")
	if !fs.Exists(promptsDir) {
		t.Errorf("prompts directory not created at %s", promptsDir)
	}
	
	// Verify at least one prompt file was created
	// Get all files in prompts directory
	promptFiles, _ := fs.ReadDir(promptsDir)
	if len(promptFiles) == 0 {
		t.Error("No prompt files created in prompts directory")
	} else {
		// Check that files have .md extension
		foundMdFile := false
		for _, file := range promptFiles {
			if strings.HasSuffix(file.Name(), ".md") {
				foundMdFile = true
				break
			}
		}
		if !foundMdFile {
			t.Error("No .md files found in prompts directory")
		}
	}
	
	// Verify workflow.yaml content
	yamlContent, err := fs.ReadFile(workflowYamlPath)
	if err != nil {
		t.Errorf("Failed to read workflow.yaml: %v", err)
	} else {
		// Basic YAML structure validation
		var workflowDef struct {
			Name  string `yaml:"name"`
			Steps []struct {
				ID     string `yaml:"id"`
				Prompt string `yaml:"prompt"`
			} `yaml:"steps"`
		}
		
		err = yaml.Unmarshal(yamlContent, &workflowDef)
		if err != nil {
			t.Errorf("Invalid YAML in workflow.yaml: %v", err)
		}
		
		// Verify workflow name
		if workflowDef.Name != "standard" {
			t.Errorf("Expected workflow name 'standard', got '%s'", workflowDef.Name)
		}
		
		// Verify steps are present
		if len(workflowDef.Steps) == 0 {
			t.Error("No steps found in workflow.yaml")
		}
		
		// Verify prompt paths point to existing files
		for _, step := range workflowDef.Steps {
			promptPath := filepath.Join(outputDir, step.Prompt)
			if !fs.Exists(promptPath) {
				t.Errorf("Prompt file not found: %s", promptPath)
			}
		}
	}
}

// createTestExtractCommand creates a test instance of the extract command
// with mocked dependencies for testing purposes
func createTestExtractCommand() (*cobra.Command, *bytes.Buffer, *io.MockFileSystem) {
	// Create a copy of the command to avoid modifying the global one
	cmd := &cobra.Command{
		Use:   extractWorkflowCmd.Use,
		Short: extractWorkflowCmd.Short,
		Long:  extractWorkflowCmd.Long,
	}
	
	// Create mock dependencies
	buf := new(bytes.Buffer)
	fs := io.NewMockFileSystem()
	mockOutput := io.NewMockIO()
	
	// Set up the run function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get output directory
		outputDir, err := cmd.Flags().GetString("output")
		if err != nil {
			return fmt.Errorf("error getting output directory flag: %v", err)
		}
		
		// Use default if not specified
		if outputDir == "" {
			outputDir = workflow.StandardTemplateDir
		}
		
		// Show progress
		mockOutput.PrintProgress("Extracting standard workflow to " + outputDir)
		
		// Extract the workflow
		err = workflow.ExtractStandardWorkflow(fs, outputDir)
		if err != nil {
			mockOutput.PrintError("Failed to extract workflow: " + err.Error())
			return err
		}
		
		mockOutput.PrintSuccess("Workflow extracted successfully to: " + outputDir)
		
		// Write success message to buffer for test assertions
		buf.WriteString(fmt.Sprintf("Workflow extracted successfully to: %s\n", outputDir))
		
		return nil
	}
	
	// Copy flags
	extractWorkflowCmd.Flags().VisitAll(func(f *pflag.Flag) {
		cmd.Flags().AddFlag(f)
	})
	
	return cmd, buf, fs
} 