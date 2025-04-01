// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.


//go:build tools
// +build tools

package tools

import (
	// Import golangci-lint for dependency management
	// This ensures go.mod tracks the version but doesn't include it in the binary
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)

// This file is used to track tool dependencies.
// It's not included in the build but helps maintain consistent tooling versions.
// To install tools: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2