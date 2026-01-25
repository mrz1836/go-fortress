package main

import "testing"

// Only test Add and Subtract - leaving many functions untested
// This results in approximately 20-30% coverage, well below typical thresholds.

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 3},
		{0, 0, 0},
		{-1, 1, 0},
	}

	for _, tt := range tests {
		got := Add(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSubtract(t *testing.T) {
	got := Subtract(5, 3)
	want := 2
	if got != want {
		t.Errorf("Subtract(5, 3) = %d; want %d", got, want)
	}
}

// Note: The following functions are NOT tested, causing low coverage:
// - Multiply
// - Divide
// - Max
// - Min
// - Abs
// - IsEven
// - Factorial
// - Fibonacci
