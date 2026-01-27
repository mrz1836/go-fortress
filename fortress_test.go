package fortress_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress"
)

// TestFortify tests the Fortify function with various input scenarios using table-driven tests.
func TestFortify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal message",
			input:    "secret data",
			expected: "🏰 secret data 🏰",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "🏰  🏰",
		},
		{
			name:     "whitespace",
			input:    " ",
			expected: "🏰   🏰",
		},
		{
			name:     "special characters",
			input:    "@!$%^&*()",
			expected: "🏰 @!$%^&*() 🏰",
		},
		{
			name:     "unicode characters",
			input:    "日本語",     //nolint:gosmopolitan // Test with non-ASCII characters
			expected: "🏰 日本語 🏰", //nolint:gosmopolitan // Test with non-ASCII characters
		},
		{
			name:     "multiline message",
			input:    "line1\nline2",
			expected: "🏰 line1\nline2 🏰",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fortress.Fortify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGuard tests the Guard function with various input scenarios using table-driven tests.
func TestGuard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		forbidden   []string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "safe input",
			input:       "hello world",
			forbidden:   []string{"evil", "bad"},
			expected:    "hello world",
			expectError: false,
		},
		{
			name:        "forbidden word detected",
			input:       "this is evil",
			forbidden:   []string{"evil", "bad"},
			expected:    "",
			expectError: true,
			errorMsg:    "🚫 breach detected: contains \"evil\"",
		},
		{
			name:        "empty input",
			input:       "",
			forbidden:   []string{"evil"},
			expected:    "",
			expectError: false,
		},
		{
			name:        "empty forbidden list",
			input:       "anything goes",
			forbidden:   []string{},
			expected:    "anything goes",
			expectError: false,
		},
		{
			name:        "nil forbidden list",
			input:       "anything goes",
			forbidden:   nil,
			expected:    "anything goes",
			expectError: false,
		},
		{
			name:        "partial match",
			input:       "basketball",
			forbidden:   []string{"bad"},
			expected:    "basketball",
			expectError: false,
		},
		{
			name:        "substring match",
			input:       "embed bad word",
			forbidden:   []string{"bad"},
			expected:    "",
			expectError: true,
			errorMsg:    "🚫 breach detected: contains \"bad\"",
		},
		{
			name:        "case sensitive - no match",
			input:       "EVIL in caps",
			forbidden:   []string{"evil"},
			expected:    "EVIL in caps",
			expectError: false,
		},
		{
			name:        "first forbidden match wins",
			input:       "bad and evil",
			forbidden:   []string{"bad", "evil"},
			expected:    "",
			expectError: true,
			errorMsg:    "🚫 breach detected: contains \"bad\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fortress.Guard(tt.input, tt.forbidden)
			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
