module fixture-vuln-fail

go 1.24

// This module intentionally depends on a package with known vulnerabilities
// for testing vulnerability detection tools.
// golang.org/x/text v0.3.0 has known CVEs
require golang.org/x/text v0.3.0
