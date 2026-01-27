// Package main demonstrates low code coverage for CI testing.
package main

import "fmt"

func main() {
	fmt.Println("Coverage fail fixture")
	result := Add(1, 2)
	fmt.Printf("1 + 2 = %d\n", result)
}

// Add adds two integers.
func Add(a, b int) int {
	return a + b
}

// Subtract subtracts b from a.
func Subtract(a, b int) int {
	return a - b
}

// Multiply multiplies two integers.
func Multiply(a, b int) int {
	return a * b
}

// Divide divides a by b.
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

// Abs returns the absolute value of an integer.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// IsEven returns true if n is even.
func IsEven(n int) bool {
	return n%2 == 0
}

// Factorial calculates the factorial of n.
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// Fibonacci returns the nth Fibonacci number.
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}
