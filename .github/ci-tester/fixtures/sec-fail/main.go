// Package main demonstrates security issues for CI testing.
// WARNING: This file contains FAKE credentials for testing secret detection.
// These are NOT real credentials and are intentionally included for testing.
package main

import "fmt"

// Fake AWS credentials for testing gitleaks/secret detection.
// These follow the AWS key pattern but are not valid credentials.
const (
	// Fake AWS Access Key ID (follows AKIA pattern)
	// gitleaks:allow - This is a test fixture
	fakeAWSAccessKeyID = "AKIAIOSFODNN7EXAMPLE"

	// Fake AWS Secret Access Key
	// gitleaks:allow - This is a test fixture
	fakeAWSSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
)

// FakePrivateKey is a fake RSA private key for testing.
// This intentionally does NOT have gitleaks:allow to trigger detection
const FakePrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIBogIBAAJBALRiMLAHudeSA2ai2E9c9A/LPWx/tSTfefMLXLSB9xPokEXAMPLE
wBXN5K+9VBU+fUKqFiGmZlSfvk0EXAMPLEsEXAMPLECAwEAAQJAYEXAMPLE
-----END RSA PRIVATE KEY-----`

func main() {
	fmt.Println("Security fail fixture")
	// Simulated credential usage (for testing only)
	fmt.Printf("Using fake key: %s...\n", fakeAWSAccessKeyID[:10])
}

// GetFakeCredentials returns fake credentials for testing.
func GetFakeCredentials() (string, string) {
	return fakeAWSAccessKeyID, fakeAWSSecretKey
}

// HardcodedSecret demonstrates a hardcoded secret pattern.
func HardcodedSecret() string {
	// Another pattern that should be detected
	apiKey := "sk-proj-EXAMPLE1234567890abcdefghijklmnop"
	return apiKey
}
