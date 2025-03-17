package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command exists
	if rootCmd == nil {
		t.Error("Root command is nil")
	}

	// Test that the root command has the correct name
	if rootCmd.Use != "usm" {
		t.Errorf("Expected root command name to be 'usm', got '%s'", rootCmd.Use)
	}

	// Test that the debug flag is defined
	if rootCmd.PersistentFlags().Lookup("debug") == nil {
		t.Error("Debug flag is not defined")
	}
}