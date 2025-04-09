// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package workflow provides functionality for managing and executing structured implementation workflows.
// It includes support for defining, registering, and executing workflow steps for user story implementation.
package workflow

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Standard workflow constants
const (
	// StandardWorkflowName is the identifier for the default workflow
	// This constant is used to retrieve the standard workflow from the registry
	StandardWorkflowName = "standard"
)

// Global registry instance with thread-safe access
var (
	globalRegistry     *WorkflowRegistry
	globalRegistryOnce sync.Once
	globalRegistryLock sync.RWMutex
)

// WorkflowDefinition represents a complete workflow with metadata and steps.
// It encapsulates all information needed to execute a workflow, including
// its name, description, and sequence of steps.
//
// This structure is part of the refactoring to support custom workflows
// while maintaining backward compatibility with the existing StandardWorkflowSteps.
type WorkflowDefinition struct {
	// Name uniquely identifies the workflow (e.g., "standard", "custom-tutorial")
	Name string

	// Description provides a human-readable explanation of the workflow's purpose
	Description string

	// Steps contains the ordered sequence of steps that make up this workflow
	Steps []WorkflowStep
}

// workflowCache stores loaded workflows for improved performance
type workflowCache struct {
	workflows map[string]*WorkflowDefinition // Cached workflow definitions
	sources   map[string]string             // Maps workflow name to source path
	modified  map[string]time.Time          // Last modified timestamps for cache invalidation
}

// WorkflowRegistry manages available workflows and provides methods for retrieving them.
// It acts as a central repository for all workflow definitions, allowing the system
// to support multiple workflows while maintaining a consistent interface.
//
// The registry is initialized with the standard workflow and can be extended
// with custom workflows in future implementations.
type WorkflowRegistry struct {
	// builtInWorkflows maps workflow names to their definitions
	builtInWorkflows map[string]*WorkflowDefinition
	// cache provides performance optimization for file-based workflows
	cache workflowCache
	// mutex protects concurrent access to the workflows map
	mutex sync.RWMutex
}

// GetGlobalRegistry returns the singleton global registry instance.
// This ensures that all workflow managers share the same registry.
// The global registry is initialized on first access.
//
// Returns:
//   - The global WorkflowRegistry instance
func GetGlobalRegistry() *WorkflowRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = newWorkflowRegistry()
	})
	return globalRegistry
}

// ResetGlobalRegistry resets the global registry to a new instance.
// This is primarily useful for testing scenarios where isolation
// between tests is required.
//
// Returns:
//   - The newly reset global WorkflowRegistry instance
func ResetGlobalRegistry() *WorkflowRegistry {
	globalRegistryLock.Lock()
	defer globalRegistryLock.Unlock()
	
	globalRegistry = newWorkflowRegistry()
	return globalRegistry
}

// NewWorkflowRegistry creates a new registry with the standard workflow pre-registered.
// This ensures that the standard workflow is always available as a fallback option.
// In most cases, GetGlobalRegistry() should be used instead to ensure all components
// share the same registry.
//
// Returns:
//   - A new WorkflowRegistry instance with the standard workflow already registered
func NewWorkflowRegistry() *WorkflowRegistry {
	globalRegistryLock.RLock()
	defer globalRegistryLock.RUnlock()
	
	// Return the global registry to ensure registry sharing
	if globalRegistry == nil {
		// Initialize the global registry if it doesn't exist yet
		globalRegistry = newWorkflowRegistry()
	}
	
	return globalRegistry
}

// newWorkflowRegistry creates a new isolated registry instance.
// This is internal and should not be used directly except by
// the global registry management functions.
//
// Returns:
//   - A new independent WorkflowRegistry instance
func newWorkflowRegistry() *WorkflowRegistry {
	registry := &WorkflowRegistry{
		builtInWorkflows: make(map[string]*WorkflowDefinition),
		cache: workflowCache{
			workflows: make(map[string]*WorkflowDefinition),
			sources:   make(map[string]string),
			modified:  make(map[string]time.Time),
		},
		mutex: sync.RWMutex{},
	}

	// Register the standard workflow
	registry.RegisterBuiltInWorkflow(createStandardWorkflow())

	return registry
}

// RegisterBuiltInWorkflow adds a workflow to the registry.
// This method allows for the registration of new workflow definitions
// that can then be retrieved by name.
//
// Parameters:
//   - workflow: The WorkflowDefinition to register
func (r *WorkflowRegistry) RegisterBuiltInWorkflow(workflow *WorkflowDefinition) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.builtInWorkflows[workflow.Name] = workflow
}

