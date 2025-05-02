// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package workflow provides functionality for managing and executing structured implementation workflows.
// It includes support for defining, registering, and executing workflow steps for user story implementation.
package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
	"go.uber.org/zap"
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

// FileSystem is defined in the io package
// Use io.FileSystem instead of defining it here

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
	
	// Notify callbacks about the workflow change
	notifyWorkflowCallbacks(workflow.Name, workflow)
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
func (r *WorkflowRegistry) LoadFromDirectory(fs io.FileSystem, path string) (*WorkflowDefinition, error) {
	// Check if directory exists
	if !fs.Exists(path) {
		return nil, fmt.Errorf("workflow directory not found: %s", path)
	}
	
	// Check if workflow.yaml exists
	workflowYAMLPath := filepath.Join(path, StandardWorkflowYAML)
	if !fs.Exists(workflowYAMLPath) {
		// Try workflow.json as an alternative
		workflowJSONPath := filepath.Join(path, "workflow.json")
		if !fs.Exists(workflowJSONPath) {
			return nil, fmt.Errorf("neither workflow.yaml nor workflow.json found in %s", path)
		}
		
		// Load workflow from JSON
		workflow, err := LoadWorkflowFromFile(fs, workflowJSONPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load workflow from %s: %w", workflowJSONPath, err)
		}
		
		// Add to cache
		r.mutex.Lock()
		r.cache.workflows[workflow.Name] = workflow
		r.cache.sources[workflow.Name] = workflowJSONPath
		r.cache.modified[workflow.Name] = time.Now()
		r.mutex.Unlock()
		
		// Update steps with prompt content from files
		for i := range workflow.Steps {
			step := &workflow.Steps[i]
			
			// The prompt field in the file should contain a path to the prompt file
			// Check if it's a relative path or an embedded prompt
			if strings.HasPrefix(step.Prompt, "prompts/") || filepath.Ext(step.Prompt) == ".md" {
				promptPath := step.Prompt
				// If it's a relative path, resolve it
				if !filepath.IsAbs(promptPath) {
					promptPath = filepath.Join(path, promptPath)
				}
				
				// Check if prompt file exists
				if fs.Exists(promptPath) {
					// Read prompt content
					promptData, err := fs.ReadFile(promptPath)
					if err != nil {
						return nil, fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
					}
					
					// Set prompt content and mark as file-sourced
					step.Prompt = string(promptData)
					step.source = promptSource{
						sourceType: promptSourceFile,
						filePath:   promptPath,
					}
				} else {
					return nil, fmt.Errorf("prompt file %s referenced in workflow but not found", promptPath)
				}
			}
		}
		
		return workflow, nil
	}
	
	// Load workflow from YAML
	workflow, err := LoadWorkflowFromFile(fs, workflowYAMLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow from %s: %w", workflowYAMLPath, err)
	}
	
	// Add to cache
	r.mutex.Lock()
	r.cache.workflows[workflow.Name] = workflow
	r.cache.sources[workflow.Name] = workflowYAMLPath
	r.cache.modified[workflow.Name] = time.Now()
	r.mutex.Unlock()
	
	// Update steps with prompt content from files
	for i := range workflow.Steps {
		step := &workflow.Steps[i]
		
		// The prompt field in the file should contain a path to the prompt file
		// Check if it's a relative path or an embedded prompt
		if strings.HasPrefix(step.Prompt, "prompts/") || filepath.Ext(step.Prompt) == ".md" {
			promptPath := step.Prompt
			// If it's a relative path, resolve it
			if !filepath.IsAbs(promptPath) {
				promptPath = filepath.Join(path, promptPath)
			}
			
			// Check if prompt file exists
			if fs.Exists(promptPath) {
				// Read prompt content
				promptData, err := fs.ReadFile(promptPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
				}
				
				// Set prompt content and mark as file-sourced
				step.Prompt = string(promptData)
				step.source = promptSource{
					sourceType: promptSourceFile,
					filePath:   promptPath,
				}
			} else {
				return nil, fmt.Errorf("prompt file %s referenced in workflow but not found", promptPath)
			}
		}
	}
	
	return workflow, nil
}

