package main

import (
	"testing"
)

func TestCounter_Race(t *testing.T) {
	c := &Counter{}
	// This will trigger a data race when run with -race flag
	ConcurrentCounterIncrement(c, 1000)

	// Value will be unpredictable due to race
	t.Logf("Final counter value: %d (expected ~10000)", c.Value())
}

func TestMap_ConcurrentAccess(t *testing.T) {
	// This will trigger "concurrent map read and map write" panic
	// or data race when run with -race flag
	ConcurrentMapAccess()
}

func TestMap_DirectRace(t *testing.T) {
	// Direct demonstration of map race
	done := make(chan bool)

	go func() {
		for i := 0; i < 1000; i++ {
			sharedMap["race"] = i
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sharedMap["race"]
		}
		done <- true
	}()

	<-done
	<-done
}
