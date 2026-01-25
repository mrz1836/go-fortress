// Package main demonstrates data race conditions for CI testing.
package main

import (
	"fmt"
	"sync"
)

// Counter is a simple counter with a data race.
type Counter struct {
	value int
}

// sharedMap is a global map that will have concurrent access issues.
var sharedMap = make(map[string]int)

func main() {
	fmt.Println("Race condition fixture")
}

// Increment increments the counter (not thread-safe).
func (c *Counter) Increment() {
	c.value++
}

// Value returns the current counter value.
func (c *Counter) Value() int {
	return c.value
}

// UnsafeMapWrite writes to a shared map without synchronization.
func UnsafeMapWrite(key string, value int) {
	sharedMap[key] = value
}

// UnsafeMapRead reads from a shared map without synchronization.
func UnsafeMapRead(key string) int {
	return sharedMap[key]
}

// ConcurrentCounterIncrement runs multiple goroutines incrementing same counter.
func ConcurrentCounterIncrement(c *Counter, iterations int) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				c.Increment()
			}
		}()
	}
	wg.Wait()
}

// ConcurrentMapAccess performs concurrent map read/writes.
func ConcurrentMapAccess() {
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				UnsafeMapWrite(fmt.Sprintf("key%d", j), id*100+j)
			}
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = UnsafeMapRead(fmt.Sprintf("key%d", j))
			}
		}()
	}

	wg.Wait()
}
