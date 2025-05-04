// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/workflow"
)

// ValidationResult represents the result of a workflow validation operation
type ValidationResult struct {
	Success      bool
	WorkflowName string
	Errors       []string
	Warnings     []string
	ErrorMessage string
}

// ValidateCmd represents the workflow validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate [name-or-path]",
	Short: "Validate a workflow definition",
	Long: `Validate a workflow definition by name or path.

This command validates that:
- The workflow.yaml file is properly formatted
- Required fields are present and valid
- Step IDs are unique
- Referenced prompt files exist
- Template syntax in prompt files is valid

You can provide either:
- A workflow name (to validate a registered workflow)
- A path to a workflow directory

Examples:
  # Validate a workflow by name
  usm workflow validate my-workflow

  # Validate a workflow by path
  usm workflow validate path/to/workflow
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameOrPath := args[0]
		
		output := io.NewTerminalIO()
		fs := io.NewOSFileSystem()
		
		// Call the business logic function
		result := validateWorkflow(nameOrPath, fs, output)
		
		// Display progress message first
		output.PrintProgress(fmt.Sprintf("Validating workflow '%s'", result.WorkflowName))
		
		// Handle the result
		if !result.Success {
			// Display any errors
			output.PrintError(result.ErrorMessage)
			
			// Display specific errors if available
			if len(result.Errors) > 0 {
				for _, err := range result.Errors {
					output.PrintError(fmt.Sprintf("- %s", err))
				}
			}
			
			os.Exit(1)
			return
		}
		
		// Display successful validation results
		output.PrintSuccess("✅ Workflow is valid")
		
		// Always display warnings when present, even for valid workflows
		if len(result.Warnings) > 0 {
			output.PrintWarning(fmt.Sprintf("\n⚠️  Found %d warning(s) that should be addressed:", len(result.Warnings)))
			
			// Group warnings by type for better organization
			warningsByStep := make(map[string][]string)
			unusedVarWarnings := []string{}
			missingVarWarnings := []string{}
			
			for _, warning := range result.Warnings {
				if strings.Contains(warning, "not used in template") {
					unusedVarWarnings = append(unusedVarWarnings, warning)
				} else if strings.Contains(warning, "uses variable") && strings.Contains(warning, "but it is not provided") {
					missingVarWarnings = append(missingVarWarnings, warning)
				} else {
					// Extract step ID from the warning
					stepID := extractStepIDFromWarning(warning)
					warningsByStep[stepID] = append(warningsByStep[stepID], warning)
				}
			}
			
			// Display missing variable warnings with suggested fixes
			if len(missingVarWarnings) > 0 {
				output.PrintWarning("\n🔍 MISSING VARIABLES")
				output.Print("   Variables used in templates but not defined in the workflow:")
				for _, warning := range missingVarWarnings {
					// Extract key information from the warning
					stepID, varName, templatePath := extractInfoFromMissingVarWarning(warning)
					
					// Create more user-friendly message with fix suggestion
					output.PrintWarning(fmt.Sprintf("   • Step '%s': Variable '%s' in '%s'", 
						stepID, varName, templatePath))
					output.Print(fmt.Sprintf("     ↳ Fix: Add to workflow.yaml in step '%s':", stepID))
					output.Print(fmt.Sprintf("         variables:"))
					output.Print(fmt.Sprintf("           %s: \"your_value_here\"", varName))
				}
				
				output.Print("\n   Example variable usage in templates:")
				output.Print("     {{.variable_name}} - Basic usage")
				output.Print("     {{.variable_name | default \"fallback\"}} - With default value")
			}
			
			// Display unused variable warnings
			if len(unusedVarWarnings) > 0 {
				output.PrintWarning("\n⚠️  UNUSED VARIABLES")
				output.Print("   Variables defined in the workflow but not used in templates:")
				for _, warning := range unusedVarWarnings {
					// Extract key information from the warning
					varName, stepID, templatePath := extractInfoFromUnusedVarWarning(warning)
					
					// Create more user-friendly message with fix suggestion
					output.PrintWarning(fmt.Sprintf("   • Step '%s': Variable '%s' in '%s'", 
						stepID, varName, templatePath))
					output.Print(fmt.Sprintf("     ↳ Fix options:"))
					output.Print(fmt.Sprintf("       1. Use the variable in template: {{.%s}}", varName))
					output.Print(fmt.Sprintf("       2. Remove from workflow.yaml if not needed"))
				}
			}
			
			// Display other warnings
			if len(warningsByStep) > 0 {
				output.PrintWarning("\n⚙️  OTHER WARNINGS")
				for stepID, warnings := range warningsByStep {
					if len(warnings) > 0 {
						output.PrintWarning(fmt.Sprintf("   Step '%s':", stepID))
						for _, warning := range warnings {
							output.Print(fmt.Sprintf("   • %s", warning))
						}
					}
				}
			}
			
			// Provide a summary footer with next steps
			output.PrintWarning("\n🛠️  RECOMMENDED NEXT STEPS")
			output.Print("   1. Fix the warnings shown above")
			output.Print("   2. Run 'usm workflow validate' again to confirm")
		}
	},
}

// validateWorkflow implements the business logic for workflow validation
// It extracts the validation logic from the command's Run function to make it testable
//
// Parameters:
//   - nameOrPath: Either a workflow name or path to a workflow directory
//   - fs: The filesystem interface for file operations
//   - output: The output interface for user feedback (used only for logging, not errors)
//
// Returns:
//   - ValidationResult containing success status and validation details
func validateWorkflow(nameOrPath string, fs io.FileSystem, output io.UserOutput) ValidationResult {
	// Get the global registry
	registry := workflow.GetGlobalRegistry()
	
	// Discover available workflows
	// This is required to find workflows that were created by the user
	// but might not have been loaded into the registry yet
	discoveredWorkflows := registry.DiscoverWorkflows(fs)
	
	// Determine if input is a path or a name
	var workflowPath string
	var workflowDef *workflow.WorkflowDefinition
	
	if fs.Exists(nameOrPath) && isValidWorkflowDir(fs, nameOrPath) {
		// Input is a path
		workflowPath = nameOrPath
		
		// Try to load the workflow from this path
		workflowYAMLPath := filepath.Join(workflowPath, workflow.WorkflowConfigFile)
		var err error
		workflowDef, err = workflow.LoadWorkflowFromFile(fs, workflowYAMLPath)
		if err != nil {
			return ValidationResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Failed to load workflow: %s", err.Error()),
			}
		}
	} else {
		// Assume input is a name
		// First, check if the workflow was just discovered
		if discoveredWf, exists := discoveredWorkflows[nameOrPath]; exists {
			workflowDef = discoveredWf
		} else {
			// If not found in discoveries, try to get it from the registry
			var err error
			workflowDef, err = registry.GetWorkflow(nameOrPath)
			if err != nil {
				return ValidationResult{
					Success:      false,
					ErrorMessage: fmt.Sprintf("Workflow '%s' not found. Use 'usm workflow list' to see available workflows.", nameOrPath),
				}
			}
		}
		
		// Now that we have the workflow definition, find its path
		// First check for source path in the registry's cache
		source, registryPath := registry.GetWorkflowSourceInfo(nameOrPath)
		
		if registryPath != "-" && registryPath != "" {
			// Important: If the registryPath doesn't include the full workflow path,
			// we need to construct it correctly to prevent path resolution issues
			
			// If path already refers to the specific workflow directory
			if fs.Exists(filepath.Join(registryPath, workflow.WorkflowConfigFile)) {
				workflowPath = registryPath
			} else if fs.Exists(filepath.Join(registryPath, nameOrPath, workflow.WorkflowConfigFile)) {
				// If path is a parent directory containing the workflow
				workflowPath = filepath.Join(registryPath, nameOrPath)
			} else if fs.Exists(filepath.Join(registryPath, "workflows", nameOrPath, workflow.WorkflowConfigFile)) {
				// Handle case where registryPath is something like ".usm" 
				// and the full path should be ".usm/workflows/workflow-name"
				workflowPath = filepath.Join(registryPath, "workflows", nameOrPath)
			} else {
				// Just use the path as provided
				workflowPath = registryPath
			}
		}
		
		// If still no path, try to find it by name
		if workflowPath == "" {
			workflowPath = findWorkflowByName(fs, nameOrPath)
			if workflowPath == "" {
				// For built-in workflows, we don't need a path for validation
				if source == workflow.SourceBuiltIn {
					// Built-in workflows don't need prompt validation
					return ValidationResult{
						Success:      true,
						WorkflowName: workflowDef.Name,
						Warnings:     []string{"Built-in workflow validated without prompt reference checks"},
					}
				}
				
				// For other workflow types, we need a path to validate prompt references
				return ValidationResult{
					Success:      false,
					ErrorMessage: fmt.Sprintf("Cannot find workflow directory for '%s'", nameOrPath),
				}
			}
		}
	}
	
	// If workflowPath is not a complete workflow directory path, ensure it includes the workflow name
	if !isValidWorkflowDir(fs, workflowPath) && workflowDef != nil {
		// Check if it's a parent directory that needs the workflow name appended
		possiblePath := filepath.Join(workflowPath, workflowDef.Name)
		if isValidWorkflowDir(fs, possiblePath) {
			workflowPath = possiblePath
		}
	}
	
	// Debug: Log the path being used for validation
	output.PrintProgress(fmt.Sprintf("Using workflow path: %s", workflowPath))
	
	// Create validator with the workflow path
	validator := workflow.NewWorkflowValidator(fs, workflowPath)
	
	// Validate workflow
	result, err := validator.ValidateWorkflow(workflowDef)
	if err != nil {
		return ValidationResult{
			Success:      false,
			WorkflowName: workflowDef.Name,
			ErrorMessage: fmt.Sprintf("Validation failed: %s", err.Error()),
		}
	}
	
	// Create the result based on validation
	if result.IsValid() {
		return ValidationResult{
			Success:      true,
			WorkflowName: workflowDef.Name,
			Warnings:     result.Warnings,
		}
	} else {
		return ValidationResult{
			Success:      false,
			WorkflowName: workflowDef.Name,
			Errors:       result.Errors,
			ErrorMessage: fmt.Sprintf("Workflow validation failed with %d errors", len(result.Errors)),
		}
	}
}

// isValidWorkflowDir checks if a directory contains a valid workflow structure
func isValidWorkflowDir(fs io.FileSystem, dirPath string) bool {
	workflowYAMLPath := filepath.Join(dirPath, workflow.WorkflowConfigFile)
	promptsDirPath := filepath.Join(dirPath, workflow.PromptsDir)
	
	return fs.Exists(workflowYAMLPath) && fs.Exists(promptsDirPath)
}

// findWorkflowByName looks for a workflow by name in standard locations
func findWorkflowByName(fs io.FileSystem, name string) string {
	// Get standard workflow directories
	dirs := workflow.GetStandardWorkflowDirectories()
	
	// Check each directory for the named workflow
	for _, dir := range dirs {
		if !fs.Exists(dir) {
			continue
		}
		
		// Check for a direct match in the workflow directory
		workflowPath := filepath.Join(dir, name)
		if isValidWorkflowDir(fs, workflowPath) {
			return workflowPath
		}
		
		// If direct match not found, check subdirectories
		// This helps with discovered workflows in directories like .usm/workflows
		entries, err := fs.ReadDir(dir)
		if err != nil {
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				subDirPath := filepath.Join(dir, entry.Name())
				
				// Check if this is the workflow we're looking for
				if entry.Name() == name && isValidWorkflowDir(fs, subDirPath) {
					return subDirPath
				}
			}
		}
	}
	
	return ""
}

// Helper function to extract step ID from a warning message
func extractStepIDFromWarning(warning string) string {
	// Default if we can't extract
	stepID := "unknown"
	
	// Try to extract step ID using regex
	re := regexp.MustCompile(`step '([^']+)'`)
	matches := re.FindStringSubmatch(warning)
	if len(matches) >= 2 {
		stepID = matches[1]
	}
	
	return stepID
}

