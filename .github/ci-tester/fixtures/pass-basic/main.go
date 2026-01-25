// Package main provides a well-written Go program for CI success testing.
package main

import "fmt"

func main() {
	fmt.Println("Pass basic fixture - all checks should pass")

	result := Add(2, 3)
	fmt.Printf("2 + 3 = %d\n", result)
}

// Add adds two integers and returns the result.
func Add(a, b int) int {
	return a + b
}

// Subtract subtracts b from a and returns the result.
func Subtract(a, b int) int {
	return a - b
}

// Multiply multiplies two integers and returns the result.
func Multiply(a, b int) int {
	return a * b
}

// Divide divides a by b and returns the result and any error.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
