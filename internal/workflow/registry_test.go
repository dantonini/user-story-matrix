// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

import (
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Setup and cleanup for all tests in this file
func TestMain(m *testing.M) {
	// Run tests
	result := m.Run()
	
	// Always reset the global registry after all tests complete
	ResetGlobalRegistry()
	
	os.Exit(result)
}

// Helper to reset registry before each test
func resetRegistryForTest() {
	ResetGlobalRegistry()
}

func TestNewWorkflowRegistry(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Create a new registry
	registry := NewWorkflowRegistry()

	// Check that it's not nil
	if registry == nil {
		t.Fatal("Expected registry to be non-nil")
	}

	// Check that the standard workflow is registered
	workflow, err := registry.GetWorkflow(StandardWorkflowName)
	if err != nil {
		t.Fatalf("Expected standard workflow to be registered, got error: %v", err)
	}

	// Check that the workflow has the correct properties
	if workflow.Name != StandardWorkflowName {
		t.Errorf("Expected workflow name to be %q, got %q", StandardWorkflowName, workflow.Name)
	}

	if workflow.Description != "Standard USM implementation workflow" {
		t.Errorf("Expected workflow description to be %q, got %q", "Standard USM implementation workflow", workflow.Description)
	}

	if len(workflow.Steps) != len(StandardWorkflowSteps) {
		t.Errorf("Expected workflow to have %d steps, got %d", len(StandardWorkflowSteps), len(workflow.Steps))
	}
}

func TestWorkflowRegistry_RegisterBuiltInWorkflow(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Create a new registry
	registry := NewWorkflowRegistry()

	// Create a test workflow
	testWorkflow := &WorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow",
		Steps:       []WorkflowStep{{ID: "test-step", Description: "Test step", Prompt: "Test prompt"}},
	}

	// Register the workflow
	registry.RegisterBuiltInWorkflow(testWorkflow)

	// Check that it was registered correctly
	workflow, err := registry.GetWorkflow("test-workflow")
	if err != nil {
		t.Fatalf("Expected test workflow to be registered, got error: %v", err)
	}

	if workflow != testWorkflow {
		t.Error("Expected retrieved workflow to be the same as the registered one")
	}
}

func TestWorkflowRegistry_GetWorkflow(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Create a new registry
	registry := NewWorkflowRegistry()

	// Test retrieving the standard workflow
	workflow, err := registry.GetWorkflow(StandardWorkflowName)
	if err != nil {
		t.Fatalf("Expected to find standard workflow, got error: %v", err)
	}

	if workflow.Name != StandardWorkflowName {
		t.Errorf("Expected workflow name to be %q, got %q", StandardWorkflowName, workflow.Name)
	}

	// Test retrieving a non-existent workflow
	_, err = registry.GetWorkflow("non-existent")
	if err == nil {
		t.Error("Expected error when retrieving non-existent workflow, got nil")
	}
}

