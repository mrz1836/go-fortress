// Package main is an example of how to use the go-fortress package
package main

import (
	"log"

	"github.com/mrz1836/go-fortress"
)

func main() {
	// Fortify a message with fortress markers
	message := "secret data"
	fortified := fortress.Fortify(message)
	log.Println(fortified)

	// Guard input against forbidden values
	forbidden := []string{"evil", "bad"}
	result, err := fortress.Guard("hello world", forbidden)
	if err != nil {
		log.Println("Breach detected:", err)
		return
	}
	log.Println("Safe input:", result)
}
