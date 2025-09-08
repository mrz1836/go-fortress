// Package fortress provides some simple Go functions to showcase the Go-Fortress CI/CD capabilities.
package fortress

import "fmt"

// Greet returns a greeting message for the given first name.
//
// This function performs the following steps:
// - Formats a greeting string using the provided first name.
//
// Parameters:
// - firstname: The first name to include in the greeting message.
//
// Returns:
// - A string containing the greeting message.
//
// Side Effects:
// - None.
//
// Notes:
// - Assumes firstname is a non-empty string; no validation is performed.
// - This function is standalone and not part of a larger workflow.
func Greet(firstname string) string {
	return fmt.Sprintf("Hello %s", firstname)
}
