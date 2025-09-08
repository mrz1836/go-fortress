package fortress_test

import (
	"testing"

	"github.com/mrz1836/go-fortress"
)

// BenchmarkGreet benchmarks the Greet function.
func BenchmarkGreet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fortress.Greet("BenchmarkUser")
	}
}
