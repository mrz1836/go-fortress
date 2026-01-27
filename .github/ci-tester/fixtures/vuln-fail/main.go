// Package main demonstrates vulnerable dependency for CI testing.
package main

import (
	"fmt"

	"golang.org/x/text/language"
)

func main() {
	fmt.Println("Vulnerable dependency fixture")

	// Use the vulnerable package to ensure it's not pruned
	tag, _ := language.Parse("en-US")
	fmt.Printf("Language tag: %s\n", tag)
}

// GetLanguageTag returns a parsed language tag using the vulnerable package.
func GetLanguageTag(s string) (language.Tag, error) {
	return language.Parse(s)
}
