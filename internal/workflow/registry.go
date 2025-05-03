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
	return r.GetWorkflowWithFS(name, nil)
}

// GetWorkflowWithFS retrieves a workflow by name, using the provided filesystem
// for auto-discovery if needed. If fs is nil, it creates an OSFileSystem.
//
// Parameters:
//   - name: The unique identifier of the workflow to retrieve
//   - fs: Optional filesystem to use for auto-discovery (can be nil)
//
// Returns:
//   - The requested WorkflowDefinition, or an error if it doesn't exist
func (r *WorkflowRegistry) GetWorkflowWithFS(name string, fs io.FileSystem) (*WorkflowDefinition, error) {
	// First try with a read lock
	r.mutex.RLock()
	
	// Check built-in workflows
	workflow, exists := r.builtInWorkflows[name]
	if exists {
		r.mutex.RUnlock()
		return workflow, nil
	}
	
	// Check cached workflows
	workflow, exists = r.cache.workflows[name]
	if exists {
		r.mutex.RUnlock()
		return workflow, nil
	}
	
	// Release read lock before trying discovery
	r.mutex.RUnlock()
	
	// Try discovery if not found in cache or built-in
	logger.Debug("Workflow not found in registry, attempting discovery", 
		zap.String("name", name))
	
	// Create a filesystem if one wasn't provided
	if fs == nil {
		fs = io.NewOSFileSystem()
	}
	
	// Get standard workflow directories
	directories := GetStandardWorkflowDirectories()
	
	// Check for the specific workflow in these directories
	workflowFound := false
	var discoveredWorkflow *WorkflowDefinition
	
	// Check for the specific workflow in each standard directory
	for _, dir := range directories {
		if !fs.Exists(dir) {
			continue
		}
		
		// Check if the directory contains a workflow with this name
		workflowDir := filepath.Join(dir, name)
		workflowYAMLPath := filepath.Join(workflowDir, "workflow.yaml")
		
		if fs.Exists(workflowYAMLPath) {
			// Found the workflow, try to load it
			loadedWorkflow, _, err := LoadWorkflowFromDirectory(fs, workflowDir)
			if err == nil {
				// Successfully loaded the workflow
				workflowFound = true
				discoveredWorkflow = loadedWorkflow
				break
			}
		}
	}
	
	// If we found the workflow, add it to the registry cache
	if workflowFound && discoveredWorkflow != nil {
		// Now get a write lock to update the cache
		r.mutex.Lock()
		r.cache.workflows[name] = discoveredWorkflow
		r.cache.sources[name] = filepath.Join(".usm/workflows", name)
		r.cache.modified[name] = time.Now()
		r.mutex.Unlock()
		
		return discoveredWorkflow, nil
	}
	
	// Still not found after targeted discovery
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
		
		// Update steps with prompt content from files using improved path resolution
		for i := range workflow.Steps {
			step := &workflow.Steps[i]
			
			// The prompt field in the file should contain a path to the prompt file
			// Check if it's a file path or an embedded prompt
			if strings.HasPrefix(step.Prompt, "prompts/") || filepath.Ext(step.Prompt) == ".md" {
				promptPath := step.Prompt
				resolvedPath := ""
				
				// Try multiple possible locations for the prompt file
				possiblePaths := []string{
					// Original path as specified
					promptPath,
					// Absolute path relative to workflow directory
					filepath.Join(path, promptPath),
				}
				
				// If the path doesn't include prompts/ prefix, add it as another possibility
				if !strings.HasPrefix(promptPath, "prompts/") && !strings.HasPrefix(promptPath, "prompts\\") {
					possiblePaths = append(possiblePaths, 
						filepath.Join(path, "prompts", filepath.Base(promptPath)))
				}
				
				// Check each possible path
				for _, possPath := range possiblePaths {
					if fs.Exists(possPath) {
						resolvedPath = possPath
						break
					}
				}
				
				// If we found the file, read it and update the step
				if resolvedPath != "" {
					// Read prompt content
					promptData, err := fs.ReadFile(resolvedPath)
					if err != nil {
						return nil, fmt.Errorf("failed to read prompt file %s: %w", resolvedPath, err)
					}
					
					// Set prompt content and mark as file-sourced
					step.Prompt = string(promptData)
					step.source = promptSource{
						sourceType: promptSourceFile,
						filePath:   resolvedPath,
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
	
	// Update steps with prompt content from files using improved path resolution
	for i := range workflow.Steps {
		step := &workflow.Steps[i]
		
		// The prompt field in the file should contain a path to the prompt file
		// Check if it's a file path or an embedded prompt
		if strings.HasPrefix(step.Prompt, "prompts/") || filepath.Ext(step.Prompt) == ".md" {
			promptPath := step.Prompt
			resolvedPath := ""
			
			// Try multiple possible locations for the prompt file
			possiblePaths := []string{
				// Original path as specified
				promptPath,
				// Absolute path relative to workflow directory
				filepath.Join(path, promptPath),
			}
			
			// If the path doesn't include prompts/ prefix, add it as another possibility
			if !strings.HasPrefix(promptPath, "prompts/") && !strings.HasPrefix(promptPath, "prompts\\") {
				possiblePaths = append(possiblePaths, 
					filepath.Join(path, "prompts", filepath.Base(promptPath)))
			}
			
			// Check each possible path
			for _, possPath := range possiblePaths {
				if fs.Exists(possPath) {
					resolvedPath = possPath
					break
				}
			}
			
			// If we found the file, read it and update the step
			if resolvedPath != "" {
				// Read prompt content
				promptData, err := fs.ReadFile(resolvedPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read prompt file %s: %w", resolvedPath, err)
				}
				
				// Set prompt content and mark as file-sourced
				step.Prompt = string(promptData)
				step.source = promptSource{
					sourceType: promptSourceFile,
					filePath:   resolvedPath,
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
	
	// Check if each directory exists
	for _, dir := range directories {
		exists := fs.Exists(dir)
		logger.Debug("Directory existence check", 
			zap.String("dir", dir), 
			zap.Bool("exists", exists))
	}
	
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
			
			// For registry tests with mock file system, we need this special handling
			// When using the mock file system in tests, sometimes directory entries aren't
			// properly reported by ReadDir even if the directory exists
			if fs.Exists(workflowDir) && !containsDirectoryEntry(entries, entry.Name()) {
				// This is a test-specific workaround
				logger.Debug("Directory exists but wasn't reported in entries, checking for workflow file", 
					zap.String("dir", workflowDir))
				
				// Check for workflow.yaml
				workflowYAMLPath := filepath.Join(workflowDir, StandardWorkflowYAML)
				if fs.Exists(workflowYAMLPath) {
					// Load using LoadWorkflowFromDirectory for better prompt resolution
					workflow, info, err := LoadWorkflowFromDirectory(fs, workflowDir)
					if err != nil {
						logger.Error("Error loading workflow from directory", 
							zap.String("dir", workflowDir), zap.Error(err))
						continue
					}
					
					// Add to cache with proper source tracking
					r.cache.workflows[workflow.Name] = workflow
					r.cache.sources[workflow.Name] = workflowDir
					r.cache.modified[workflow.Name] = time.Now()
					discoveredWorkflows[workflow.Name] = workflow
					
					logger.Debug("Successfully loaded workflow from directory", 
						zap.String("name", workflow.Name), 
						zap.String("dir", workflowDir),
						zap.String("source", info.Source))
				}
			}
			
			// Regular flow - Check for workflow.yaml
			workflowYAMLPath := filepath.Join(workflowDir, StandardWorkflowYAML)
			workflowYAMLExists := fs.Exists(workflowYAMLPath)
			logger.Debug("Workflow YAML existence check", 
				zap.String("path", workflowYAMLPath), 
				zap.Bool("exists", workflowYAMLExists))
			
			if workflowYAMLExists {
				logger.Debug("Found workflow YAML file", zap.String("path", workflowYAMLPath))
				// Load using LoadWorkflowFromDirectory for better prompt resolution
				workflow, info, err := LoadWorkflowFromDirectory(fs, workflowDir)
				if err != nil {
					logger.Error("Error loading workflow from directory", 
						zap.String("dir", workflowDir), zap.Error(err))
					continue
				}
				
				// Add to cache with proper source tracking
				r.cache.workflows[workflow.Name] = workflow
				r.cache.sources[workflow.Name] = workflowDir
				r.cache.modified[workflow.Name] = time.Now()
				discoveredWorkflows[workflow.Name] = workflow
				
				logger.Debug("Successfully loaded workflow from directory", 
					zap.String("name", workflow.Name), 
					zap.String("dir", workflowDir),
					zap.String("source", info.Source))
				
				continue
			}
			
			// Check for workflow.json as fallback
			workflowJSONPath := filepath.Join(workflowDir, "workflow.json")
			workflowJSONExists := fs.Exists(workflowJSONPath)
			logger.Debug("Workflow JSON existence check", 
				zap.String("path", workflowJSONPath), 
				zap.Bool("exists", workflowJSONExists))
			
			if workflowJSONExists {
				logger.Debug("Found workflow JSON file", zap.String("path", workflowJSONPath))
				processWorkflowFile(workflowJSONPath)
			}
		}
	}
	
	// Add built-in workflows to the discovered workflows
	for name, workflow := range r.builtInWorkflows {
		discoveredWorkflows[name] = workflow
	}
	
	// Extract workflow names for logging
	workflowNames := make([]string, 0, len(discoveredWorkflows))
	for name := range discoveredWorkflows {
		workflowNames = append(workflowNames, name)
	}
	
	logger.Debug("Discovered workflows", 
		zap.Int("count", len(discoveredWorkflows)),
		zap.Strings("names", workflowNames))
	
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
	dirs := []string{
		// Local development paths
		StandardTemplateDir,
		"internal/workflow/templates",
		"templates",
		
		// User directories
		filepath.Join(getUserHomeDir(), ".usm", "workflows"),
		
		// Project-specific workflows in .usm directory
		".usm/workflows",
	}
	
	logger.Debug("Standard workflow directories", zap.Strings("directories", dirs))
	
	return dirs
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

// GetWorkflowSourceInfo returns the source and path information for a workflow.
// This helps determine where a workflow came from (built-in, user, project)
// and its filesystem path (if applicable).
//
// Parameters:
//   - name: The name of the workflow to get source info for
//
// Returns:
//   - source: The source of the workflow (built-in, user, project)
//   - path: The filesystem path of the workflow (if available)
func (r *WorkflowRegistry) GetWorkflowSourceInfo(name string) (string, string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	// Check if it's a built-in workflow
	if _, exists := r.builtInWorkflows[name]; exists {
		return SourceBuiltIn, "-"
	}
	
	// Check if it's in the cache and has source info
	if sourcePath, exists := r.cache.sources[name]; exists {
		// If we have a source path, determine the source type
		if sourcePath != "" {
			if strings.Contains(sourcePath, ".usm/workflows") || 
			   strings.Contains(sourcePath, ".usm\\workflows") {
				// Extract the base directory to keep path consistent
				sourceDir := filepath.Dir(sourcePath)
				return SourceProject, sourceDir
			} else if homeDir := getUserHomeDir(); homeDir != "" && 
				(strings.Contains(sourcePath, filepath.Join(homeDir, ".usm/workflows")) || 
				 strings.Contains(sourcePath, filepath.Join(homeDir, ".usm\\workflows"))) {
				return SourceUser, sourcePath
			}
			
			// If we have a path but can't determine the source type,
			// just return the path with unknown source
			return "unknown", sourcePath
		}
	}
	
	// If we don't have source info, try to determine it from the name
	if strings.HasPrefix(name, "project-") {
		return SourceProject, "-"
	} else if strings.HasPrefix(name, "user-") {
		return SourceUser, "-"
	}
	
	// Default values if we can't determine the source
	return "unknown", "-"
}

// AddToCache adds a workflow to the registry's cache.
// This is useful for testing scenarios where we want to simulate
// a workflow being loaded from a file, but don't want to go through
// the full discovery process.
//
// Parameters:
//   - workflow: The WorkflowDefinition to add to the cache
//   - sourcePath: The path to associate with this workflow
func (r *WorkflowRegistry) AddToCache(workflow *WorkflowDefinition, sourcePath string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.cache.workflows[workflow.Name] = workflow
	r.cache.sources[workflow.Name] = sourcePath
	r.cache.modified[workflow.Name] = time.Now()
}

// ClearBuiltInWorkflows removes all built-in workflows from the registry.
// This is primarily used for testing scenarios where a completely empty
// registry is needed to test behavior with no workflows present.
//
// IMPORTANT: This method should only be used in tests and should never be
// called in production code, as it will make the standard workflow unavailable.
func (r *WorkflowRegistry) ClearBuiltInWorkflows() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Clear all built-in workflows
	r.builtInWorkflows = make(map[string]*WorkflowDefinition)
}

// containsDirectoryEntry checks if a directory name is found in a list of directory entries
// This is used to handle mock filesystem inconsistencies in tests
func containsDirectoryEntry(entries []os.DirEntry, name string) bool {
	for _, entry := range entries {
		if entry.Name() == name && entry.IsDir() {
			return true
		}
	}
	return false
}
