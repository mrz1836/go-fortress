// Package main provides functions for testing failure scenarios.
package main

import "fmt"

func main() {
	fmt.Println("Test fail fixture")
}

// Add adds two numbers and returns the result.
func Add(a, b int) int {
	return a + b
}

// Divide divides a by b. Panics if b is zero.
func Divide(a, b int) int {
	return a / b
}

// GetValue returns a value from a pointer. May panic.
func GetValue(p *int) int {
	return *p
}

// SlowOperation simulates a slow operation.
func SlowOperation() {
	// This function is intentionally slow
	select {}
}