// GetWorkflow retrieves a workflow by name.
// It returns an error if the requested workflow is not found in the registry.
//
// Parameters:
//   - name: The unique identifier of the workflow to retrieve
//
// Returns:
//   - The requested WorkflowDefinition, or an error if it doesn't exist
func (r *WorkflowRegistry) GetWorkflow(name string) (*WorkflowDefinition, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// First check built-in workflows
	workflow, exists := r.builtInWorkflows[name]
	if exists {
		return workflow, nil
	}
	
	// Then check cached workflows from filesystem
	workflow, exists = r.cache.workflows[name]
	if exists {
		return workflow, nil
	}
	
	return nil, fmt.Errorf("workflow '%s' not found", name)
}

// GetStandardWorkflow returns the standard workflow.
// This is a convenience method that always returns the standard workflow,
// which is guaranteed to exist in the registry.
//
// Returns:
//   - The standard WorkflowDefinition
func (r *WorkflowRegistry) GetStandardWorkflow() *WorkflowDefinition {
	// The standard workflow is always registered during registry initialization,
	// so if it's not found, it's a programming error, not a runtime error
	workflow, err := r.GetWorkflow(StandardWorkflowName)
	if err != nil {
		// This should never happen as the standard workflow is registered in the constructor
		panic(fmt.Sprintf("Standard workflow not found in registry: %v", err))
	}
	return workflow
}

// ListWorkflows returns a list of names of all workflows in the registry.
// This is useful for displaying available workflows to users.
//
// Returns:
//   - A slice of workflow names
func (r *WorkflowRegistry) ListWorkflows() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Collect all workflow names (built-in and file-based)
	workflows := make([]string, 0, len(r.builtInWorkflows)+len(r.cache.workflows))
	
	for name := range r.builtInWorkflows {
		workflows = append(workflows, name)
	}
	
	for name := range r.cache.workflows {
		// Skip if already added from built-in workflows
		if _, exists := r.builtInWorkflows[name]; !exists {
			workflows = append(workflows, name)
		}
	}
	
	return workflows
}

// LoadFromDirectory loads a workflow from a directory
// The directory should contain a workflow.yaml file and a prompts/ subdirectory
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - path: Path to the workflow directory
//
// Returns:
//   - The loaded WorkflowDefinition, or an error if loading failed
func (r *WorkflowRegistry) LoadFromDirectory(fs FileSystem, path string) (*WorkflowDefinition, error) {
	// Check if directory exists
	if !fs.Exists(path) {
		return nil, fmt.Errorf("workflow directory not found: %s", path)
	}
	
	// Check if workflow.yaml exists
	workflowYAMLPath := filepath.Join(path, StandardWorkflowYAML)
	if !fs.Exists(workflowYAMLPath) {
		return nil, fmt.Errorf("workflow.yaml not found in %s", path)
	}
	
	// Read and parse workflow.yaml
	data, err := fs.ReadFile(workflowYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow.yaml: %w", err)
	}
	
	var fileDef WorkflowFileDefinition
	if err := yaml.Unmarshal(data, &fileDef); err != nil {
		return nil, fmt.Errorf("invalid YAML in workflow.yaml: %w", err)
	}
	
	// Validate the workflow definition
	if fileDef.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	
	// Create workflow definition
	workflow := &WorkflowDefinition{
		Name:        fileDef.Name,
		Description: fileDef.Description,
		Steps:       make([]WorkflowStep, len(fileDef.Steps)),
	}
	
	// Process each step and load its prompt
	for i, fileStep := range fileDef.Steps {
		// Validate step
		if fileStep.ID == "" {
			return nil, fmt.Errorf("step ID is required for step %d", i)
		}
		
		// Create the step
		step := WorkflowStep{
			ID:          fileStep.ID,
			Description: fileStep.Description,
			Prompt:      "", // Will be loaded from file
		}
		
		// Resolve prompt file path
		promptPath := filepath.Join(path, fileStep.Prompt)
		if !fs.Exists(promptPath) {
			return nil, fmt.Errorf("prompt file not found: %s", promptPath)
		}
		
		// Load prompt content
		promptData, err := fs.ReadFile(promptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
		}
		
		// Set prompt content and source
		step.Prompt = string(promptData)
		step.source = promptSource{
			sourceType: promptSourceFile,
			filePath:   promptPath,
		}
		
		workflow.Steps[i] = step
	}
	
	// Add to cache
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.cache.workflows[workflow.Name] = workflow
	r.cache.sources[workflow.Name] = path
	
	// Get last modified time for caching
	// For simplicity in MVI, we'll just use current time
	r.cache.modified[workflow.Name] = time.Now()
	
	return workflow, nil
}

