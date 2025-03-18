package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAskCommandStructure(t *testing.T) {
	// Check that the ask command is properly configured
	assert.Equal(t, "ask", askCmd.Use)
	assert.NotEmpty(t, askCmd.Short)
	assert.NotEmpty(t, askCmd.Long)
	
	// Check that it has the feature subcommand
	found := false
	for _, subCmd := range askCmd.Commands() {
		if subCmd.Use == "feature" {
			found = true
			break
		}
	}
	assert.True(t, found, "Feature subcommand should be added to ask command")
}

func TestAskFeatureCommandConfig(t *testing.T) {
	// Check that the feature command is properly configured
	assert.Equal(t, "feature", askFeatureCmd.Use)
	assert.NotEmpty(t, askFeatureCmd.Short)
	assert.NotEmpty(t, askFeatureCmd.Long)
	assert.NotNil(t, askFeatureCmd.Run, "Run function should be defined")
}

func TestAskCommandRegistration(t *testing.T) {
	// Test that the ask command is registered with the root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "ask" {
			found = true
			break
		}
	}
	assert.True(t, found, "Ask command should be registered with the root command")
}

func TestExecuteAskCommand(t *testing.T) {
	// Create a test cobra command
	testCmd := &cobra.Command{
		Use: "test",
	}
	testCmd.AddCommand(askCmd)
	
	// This is a limited test as we can't easily test the interactive form
	// We're just making sure the command structure is correct
	assert.NotPanics(t, func() {
		testCmd.SetArgs([]string{"ask", "--help"})
		testCmd.Execute()
	})
} 