func TestWorkflowRegistry_GetStandardWorkflow(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Create a new registry
	registry := NewWorkflowRegistry()

	// Get the standard workflow
	workflow := registry.GetStandardWorkflow()

	// Check that it's not nil
	if workflow == nil {
		t.Fatal("Expected standard workflow to be non-nil")
	}

	// Check that it has the correct name
	if workflow.Name != StandardWorkflowName {
		t.Errorf("Expected workflow name to be %q, got %q", StandardWorkflowName, workflow.Name)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Test that StandardWorkflowSteps are the same as those in the standard workflow
	registry := NewWorkflowRegistry()
	standardWorkflow := registry.GetStandardWorkflow()

	// Check if lengths match
	if len(StandardWorkflowSteps) != len(standardWorkflow.Steps) {
		t.Errorf("StandardWorkflowSteps has %d steps, but standard workflow has %d steps",
			len(StandardWorkflowSteps), len(standardWorkflow.Steps))
	}

	// Check each step's ID matches
	for i := 0; i < len(StandardWorkflowSteps) && i < len(standardWorkflow.Steps); i++ {
		if StandardWorkflowSteps[i].ID != standardWorkflow.Steps[i].ID {
			t.Errorf("Step %d: StandardWorkflowSteps ID %q doesn't match workflow step ID %q",
				i, StandardWorkflowSteps[i].ID, standardWorkflow.Steps[i].ID)
		}
	}

	// Verify that using the workflow manager with different constructors produces consistent results
	fs := &mockFileSystem{}
	io := &mockUserOutput{}

	// Create with default constructor (uses standard workflow)
	defaultWm := NewDefaultWorkflowManager(fs, io)

	// Create with explicit standard workflow
	explicitWm := NewWorkflowManager(fs, io, StandardWorkflowName, nil)

	// Check that they have the same number of steps
	if len(defaultWm.workflow.Steps) != len(explicitWm.workflow.Steps) {
		t.Errorf("Default workflow has %d steps, but explicit standard workflow has %d steps",
			len(defaultWm.workflow.Steps), len(explicitWm.workflow.Steps))
	}
}

// Mock implementations for testing
type mockFileSystem struct{}

func (m *mockFileSystem) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (m *mockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return nil
}

func (m *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (m *mockFileSystem) Exists(path string) bool {
	return false
}

func (m *mockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return nil, nil
}

func (m *mockFileSystem) Stat(path string) (os.FileInfo, error) {
	return nil, nil
}

func (m *mockFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return nil
}

type mockUserOutput struct{}

func (m *mockUserOutput) Print(message string) {}

func (m *mockUserOutput) PrintSuccess(message string) {}

func (m *mockUserOutput) PrintError(message string) {}

func (m *mockUserOutput) PrintWarning(message string) {}

func (m *mockUserOutput) PrintProgress(message string) {}

func (m *mockUserOutput) PrintStep(stepNumber int, totalSteps int, description string) {}

func (m *mockUserOutput) IsDebugEnabled() bool {
	return false
}

// TestWorkflowRegistrySharing validates that the global registry mechanism
// allows workflows registered with one WorkflowManager to be accessed by another.
// This test replaces the previous TestWorkflowRegistrySharingLimitation which only
// documented a limitation.
func TestWorkflowRegistrySharing(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Setup
	fs := &mockFileSystem{}
	io := &mockUserOutput{}

	// Create two managers
	manager1 := NewDefaultWorkflowManager(fs, io)
	manager2 := NewDefaultWorkflowManager(fs, io)

	// Create and register a custom workflow with manager1
	customWorkflow := &WorkflowDefinition{
		Name:        "custom-test-workflow",
		Description: "Test workflow for registry sharing",
		Steps:       []WorkflowStep{{ID: "test-step", Description: "Test step", Prompt: "Test prompt"}},
	}
	manager1.RegisterWorkflow(customWorkflow)

	// Now manager2 should have access to the workflow registered with manager1
	workflow, err := manager2.registry.GetWorkflow("custom-test-workflow")
	if err != nil {
		t.Errorf("Expected manager2 to have access to workflows registered with manager1, got error: %v", err)
	}

	// Verify the workflow properties
	if workflow == nil {
		t.Error("Retrieved workflow should not be nil")
	} else {
		if workflow.Name != "custom-test-workflow" {
			t.Errorf("Expected workflow name to be %q, got %q", "custom-test-workflow", workflow.Name)
		}
		if len(workflow.Steps) != 1 {
			t.Errorf("Expected workflow to have 1 step, got %d", len(workflow.Steps))
		}
	}
}

// TestGlobalRegistry_Singleton verifies that all calls to GetGlobalRegistry
// return the same instance of the registry
func TestGlobalRegistry_Singleton(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Get the registry twice
	registry1 := GetGlobalRegistry()
	registry2 := GetGlobalRegistry()

	// They should be the same instance
	if registry1 != registry2 {
		t.Error("Expected registry1 and registry2 to be the same instance")
	}

	// Register a workflow with registry1
	testWorkflow := &WorkflowDefinition{
		Name:        "singleton-test",
		Description: "Test singleton pattern",
		Steps:       []WorkflowStep{{ID: "test-step", Description: "Test step", Prompt: "Test prompt"}},
	}
	registry1.RegisterBuiltInWorkflow(testWorkflow)

	// It should be available through registry2
	workflow, err := registry2.GetWorkflow("singleton-test")
	if err != nil {
		t.Errorf("Expected workflow to be available through registry2, got error: %v", err)
	}
	if workflow == nil || workflow.Name != "singleton-test" {
		t.Error("Expected to retrieve the correct workflow through registry2")
	}
}

// TestResetGlobalRegistry verifies that the global registry can be reset
func TestResetGlobalRegistry(t *testing.T) {
	// Get the current registry
	registry1 := GetGlobalRegistry()
	
	// Register a workflow
	registry1.RegisterBuiltInWorkflow(&WorkflowDefinition{
		Name:        "reset-test",
		Description: "Test registry reset",
		Steps:       []WorkflowStep{},
	})
	
	// Verify the workflow is registered
	_, err := registry1.GetWorkflow("reset-test")
	if err != nil {
		t.Errorf("Expected workflow to be registered, got error: %v", err)
	}
	
	// Reset the registry
	registry2 := ResetGlobalRegistry()
	
	// Verify they're different instances
	if registry1 == registry2 {
		t.Error("Expected registry1 and registry2 to be different instances after reset")
	}
	
	// The workflow should no longer be available
	_, err = registry2.GetWorkflow("reset-test")
	if err == nil {
		t.Error("Expected workflow to be removed after reset")
	}
}

// TestWorkflowRegistry_ListWorkflows verifies that the registry can list all workflows
func TestWorkflowRegistry_ListWorkflows(t *testing.T) {
	// Reset for test isolation
	resetRegistryForTest()
	
	// Create a new registry
	registry := NewWorkflowRegistry()
	
	// Initially, only the standard workflow should be present
	workflows := registry.ListWorkflows()
	if len(workflows) != 1 {
		t.Errorf("Expected 1 workflow, got %d", len(workflows))
	}
	if len(workflows) > 0 && workflows[0] != StandardWorkflowName {
		t.Errorf("Expected standard workflow name to be %q, got %q", StandardWorkflowName, workflows[0])
	}
	
	// Register two more workflows
	registry.RegisterBuiltInWorkflow(&WorkflowDefinition{
		Name:        "test-workflow-1",
		Description: "Test workflow 1",
		Steps:       []WorkflowStep{},
	})
	registry.RegisterBuiltInWorkflow(&WorkflowDefinition{
		Name:        "test-workflow-2",
		Description: "Test workflow 2",
		Steps:       []WorkflowStep{},
	})
	
	// Get the list of workflows
	workflows = registry.ListWorkflows()
	
	// There should be 3 workflows now
	if len(workflows) != 3 {
		t.Errorf("Expected 3 workflows, got %d", len(workflows))
	}
	
	// Sort the workflows for consistent testing
	sort.Strings(workflows)
	
	// Check the workflow names
	expectedNames := []string{StandardWorkflowName, "test-workflow-1", "test-workflow-2"}
	sort.Strings(expectedNames)
	
	for i, name := range workflows {
		if i < len(expectedNames) && name != expectedNames[i] {
			t.Errorf("Expected workflow at index %d to be %q, got %q", i, expectedNames[i], name)
		}
	}
}

// TestWorkflowRegistry_GetStandardWorkflowPanic verifies the panic handling in GetStandardWorkflow
func TestWorkflowRegistry_GetStandardWorkflowPanic(t *testing.T) {
	// Create a registry that is intentionally corrupted by having an empty builtInWorkflows map
	// Note: This is only possible in testing since we're bypassing the normal constructor
	corruptedRegistry := &WorkflowRegistry{
		builtInWorkflows: make(map[string]*WorkflowDefinition),
		mutex:            sync.RWMutex{},
	}

	// The GetStandardWorkflow method should panic when the standard workflow is not found
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected GetStandardWorkflow to panic when standard workflow is not found")
		} else {
			// Verify the panic message
			panicMsg, ok := r.(string)
			if !ok {
				t.Errorf("Expected panic message to be a string, got %T", r)
			} else if !strings.Contains(panicMsg, "Standard workflow not found") {
				t.Errorf("Expected panic message to contain 'Standard workflow not found', got %q", panicMsg)
			}
		}
	}()

	// This should panic
	_ = corruptedRegistry.GetStandardWorkflow()
}

