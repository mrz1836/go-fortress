package fortress_test

import (
	"fmt"

	"github.com/mrz1836/go-fortress"
)

// ExampleGreet demonstrates the usage of the Greet function.
func ExampleGreet() {
	msg := fortress.Greet("Alice")
	fmt.Println(msg)
	// Output: Hello Alice
}
