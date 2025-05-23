// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package metadata

import (
	"testing"
)

func TestNewDefaultMetadataOptions(t *testing.T) {
	options := NewDefaultMetadataOptions()
	
	// Test that default values are set correctly
	if options.SkipReferences != false {
		t.Errorf("NewDefaultMetadataOptions() SkipReferences = %v, want %v", options.SkipReferences, false)
	}
	
	if options.Debug != false {
		t.Errorf("NewDefaultMetadataOptions() Debug = %v, want %v", options.Debug, false)
	}
} 