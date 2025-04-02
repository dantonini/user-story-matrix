// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package changerequest

import (
	"errors"
)

// Static error variables for the changerequest package
var (
	ErrDirectoryNotFound = errors.New("change requests directory not found")
	ErrReadDirectory     = errors.New("failed to read directory")
) 