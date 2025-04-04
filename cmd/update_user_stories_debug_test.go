// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	fsio "github.com/user-story-matrix/usm/internal/io"
	"github.com/user-story-matrix/usm/internal/logger"
)

func TestDebugFlag(t *testing.T) {
	// Save the original stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr
	
	// Create pipes for capturing output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	
	// Redirect stdout and stderr
	os.Stdout = wOut
	os.Stderr = wErr
	
	// Ensure we restore the original stdout and stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		logger.SetDebugMode(false) // Reset debug mode
	}()
	
	// Create a test command with debug flag set to true
	cmd := &cobra.Command{}
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().Bool("skip-references", false, "")
	cmd.Flags().String("test-root", "", "")
	
	_ = cmd.Flags().Set("debug", "true")
	
	// Set up a test root for the command
	tempDir := t.TempDir()
	_ = cmd.Flags().Set("test-root", tempDir)
	
	// Set up a test filesystem
	fs := fsio.NewMockFileSystem()
	
	// Create a mock user stories directory
	userStoriesDir := "docs/user-stories"
	fs.AddDirectory(userStoriesDir)
	
	// Add a test user story file
	fs.AddFile("docs/user-stories/test-story.md", []byte("# Test User Story\n\nThis is a test user story.\n"))
	
	// Create a new command with the run function
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			if debug {
				logger.SetDebugMode(true)
				logger.Debug("Debug mode is enabled") // This message should appear in the output
			}
		},
	}
	
	// Add the same flags as the update command
	testCmd.Flags().Bool("debug", false, "")
	testCmd.Flags().Bool("skip-references", false, "")
	testCmd.Flags().String("test-root", "", "")
	
	// Set the debug flag
	_ = testCmd.Flags().Set("debug", "true")
	
	// Run the command
	testCmd.Run(testCmd, []string{})
	
	// Close the write end of the pipes
	wOut.Close()
	wErr.Close()
	
	// Read the output
	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)
	
	// Check that the output contains debug messages
	output := bufOut.String() + bufErr.String()
	assert.Contains(t, output, "Debug mode is enabled", "Debug output should be visible when debug flag is set")
}

func TestNoDebugFlag(t *testing.T) {
	// Similar to TestDebugFlag, but without setting the debug flag
	// Save the original stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr
	
	// Create pipes for capturing output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	
	// Redirect stdout and stderr
	os.Stdout = wOut
	os.Stderr = wErr
	
	// Ensure we restore the original stdout and stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		logger.SetDebugMode(false) // Reset debug mode
	}()
	
	// Create a new command without the debug flag
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			if debug {
				logger.SetDebugMode(true)
				logger.Debug("Debug mode is enabled") // This should not appear
			} else {
				logger.Debug("This debug message should not appear") // This should not appear
			}
		},
	}
	
	// Add the same flags as the update command
	testCmd.Flags().Bool("debug", false, "")
	testCmd.Flags().Bool("skip-references", false, "")
	
	// Run the command (without setting the debug flag)
	testCmd.Run(testCmd, []string{})
	
	// Close the write end of the pipes
	wOut.Close()
	wErr.Close()
	
	// Read the output
	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)
	
	// Check that the output does not contain debug messages
	output := bufOut.String() + bufErr.String()
	assert.NotContains(t, output, "Debug mode is enabled", "Debug output should not be visible when debug flag is not set")
	assert.NotContains(t, output, "This debug message should not appear", "Debug output should not be visible when debug flag is not set")
} 