func TestWorkflowRegistry_LoadFromDirectory(t *testing.T) {
	// TODO: Test loading a workflow from a directory
	// 1. Create a mock filesystem with a workflow directory
	// 2. Set up workflow.yaml and prompt files
	// 3. Call LoadFromDirectory
	// 4. Verify the loaded workflow definition
}

func TestWorkflowRegistry_DiscoverWorkflows(t *testing.T) {
	// TODO: Test discovering workflows from standard locations
	// 1. Create a mock filesystem with workflow directories in standard locations
	// 2. Call DiscoverWorkflows
	// 3. Verify all expected workflows are discovered
}

func TestWorkflowRegistry_ReloadChangedWorkflows(t *testing.T) {
	// TODO: Test reloading workflows when they change on disk
	// 1. Create a mock filesystem with a workflow
	// 2. Load the workflow into the registry
	// 3. Modify the workflow files
	// 4. Call ReloadChangedWorkflows
	// 5. Verify the workflow was reloaded with the new content
}

func TestWorkflowRegistry_GetWorkflow_FileSystem(t *testing.T) {
	// TODO: Test retrieving workflows from both built-in and file-based sources
	// 1. Create a registry with built-in workflows
	// 2. Load a file-based workflow
	// 3. Test retrieving both types of workflows
	// 4. Test retrieving a non-existent workflow
}

func TestGetStandardWorkflowDirectories(t *testing.T) {
	// Get the standard workflow directories
	dirs := GetStandardWorkflowDirectories()
	
	// Verify that the result is not nil
	assert.NotNil(t, dirs, "GetStandardWorkflowDirectories should not return nil")
	
	// Verify that the result is not empty
	assert.NotEmpty(t, dirs, "GetStandardWorkflowDirectories should not return an empty slice")
	
	// Verify the returned directories include StandardTemplateDir
	assert.Contains(t, dirs, StandardTemplateDir, 
		"GetStandardWorkflowDirectories should include StandardTemplateDir")
}