// Helper function to extract information from missing variable warnings
func extractInfoFromMissingVarWarning(warning string) (string, string, string) {
	stepID := "unknown"
	varName := "unknown"
	templatePath := "unknown"
	
	// Extract step ID
	stepRe := regexp.MustCompile(`step '([^']+)'`)
	stepMatches := stepRe.FindStringSubmatch(warning)
	if len(stepMatches) >= 2 {
		stepID = stepMatches[1]
	}
	
	// Extract variable name
	varRe := regexp.MustCompile(`variable '([^']+)'`)
	varMatches := varRe.FindStringSubmatch(warning)
	if len(varMatches) >= 2 {
		varName = varMatches[1]
	}
	
	// Extract template path
	templateRe := regexp.MustCompile(`template '([^']+)'`)
	templateMatches := templateRe.FindStringSubmatch(warning)
	if len(templateMatches) >= 2 {
		templatePath = templateMatches[1]
	}
	
	return stepID, varName, templatePath
}

// Helper function to extract information from unused variable warnings
func extractInfoFromUnusedVarWarning(warning string) (string, string, string) {
	varName := "unknown"
	stepID := "unknown"
	templatePath := "unknown"
	
	// Extract variable name
	varRe := regexp.MustCompile(`variable '([^']+)'`)
	varMatches := varRe.FindStringSubmatch(warning)
	if len(varMatches) >= 2 {
		varName = varMatches[1]
	}
	
	// Extract step ID
	stepRe := regexp.MustCompile(`step '([^']+)'`)
	stepMatches := stepRe.FindStringSubmatch(warning)
	if len(stepMatches) >= 2 {
		stepID = stepMatches[1]
	}
	
	// Extract template path
	templateRe := regexp.MustCompile(`template '([^']+)'`)
	templateMatches := templateRe.FindStringSubmatch(warning)
	if len(templateMatches) >= 2 {
		templatePath = templateMatches[1]
	}
	
	return varName, stepID, templatePath
}

func init() {
	// No flags required for this command
} 