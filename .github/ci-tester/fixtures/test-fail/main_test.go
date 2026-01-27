package main

import (
	"testing"
	"time"
)

func TestAdd_Failure(t *testing.T) {
	// This test intentionally fails
	result := Add(2, 2)
	if result != 5 { // Wrong expectation - should be 4
		t.Errorf("Add(2, 2) = %d; want 5", result)
	}
}

func TestAdd_AssertionFailure(t *testing.T) {
	// Another failing test with clear assertion
	got := Add(1, 1)
	want := 3 // Wrong expectation
	if got != want {
		t.Fatalf("assertion failed: Add(1, 1) = %d, want %d", got, want)
	}
}

func TestDivide_Panic(t *testing.T) {
	// This test triggers a panic - divide by zero
	_ = Divide(10, 0)
}

func TestGetValue_NilPointer(t *testing.T) {
	// This test triggers nil pointer dereference
	var p *int = nil
	_ = GetValue(p)
}

func TestSlowOperation_Timeout(t *testing.T) {
	// This test will timeout
	done := make(chan bool)
	go func() {
		SlowOperation()
		done <- true
	}()

	select {
	case <-done:
		// Should never reach here
	case <-time.After(100 * time.Millisecond):
		t.Fatal("test timed out waiting for SlowOperation")
	}
}
