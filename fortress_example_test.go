package fortress_test

import (
	"fmt"

	"github.com/mrz1836/go-fortress"
)

// ExampleFortify demonstrates the usage of the Fortify function.
func ExampleFortify() {
	msg := fortress.Fortify("secret message")
	fmt.Println(msg)
	// Output: 🏰 secret message 🏰
}

// ExampleFortify_empty demonstrates fortifying an empty string.
func ExampleFortify_empty() {
	msg := fortress.Fortify("")
	fmt.Println(msg)
	// Output: 🏰  🏰
}

// ExampleGuard demonstrates the usage of the Guard function with safe input.
func ExampleGuard() {
	result, err := fortress.Guard("hello world", []string{"evil", "bad"})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(result)
	// Output: hello world
}

// ExampleGuard_breach demonstrates the Guard function detecting forbidden content.
func ExampleGuard_breach() {
	result, err := fortress.Guard("this is evil", []string{"evil", "bad"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
	// Output: 🚫 breach detected: contains "evil"
}

// ExampleGuard_emptyForbidden demonstrates Guard with an empty forbidden list.
func ExampleGuard_emptyForbidden() {
	result, err := fortress.Guard("anything is allowed", []string{})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(result)
	// Output: anything is allowed
}
