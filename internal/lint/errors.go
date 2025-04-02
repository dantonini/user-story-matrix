// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package lint

import (
	"errors"
)

// Static error variables for the lint package
var (
	ErrLintNotInstalled = errors.New("golangci-lint is not installed")
	ErrRootNotFound    = errors.New("could not find project root")
) 