// DiscoverWorkflows finds and loads workflows from standard locations
// This method searches in multiple directories and loads workflows from each.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//
// Returns:
//   - A map of workflow names to their definitions
func (r *WorkflowRegistry) DiscoverWorkflows(fs io.FileSystem) map[string]*WorkflowDefinition {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	discoveredWorkflows := make(map[string]*WorkflowDefinition)
	
	// Get standard workflow directories
	directories := GetStandardWorkflowDirectories()
	
	// Use logger instead of fmt.Printf for debug output
	logger.Debug("Searching directories", zap.Strings("directories", directories))
	
	// Function to process a workflow YAML or JSON file
	processWorkflowFile := func(filePath string) {
		logger.Debug("Processing workflow file", zap.String("file", filePath))
		workflow, err := LoadWorkflowFromFile(fs, filePath)
		if err != nil {
			logger.Error("Error loading workflow", zap.String("file", filePath), zap.Error(err))
			return
		}
		
		logger.Debug("Successfully loaded workflow", 
			zap.String("name", workflow.Name), 
			zap.String("file", filePath))
		
		// Add to cache with proper source tracking
		r.cache.workflows[workflow.Name] = workflow
		r.cache.sources[workflow.Name] = filePath
		r.cache.modified[workflow.Name] = time.Now()
		discoveredWorkflows[workflow.Name] = workflow
	}
	
	// Load workflows from each directory
	for _, dir := range directories {
		logger.Debug("Checking directory", zap.String("dir", dir))
		if !fs.Exists(dir) {
			logger.Debug("Directory does not exist", zap.String("dir", dir))
			continue
		}
		
		// Check for workflow files directly in this directory
		entries, err := fs.ReadDir(dir)
		if err != nil {
			logger.Error("Error reading workflow directory", 
				zap.String("dir", dir), 
				zap.Error(err))
			continue
		}
		
		logger.Debug("Found entries in directory", 
			zap.String("dir", dir), 
			zap.Int("count", len(entries)))
		
		// First look for direct workflow files (workflow.yaml or workflow.json)
		for _, entry := range entries {
			entryName := entry.Name()
			isDir := entry.IsDir()
			logger.Debug("Found entry", 
				zap.String("name", entryName), 
				zap.Bool("isDir", isDir))
			
			if entry.IsDir() {
				continue
			}
			
			name := entry.Name()
			if name == "workflow.yaml" || name == "workflow.yml" || name == "workflow.json" {
				workflowPath := filepath.Join(dir, name)
				logger.Debug("Found workflow file", zap.String("path", workflowPath))
				processWorkflowFile(workflowPath)
			}
		}
		
		// Then check for workflow directories
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			
			workflowDir := filepath.Join(dir, entry.Name())
			logger.Debug("Checking workflow dir", zap.String("dir", workflowDir))
			
			// Check for workflow.yaml
			workflowYAMLPath := filepath.Join(workflowDir, StandardWorkflowYAML)
			if fs.Exists(workflowYAMLPath) {
				logger.Debug("Found workflow YAML file", zap.String("path", workflowYAMLPath))
				processWorkflowFile(workflowYAMLPath)
				continue
			}
			
			// Check for workflow.json as fallback
			workflowJSONPath := filepath.Join(workflowDir, "workflow.json")
			if fs.Exists(workflowJSONPath) {
				logger.Debug("Found workflow JSON file", zap.String("path", workflowJSONPath))
				processWorkflowFile(workflowJSONPath)
			}
		}
	}
	
	return discoveredWorkflows
}

// ReloadChangedWorkflows checks for modified workflow files and reloads them
// It returns a list of reloaded workflow names for informational purposes.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//
// Returns:
//   - A slice of reloaded workflow names
func (r *WorkflowRegistry) ReloadChangedWorkflows(fs io.FileSystem) []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	reloaded := []string{}
	
	// Check each cached workflow for modifications
	for name, path := range r.cache.sources {
		modified, err := r.isWorkflowModified(fs, name)
		if err != nil {
			// Log the error but continue with other workflows
			fmt.Printf("Error checking if workflow %s is modified: %v\n", name, err)
			continue
		}
		
		if modified {
			// Reload the workflow
			workflow, err := LoadWorkflowFromFile(fs, path)
			if err != nil {
				// Log the error but keep using the cached version
				fmt.Printf("Error reloading workflow %s: %v\n", name, err)
				continue
			}
			
			// Update the cache
			r.cache.workflows[name] = workflow
			r.cache.modified[name] = time.Now()
			reloaded = append(reloaded, name)
		}
	}
	
	return reloaded
}

// isWorkflowModified checks if a workflow file has been modified since it was last loaded
// This is used for cache invalidation to ensure we're using the latest workflow definition.
//
// Parameters:
//   - fs: FileSystem interface for file operations
//   - name: Name of the workflow to check
//
// Returns:
//   - true if the workflow has been modified, false otherwise
//   - error if there was a problem checking the modification time
func (r *WorkflowRegistry) isWorkflowModified(fs io.FileSystem, name string) (bool, error) {
	path, exists := r.cache.sources[name]
	if !exists {
		return false, fmt.Errorf("workflow %s not found in cache", name)
	}
	
	// Get file info to check modification time
	fileInfo, err := fs.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error getting file info for %s: %w", path, err)
	}
	
	// Get the last modified time from cache
	lastModified, exists := r.cache.modified[name]
	if !exists {
		// If we don't have a last modified time, consider it modified
		return true, nil
	}
	
	// Get the last modification time of the file
	fileModTime := fileInfo.ModTime()
	
	// Add a small buffer (1 second) to avoid false positives due to filesystem timestamp precision
	lastModified = lastModified.Add(-1 * time.Second)
	
	// Check if file has been modified since last loaded
	return fileModTime.After(lastModified), nil
}

// GetStandardWorkflowDirectories returns potential workflow locations
//
// Returns:
//   - Slice of directory paths where workflows might be found
func GetStandardWorkflowDirectories() []string {
	// Standard locations to look for workflows
	return []string{
		// Local development paths
		StandardTemplateDir,
		"internal/workflow/templates",
		"templates",
		
		// User directories
		filepath.Join(getUserHomeDir(), ".usm", "workflows"),
		
		// Project-specific workflows in .usm directory
		".usm/workflows",
	}
}

// getUserHomeDir returns the user's home directory
// It handles platform-specific differences and returns an empty string if unavailable
//
// Returns:
//   - The user's home directory path, or empty string if not available
func getUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Log the error but don't fail - features requiring home dir will be degraded
		fmt.Printf("Warning: Could not determine user home directory: %v\n", err)
		return ""
	}
	return home
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
		Description: "The default USM workflow for implementation",
		Steps:       StandardWorkflowSteps,
	}
}

// LoadWorkflowFromFile loads a workflow definition from the given file path
