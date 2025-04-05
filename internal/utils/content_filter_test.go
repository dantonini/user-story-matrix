// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterContentHash(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		includeContentHash bool
		expected          string
	}{
		{
			name: "Filter content hash",
			content: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
_content_hash: abcdef1234567890
---

# Test Content
This is a test.`,
			includeContentHash: false,
			expected: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
---

# Test Content
This is a test.`,
		},
		{
			name: "Include content hash when flag is set",
			content: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
_content_hash: abcdef1234567890
---

# Test Content
This is a test.`,
			includeContentHash: true,
			expected: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
_content_hash: abcdef1234567890
---

# Test Content
This is a test.`,
		},
		{
			name: "No content hash to filter",
			content: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
---

# Test Content
This is a test.`,
			includeContentHash: false,
			expected: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
---

# Test Content
This is a test.`,
		},
		{
			name:              "Empty content",
			content:           "",
			includeContentHash: false,
			expected:          "",
		},
		{
			name: "Multiple metadata sections",
			content: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
_content_hash: abcdef1234567890
---

# Test Content
This is a test.

---
Another section:
_content_hash: should not be filtered
---`,
			includeContentHash: false,
			expected: `---
file_path: path/to/file.md
created_at: 2025-04-05T15:19:46+02:00
---

# Test Content
This is a test.

---
Another section:
_content_hash: should not be filtered
---`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterContentHash(tt.content, tt.includeContentHash)
			assert.Equal(t, tt.expected, result)
		})
	}
} 