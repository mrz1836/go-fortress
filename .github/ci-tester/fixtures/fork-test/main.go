// Package main demonstrates fork detection for CI testing.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Fork test fixture")

	// Check if running from a fork
	isFork := os.Getenv("GITHUB_EVENT_NAME") == "pull_request" &&
		os.Getenv("GITHUB_HEAD_REF") != "" &&
		os.Getenv("GITHUB_BASE_REF") != ""

	if isFork {
		fmt.Println("is_fork: true")
	} else {
		fmt.Println("is_fork: false")
	}
}
