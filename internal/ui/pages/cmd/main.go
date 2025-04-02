// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This is a demo application to showcase the selection page.
// It is not part of the main application.

package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user-story-matrix/usm/internal/models"
	"github.com/user-story-matrix/usm/internal/ui/pages"
)

func main() {
	// Create sample user stories
	stories := []models.UserStory{
		{
			Title:         "Add login functionality",
			FilePath:      "docs/user-stories/auth/01-add-login-functionality.md",
			Description:   "Users should be able to log in with their credentials",
			IsImplemented: false,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
		{
			Title:         "Integrate payment provider",
			FilePath:      "docs/user-stories/payment/01-integrate-payment-provider.md",
			Description:   "Users should be able to pay for services",
			IsImplemented: false,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
		{
			Title:         "Export user data to CSV",
			FilePath:      "docs/user-stories/export/01-export-user-data-to-csv.md",
			Description:   "Users should be able to export their data",
			IsImplemented: true,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
		{
			Title:         "User profile page",
			FilePath:      "docs/user-stories/profile/01-user-profile-page.md",
			Description:   "Users should be able to view and edit their profile",
			IsImplemented: false,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
		{
			Title:         "Password reset",
			FilePath:      "docs/user-stories/auth/02-password-reset.md",
			Description:   "Users should be able to reset their password",
			IsImplemented: false,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		},
	}

	// Create selection page
	p := pages.New(stories, false)

	// Run the program
	program := tea.NewProgram(p, tea.WithAltScreen())
	model, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get selected stories
	selectionPage, ok := model.(*pages.SelectionPage)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: Could not convert model to SelectionPage\n")
		os.Exit(1)
	}
	selected := selectionPage.GetSelected()
	
	// Print selected stories
	fmt.Println("\nSelected stories:")
	for _, idx := range selected {
		fmt.Printf("- %s\n", stories[idx].Title)
	}
} 