// DiscoverWorkflows finds and loads workflows from standard locations
//
// Parameters:
//   - fs: FileSystem interface for file operations
//
// Returns:
//   - Map of workflow names to their definitions
func (r *WorkflowRegistry) DiscoverWorkflows(fs FileSystem) map[string]*WorkflowDefinition {
	results := make(map[string]*WorkflowDefinition)
	
	// Get standard workflow directories
	directories := GetStandardWorkflowDirectories()
	
	// Try to load workflow from each directory
	for _, dir := range directories {
		if !fs.Exists(dir) {
			continue
		}
		
		workflow, err := r.LoadFromDirectory(fs, dir)
		if err != nil {
			// Log the error but continue with other directories
			fmt.Printf("Error loading workflow from %s: %v\n", dir, err)
			continue
		}
		
		// Add to results
		results[workflow.Name] = workflow
	}
	
	return results
}

// ReloadChangedWorkflows checks for modified workflow files and reloads them
//
// Parameters:
//   - fs: FileSystem interface for file operations
//
// Returns:
//   - Slice of names of reloaded workflows
func (r *WorkflowRegistry) ReloadChangedWorkflows(fs FileSystem) []string {
	reloaded := make([]string, 0)
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Check each cached workflow
	for name, path := range r.cache.sources {
		changed, err := r.isWorkflowModified(fs, name)
		if err != nil {
			// Log the error but continue with other workflows
			fmt.Printf("Error checking if workflow %s is modified: %v\n", name, err)
			continue
		}
		
		if changed {
			// Reload the workflow
			workflow, err := r.LoadFromDirectory(fs, path)
			if err != nil {
				// Log the error but continue with other workflows
				fmt.Printf("Error reloading workflow %s: %v\n", name, err)
				continue
			}
			
			// Update cache
			r.cache.workflows[name] = workflow
			r.cache.modified[name] = time.Now()
			
			reloaded = append(reloaded, name)
		}
	}
	
	return reloaded
}

// isWorkflowModified checks if a workflow file has been modified
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - name: Name of the workflow to check
//
// Returns:
//   - true if the workflow has been modified, false otherwise
//   - error if checking failed
func (r *WorkflowRegistry) isWorkflowModified(fs FileSystem, name string) (bool, error) {
	// For MVI, we'll just assume workflows haven't changed to keep it simple
	// In a real implementation, we would check file modification times
	return false, nil
}

// GetStandardWorkflowDirectories returns potential workflow locations
//
// Returns:
//   - Slice of standard workflow directory paths
func GetStandardWorkflowDirectories() []string {
	return []string{
		StandardTemplateDir,
		"workflows",                // Project root workflows
		"~/.config/usm/workflows",  // User-specific workflows
		"/etc/usm/workflows",       // System-wide workflows
	}
}

// createStandardWorkflow converts the existing StandardWorkflowSteps to a WorkflowDefinition.
// This function provides backward compatibility by wrapping the existing global
// StandardWorkflowSteps variable in the new WorkflowDefinition structure.
//
// Returns:
//   - A WorkflowDefinition that encapsulates the existing StandardWorkflowSteps
func createStandardWorkflow() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        StandardWorkflowName,
		Description: "Standard USM implementation workflow",
		Steps:       StandardWorkflowSteps,
	}
}

// Custom Workflow Implementation Plan:
//
// Phase 1: COMPLETED - Refactor StandardWorkflowSteps structure (dev-01)
// ✓ Refactored workflow structure with WorkflowDefinition and WorkflowRegistry
// ✓ Created global registry instance for cross-component access
// ✓ Maintained backward compatibility with legacy code
//
// Phase 2: Extract prompt files from standard workflow (dev-02)
// - Extract long prompts from StandardWorkflowSteps into separate Markdown files
// - Organize files in standard directory structure for workflow templates
// - Generate workflow.yaml from the current StandardWorkflowSteps metadata
// - Implement mechanism to load prompts from files with fallback to embedded prompts
//
// Phase 3: Add workflow loading from filesystem (dev-03)
// - Extend WorkflowRegistry to load workflow definitions from disk
// - Implement discovery of workflows in standard locations
// - Add validation of workflow.yaml format and prompt references
// - Create caching mechanism for loaded workflows
// - Add logging for workflow loading operations
//
// Phase 4: Update workflow state format (dev-04)
// - Update WorkflowState to include workflow identification
// - Maintain backward compatibility with existing state files
// - Update WorkflowManager methods to handle the new state format
// - Implement validation for workflow switching
//
// Phase 5: Implement template variables support (dev-05)
// - Add Variables field to WorkflowStep struct
// - Implement template processing system using Go's text/template
// - Support variable substitution, default values, conditionals, and iteration
// - Add validation and error handling for template processing
//
// Phase 6: Deprecate StandardWorkflowSteps (dev-06)
// - Mark StandardWorkflowSteps as deprecated
// - Add linter rules to flag direct usage
// - Implement compatibility layer for legacy code
// - Update documentation and provide migration guides
//
// Phase 7: Migrate legacy workflows (dev-07)
// - Convert StandardWorkflowSteps to a built-in workflow template
// - Ensure compatibility with existing state files
// - Provide backward compatibility for direct code references
// - Create migration command for advanced users
// - Document migration process
