// Package fortress provides some simple Go functions to showcase the Go-Fortress CI/CD capabilities.
package fortress

import (
	"errors"
	"fmt"
	"strings"
)

// ErrBreachDetected is returned when Guard detects forbidden content in the input.
var ErrBreachDetected = errors.New("🚫 breach detected")

// Fortify wraps a message with fortress protection markers.
//
// This function performs the following steps:
// - Formats a protected string using the provided message.
//
// Parameters:
// - message: The message to fortify.
//
// Returns:
// - A string containing the fortified message with fortress markers.
//
// Side Effects:
// - None.
//
// Notes:
// - Accepts any string input including empty strings.
// - This function is standalone and not part of a larger workflow.
func Fortify(message string) string {
	// INTENTIONAL ERROR FOR TESTING - REVERT ME
	var unusedVariable string  // this is declared but never used - Go compiler error
	return fmt.Sprintf("🏰 %s 🏰", message)
}

// Guard validates input against a list of forbidden values.
//
// This function performs the following steps:
// - Checks if the input contains any forbidden substrings.
// - Returns the original input if safe, or an error if breached.
//
// Parameters:
// - input: The string to validate.
// - forbidden: A slice of forbidden substrings to check against.
//
// Returns:
// - The original input string if no forbidden values are found.
// - An error if any forbidden value is detected in the input.
//
// Side Effects:
// - None.
//
// Notes:
// - An empty forbidden list will always return the input as safe.
// - Matching is case-sensitive and checks for substring containment.
func Guard(input string, forbidden []string) (string, error) {
	for _, f := range forbidden {
		if strings.Contains(input, f) {
			return "", fmt.Errorf("%w: contains %q", ErrBreachDetected, f)
		}
	}
	return input, nil
}
