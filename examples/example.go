// Package main is an example of how to use the go-fortress package
package main

import (
	"log"

	"github.com/mrz1836/go-fortress"
)

func main() {
	// Greet the user with a custom name
	name := "Alice"
	greeting := fortress.Greet(name)
	log.Println(greeting)
}
