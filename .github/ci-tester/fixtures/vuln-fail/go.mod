module fixture-vuln-fail

go 1.24

// This module intentionally depends on a package with known vulnerabilities
// for testing vulnerability detection tools.
require (
	// golang.org/x/text v0.3.0 has known CVEs
	golang.org/x/text v0.3.0
)
