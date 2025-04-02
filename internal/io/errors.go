// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package io

import (
	"errors"
)

// Static error variables for the io package
var (
	ErrUnexpectedModel = errors.New("unexpected model type")
	ErrSelectionCanceled = errors.New("selection canceled")
	ErrTypeCast        = errors.New("could not cast value")
) 