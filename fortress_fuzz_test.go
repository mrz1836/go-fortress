package fortress_test

import (
	"strings"
	"testing"

	"github.com/mrz1836/go-fortress"
)

// FuzzFortify fuzzes the Fortify function with random inputs.
func FuzzFortify(f *testing.F) {
	// Add seed corpus
	f.Add("hello")
	f.Add("")
	f.Add(" ")
	f.Add("special chars: @#$%^&*()")
	f.Add("日本語") //nolint: gosmopolitan // Test with non-ASCII characters
	f.Add("line1\nline2")
	f.Add("🏰 already has emoji 🏰")

	f.Fuzz(func(t *testing.T, message string) {
		result := fortress.Fortify(message)

		// Result should always contain the original message
		if !strings.Contains(result, message) {
			t.Errorf("result %q does not contain input %q", result, message)
		}

		// Result should have fortress markers
		if !strings.HasPrefix(result, "🏰 ") {
			t.Errorf("result %q does not start with fortress marker", result)
		}
		if !strings.HasSuffix(result, " 🏰") {
			t.Errorf("result %q does not end with fortress marker", result)
		}
	})
}

// FuzzGuard fuzzes the Guard function with random inputs.
func FuzzGuard(f *testing.F) {
	// Add seed corpus
	f.Add("safe input")
	f.Add("")
	f.Add(" ")
	f.Add("contains bad word")
	f.Add("this is evil")
	f.Add("EVIL in caps")
	f.Add("special @#$%")

	f.Fuzz(func(t *testing.T, input string) {
		forbidden := []string{"bad", "evil"}
		result, err := fortress.Guard(input, forbidden)

		// Check consistency of result
		if err == nil {
			// If no error, result should equal input
			if result != input {
				t.Errorf("expected %q, got %q", input, result)
			}
			// Input should not contain any forbidden values
			for _, f := range forbidden {
				if strings.Contains(input, f) {
					t.Errorf("input %q contains forbidden value %q but no error was returned", input, f)
				}
			}
		} else {
			// If error, result should be empty
			if result != "" {
				t.Errorf("expected empty result on error, got %q", result)
			}
			// Input should contain at least one forbidden value
			containsForbidden := false
			for _, f := range forbidden {
				if strings.Contains(input, f) {
					containsForbidden = true
					break
				}
			}
			if !containsForbidden {
				t.Errorf("error returned but input %q does not contain any forbidden values", input)
			}
		}
	})
}

// FuzzGuardEmptyForbidden fuzzes Guard with an empty forbidden list.
func FuzzGuardEmptyForbidden(f *testing.F) {
	f.Add("anything")
	f.Add("")
	f.Add("evil bad words")

	f.Fuzz(func(t *testing.T, input string) {
		result, err := fortress.Guard(input, []string{})
		// With empty forbidden list, should always succeed
		if err != nil {
			t.Errorf("unexpected error with empty forbidden list: %v", err)
		}
		if result != input {
			t.Errorf("expected %q, got %q", input, result)
		}
	})
}
