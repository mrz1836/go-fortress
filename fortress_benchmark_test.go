package fortress_test

import (
	"testing"

	"github.com/mrz1836/go-fortress"
)

// BenchmarkFortify benchmarks the Fortify function.
func BenchmarkFortify(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fortress.Fortify("secret message")
	}
}

// BenchmarkFortify_EmptyString benchmarks Fortify with an empty string.
func BenchmarkFortify_EmptyString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fortress.Fortify("")
	}
}

// BenchmarkFortify_LongString benchmarks Fortify with a long string.
func BenchmarkFortify_LongString(b *testing.B) {
	longMsg := "This is a very long message that simulates a real-world use case with more content"
	for i := 0; i < b.N; i++ {
		_ = fortress.Fortify(longMsg)
	}
}

// BenchmarkGuard benchmarks the Guard function with safe input.
func BenchmarkGuard(b *testing.B) {
	forbidden := []string{"evil", "bad", "malicious"}
	for i := 0; i < b.N; i++ {
		_, _ = fortress.Guard("hello world", forbidden)
	}
}

// BenchmarkGuard_LongForbiddenList benchmarks Guard with a longer forbidden list.
func BenchmarkGuard_LongForbiddenList(b *testing.B) {
	forbidden := []string{
		"evil", "bad", "malicious", "dangerous", "harmful",
		"unsafe", "toxic", "virus", "malware", "threat",
	}
	for i := 0; i < b.N; i++ {
		_, _ = fortress.Guard("safe input here", forbidden)
	}
}

// BenchmarkGuard_BreachDetected benchmarks Guard when a breach is detected.
func BenchmarkGuard_BreachDetected(b *testing.B) {
	forbidden := []string{"evil", "bad", "malicious"}
	for i := 0; i < b.N; i++ {
		_, _ = fortress.Guard("this is evil", forbidden)
	}
}

// BenchmarkGuard_EmptyForbidden benchmarks Guard with an empty forbidden list.
func BenchmarkGuard_EmptyForbidden(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = fortress.Guard("anything goes", []string{})
	}
}
