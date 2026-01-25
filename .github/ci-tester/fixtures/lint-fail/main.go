// Package main demonstrates lint failures for CI testing.
package main

import "fmt"

func main() {
	// This variable is declared but not used - triggers unused variable error
	unusedVar := "this triggers lint failure"

	// This variable is also unused
	anotherUnused := 42

	// Only this gets printed
	fmt.Println("Hello, World!")
}

// badlyFormatted function has formatting issues (no space after func name for gofmt)
func badlyFormatted()string{return"bad"}

// UnusedFunction is never called - triggers staticcheck
func UnusedFunction() {
	fmt.Println("I am never called")
}
